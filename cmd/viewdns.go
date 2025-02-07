package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func fetchIPHistory(domain string) {
	url := fmt.Sprintf("https://viewdns.info/iphistory/?domain=%s", domain)
	client := &http.Client{}

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request for %s: %v\n", domain, err)
		return
	}

	// Set the User-Agent header
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error performing request for %s: %v\n", domain, err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body for %s: %v\n", domain, err)
		return
	}

	// Define the regex to match IP addresses
	re := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)

	// Find all matches
	ips := re.FindAllString(string(body), -1)

	// Print the extracted IP addresses
	for _, ip := range ips {
		fmt.Println(ip)
	}
}

// viewdnsCmd represents the viewdns command
var viewdnsCmd = &cobra.Command{
	Use:   "viewdns",
	Short: "Fetch IP history for a domain (Website not supports subdomains)",
	Long: `Fetch the IP history for one or multiple domains using viewdns.info.

Examples:
 echo "sqrx.com" | ipfinder viewdns
 echo "public.sqrx.com" | ipfinder viewdns (Wrong input)
 cat subs.txt | ipfinder viewdns
 `,
	Run: func(cmd *cobra.Command, args []string) {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain == "" {
				continue
			}
			fetchIPHistory(domain)
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(viewdnsCmd)
}
