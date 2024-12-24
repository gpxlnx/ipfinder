package cmd

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

	"github.com/spf13/cobra"
)

var (
	facetFlag      string
	concurrentFlag int
	delayFlag      int
)

// shodanCmd represents the shodan command
var shodanCmd = &cobra.Command{
	Use:   "shodan",
	Short: "Search for SSL details on Shodan (Website supports subdomains but recommended to use domain)",
	Long: `Search Shodan for SSL details based on the input query.
It supports extracting results by facet and querying cities concurrently.

Examples:
 echo 'ssl:"sqrx.com"' | go run main.go shodan
 echo 'hostname:"sqrx.com"' | go run main.go shodan
 echo 'ssl.cert.subject.cn:"sqrx.com"' | go run main.go shodan
 echo 'org:"FIDELITY NATIONAL INFORMATION SERVICES"' | go run main.go shodan
 echo 'asn:"AS3614"' | go run main.go shodan
 cat subs.txt | go run main.go shodan

# Use this for more filters: https://www.shodan.io/search/filters

Note: After this you have to run naabu to get port against these ips because this method only gives ips
 cat ips.txt | naabu -duc -silent -passive
 `,
	Run: func(cmd *cobra.Command, args []string) {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			query := scanner.Text()
			query = strings.Replace(query, " ", "+", -1)

			// Step 2: Run the first query to get the total count
			cmd := exec.Command("curl", "-s", fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag))
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("Error executing command:", err)
				return
			}

			// Step 3: Extract the total count from the output using regex
			re := regexp.MustCompile(`Total: (\d{1,3}(?:,\d{3})*)`)
			match := re.FindStringSubmatch(string(output))
			if len(match) < 2 {
				fmt.Println("Failed to extract total count.")
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
				cmd = exec.Command("curl", "-s", fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=%s", query, facetFlag))
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
					fmt.Println(match[1])
				}
			} else {
				// If total is greater than or equal to 1000, extract cities and perform the query for each city concurrently
				cmd = exec.Command("curl", "-s", fmt.Sprintf("https://www.shodan.io/search/facet?query=%s&facet=city", query))
				output, err = cmd.Output()
				if err != nil {
					fmt.Println("Error executing command:", err)
					return
				}

				// Extract city names from the response
				re = regexp.MustCompile(`<strong>(.*?)</strong>`)
				cityMatches := re.FindAllStringSubmatch(string(output), -1)

				var extractedResults []string
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
						cmd := exec.Command("curl", "-s", fmt.Sprintf("https://www.shodan.io/search/facet?query=%s+city:\"%s\"&facet=%s", query, cityQuery, facetFlag))
						output, err := cmd.Output()
						if err != nil {
							fmt.Println("Error executing command for city", city, ":", err)
							return
						}

						// Extract IP addresses from the response for each city
						re := regexp.MustCompile(`<strong>(.*?)</strong>`)
						matches := re.FindAllStringSubmatch(string(output), -1)

						mutex.Lock()
						for _, match := range matches {
							extractedResults = append(extractedResults, match[1])
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

				// Output the results
				for _, result := range extractedResults {
					fmt.Println(result)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(shodanCmd)

	// Define the flags
	shodanCmd.Flags().StringVarP(&facetFlag, "facet", "f", "ip", "Facet type (e.g., ip, domain, etc.)")
	shodanCmd.Flags().IntVarP(&concurrentFlag, "concurrent", "c", 1, "Number of concurrent city queries")
	shodanCmd.Flags().IntVarP(&delayFlag, "delay", "d", 0, "Delay between batches of city queries in milliseconds")
}
