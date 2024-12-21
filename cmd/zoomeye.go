package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type Response struct {
	Matches []Match `json:"matches"`
}

type Match struct {
	IP []string `json:"ip"`
}

// zoomeyeCmd represents the zoomeye command
var zoomeyeCmd = &cobra.Command{
	Use:   "zoomeye",
	Short: "Fetch IP history for a domain (Website not supports subdomains)",
	Long: `Fetch the IP history for one or multiple domains using zoomeye.hk.

Examples:
 echo "sqrx.com" | go run main.go zoomeye
 cat subs.txt | go run main.go zoomeye`,
	Run: func(cmd *cobra.Command, args []string) {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain == "" {
				continue
			}

			url := fmt.Sprintf("https://www.zoomeye.hk/api/search?q=site:%s&page=1&t=v4+v6+web", domain)

			// Create a new request
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				log.Printf("Failed to create request for domain %s: %v", domain, err)
				continue
			}
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

			// Make the HTTP request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Failed to make request for domain %s: %v", domain, err)
				continue
			}
			defer resp.Body.Close()

			// Read the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Failed to read response for domain %s: %v", domain, err)
				continue
			}

			// Parse the JSON response
			var result Response
			if err := json.Unmarshal(body, &result); err != nil {
				log.Printf("Failed to parse JSON for domain %s: %v", domain, err)
				continue
			}

			// Extract and print IP addresses
			for _, match := range result.Matches {
				for _, ip := range match.IP {
					fmt.Println(ip)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(zoomeyeCmd)
}
