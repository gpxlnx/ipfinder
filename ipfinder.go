package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rix4uni/ipfinder/banner"
	"github.com/spf13/pflag"
)

var (
	facetFlag      string
	concurrentFlag int
	delayFlag      int
	sourceFlag     bool
	silent         bool
	version        bool
	verboseFlag    bool
)

func main() {
	// Define flags
	pflag.StringVarP(&facetFlag, "facet", "f", "ip", "Facet type (e.g., ip, domain, etc.)")
	pflag.IntVarP(&concurrentFlag, "concurrent", "c", 1, "Number of concurrent city queries")
	pflag.IntVarP(&delayFlag, "delay", "d", 0, "Delay between batches of city queries in milliseconds")
	pflag.BoolVar(&sourceFlag, "source", false, "Include the source query in the output")
	pflag.BoolVar(&silent, "silent", false, "Silent mode.")
	pflag.BoolVar(&version, "version", false, "Print the version of the tool and exit.")
	pflag.BoolVar(&verboseFlag, "verbose", false, "Enable verbose output")

	pflag.Parse()

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
		rawQuery := scanner.Text()                       // keep the original
		query := strings.Replace(rawQuery, " ", "+", -1) // used for URL encoding

		// Step 2: Run the first query to get the total count
		url := fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag)
		if verboseFlag {
			fmt.Fprintf(os.Stderr, "Running: curl -s %s\n", url)
		}
		cmd := exec.Command("curl", "-s", url)
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Error executing command:", err)
			return
		}

		// Step 3: Extract the total count from the output using regex
		re := regexp.MustCompile(`Total: (\d{1,3}(?:,\d{3})*)`)
		match := re.FindStringSubmatch(string(output))
		if len(match) < 2 {
			if verboseFlag {
				fmt.Println("Failed to extract total count.")
			}
			return
		}

		totalStr := match[1]
		total, err := strconv.Atoi(strings.Replace(totalStr, ",", "", -1))
		if err != nil {
			fmt.Println("Error converting total count:", err)
			return
		}

		// Step 4: Logic based on total count
		if total < 1000 {
			// If total is less than 1000, run the command to print the SSL details
			url = fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag)
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "Running: curl -s %s\n", url)
			}
			cmd = exec.Command("curl", "-s", url)
			output, err = cmd.Output()
			if err != nil {
				fmt.Println("Error executing command:", err)
				return
			}

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
			// If total is greater than or equal to 1000, extract cities and perform the query for each city concurrently
			url = fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=city", query)
			if verboseFlag {
				fmt.Fprintf(os.Stderr, "Running: curl -s %s\n", url)
			}
			cmd = exec.Command("curl", "-s", url)
			output, err = cmd.Output()
			if err != nil {
				fmt.Println("Error executing command:", err)
				return
			}

			// Extract city names from the response
			re = regexp.MustCompile(`<strong>(.*?)</strong>`)
			cityMatches := re.FindAllStringSubmatch(string(output), -1)

			var mutex sync.Mutex
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, concurrentFlag)

			for i, city := range cityMatches {
				wg.Add(1)
				semaphore <- struct{}{} // Acquire a slot in the semaphore

				go func(city string) {
					defer wg.Done()
					defer func() { <-semaphore }() // Release the slot in the semaphore

					cityQuery := strings.Replace(city, " ", "+", -1)
					cityURL := fmt.Sprintf("https://www.shodan.io/search/facet?query=%s+city:\"%s\"&facet=%s", query, cityQuery, facetFlag)
					if verboseFlag {
						fmt.Fprintf(os.Stderr, "Running: curl -s %s\n", cityURL)
					}
					cmd := exec.Command("curl", "-s", cityURL)
					output, err := cmd.Output()
					if err != nil {
						fmt.Println("Error executing command for city", city, ":", err)
						return
					}

					// Extract IP addresses from the response for each city
					re := regexp.MustCompile(`<strong>(.*?)</strong>`)
					matches := re.FindAllStringSubmatch(string(output), -1)

					// Print results immediately
					mutex.Lock()
					for _, match := range matches {
						if sourceFlag {
							fmt.Fprintf(os.Stdout, "%s::%s\n", rawQuery, match[1])
						} else {
							fmt.Fprintf(os.Stdout, "%s\n", match[1])
						}
					}
					mutex.Unlock()
				}(city[1])

				// Introduce delay between batches
				if (i+1)%concurrentFlag == 0 {
					time.Sleep(time.Duration(delayFlag) * time.Millisecond)
				}
			}

			// Wait for all goroutines to complete
			wg.Wait()
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
