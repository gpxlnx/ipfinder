package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rix4uni/ipfinder/banner"
	"github.com/spf13/pflag"
)

var (
	facetFlag   string
	filterFlag  string
	delayFlag   int
	retriesFlag int
	sourceFlag  bool
	silent      bool
	version     bool
	verboseFlag bool
)

// makeRequestWithRetry performs HTTP GET request with retry logic for 429 errors
func makeRequestWithRetry(client *http.Client, url string, maxRetries int, verboseFlag bool) (*http.Response, error) {
	waitTime := 5 * time.Second
	attempt := 0

	for {
		resp, err := client.Get(url)
		if err != nil {
			return nil, err
		}

		if verboseFlag {
			fmt.Fprintf(os.Stderr, "Running: [%d] %s\n", resp.StatusCode, url)
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusBadRequest {
			resp.Body.Close()
			attempt++
			if maxRetries >= 0 && attempt > maxRetries {
				return nil, fmt.Errorf("max retries exceeded for 400/429 error")
			}
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "Running: [%ds] %s\n", int(waitTime.Seconds()), url)
			}
			time.Sleep(waitTime)
			waitTime *= 2 // exponential backoff: 5s, 10s, 20s, 40s...
			continue
		}

		return resp, nil
	}
}

func main() {
	// Define flags
	pflag.StringVarP(&facetFlag, "facet", "f", "ip", "Facet type (e.g., ip, domain, etc.)")
	pflag.StringVar(&filterFlag, "filter", "ssl", "Filter type (e.g., ssl, hostname, etc.)")
	pflag.IntVar(&delayFlag, "delay", 0, "Delay between city queries in seconds")
	pflag.IntVar(&retriesFlag, "retries", 4, "Maximum number of retries for 429 errors (-1 for unlimited)")
	pflag.BoolVar(&sourceFlag, "source", false, "Include the source query in the output")
	pflag.BoolVar(&silent, "silent", false, "Silent mode.")
	pflag.BoolVar(&version, "version", false, "Print the version of the tool and exit.")
	pflag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose output")

	pflag.Parse()

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Print version and exit if -version flag is provided
	if version {
		banner.PrintBanner()
		banner.PrintVersion()
		return
	}

	// Don't Print banner if -silent flag is provided
	if !silent {
		banner.PrintBanner()
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		input := scanner.Text() // read the input

		// Check if input already contains ':' (indicating it's already a full query format)
		// If not, construct query using the --filter flag
		var rawQuery string
		if strings.Contains(input, ":") {
			// Input is already in query format, use as-is (backward compatibility)
			rawQuery = input
		} else {
			// Construct query format: filterFlag:"input"
			rawQuery = fmt.Sprintf(`%s:"%s"`, filterFlag, input)
		}

		// Properly URL encode the query, then replace %20 with + for query strings
		query := strings.Replace(neturl.QueryEscape(rawQuery), "%20", "+", -1)

		// Step 2: Run the first query to get the total count
		url := fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag)
		resp, err := makeRequestWithRetry(client, url, retriesFlag, verboseFlag)
		if err != nil {
			fmt.Println("Error executing request:", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: HTTP status %d\n", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			continue
		}
		output := string(body)

		// Step 3: Extract the total count from the output using regex
		re := regexp.MustCompile(`Total: (\d{1,3}(?:,\d{3})*)`)
		match := re.FindStringSubmatch(string(output))
		if len(match) < 2 {
			if verboseFlag {
				fmt.Println("Failed to extract total count.")
			}
			continue
		}

		totalStr := match[1]
		total, err := strconv.Atoi(strings.Replace(totalStr, ",", "", -1))
		if err != nil {
			fmt.Println("Error converting total count:", err)
			continue
		}

		// Step 4: Logic based on total count
		if total < 1000 {
			// If total is less than 1000, run the command to print the SSL details
			url = fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag)
			resp, err := makeRequestWithRetry(client, url, retriesFlag, verboseFlag)
			if err != nil {
				fmt.Println("Error executing request:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: HTTP status %d\n", resp.StatusCode)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				continue
			}
			output = string(body)

			// Extract IP addresses from the response
			re = regexp.MustCompile(`<strong>(.*?)</strong>`)
			matches := re.FindAllStringSubmatch(string(output), -1)

			// Output the results
			for _, match := range matches {
				if sourceFlag {
					fmt.Printf("%s::%s\n", rawQuery, match[1])
				} else {
					fmt.Println(match[1])
				}
			}
		} else {
			// If total is greater than or equal to 1000, extract cities and perform the query for each city sequentially
			url = fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=city", query)
			resp, err := makeRequestWithRetry(client, url, retriesFlag, verboseFlag)
			if err != nil {
				fmt.Println("Error executing request:", err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: HTTP status %d\n", resp.StatusCode)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response:", err)
				continue
			}
			output = string(body)

			// Extract city names from the response
			re = regexp.MustCompile(`<strong>(.*?)</strong>`)
			cityMatches := re.FindAllStringSubmatch(string(output), -1)

			// Process cities sequentially to avoid Cloudflare captcha
			for _, city := range cityMatches {
				// Properly URL encode city name with quotes, then replace %20 with + for query strings
				cityWithQuotes := "\"" + city[1] + "\""
				cityEncoded := strings.Replace(neturl.QueryEscape(cityWithQuotes), "%20", "+", -1)
				cityURL := fmt.Sprintf("https://www.shodan.io/search/facet?query=%s+city:%s&facet=%s", query, cityEncoded, facetFlag)
				resp, err := makeRequestWithRetry(client, cityURL, retriesFlag, verboseFlag)
				if err != nil {
					fmt.Println("Error executing request for city", city[1], ":", err)
					continue
				}

				if resp.StatusCode != http.StatusOK {
					fmt.Printf("Error: HTTP status %d for city %s\n", resp.StatusCode, city[1])
					resp.Body.Close()
					continue
				}

				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					fmt.Println("Error reading response for city", city[1], ":", err)
					continue
				}
				output := string(body)

				// Extract IP addresses from the response for each city
				re := regexp.MustCompile(`<strong>(.*?)</strong>`)
				matches := re.FindAllStringSubmatch(string(output), -1)

				// Print results immediately
				for _, match := range matches {
					if sourceFlag {
						fmt.Fprintf(os.Stdout, "%s::%s\n", rawQuery, match[1])
					} else {
						fmt.Fprintf(os.Stdout, "%s\n", match[1])
					}
				}

				// Delay between city queries to avoid rate limiting
				if delayFlag > 0 {
					time.Sleep(time.Duration(delayFlag) * time.Second)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
