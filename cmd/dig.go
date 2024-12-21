package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// digCmd represents the dig command
var digCmd = &cobra.Command{
	Use:   "dig",
	Short: "Run the dig command to get DNS A records",
	Long:  `This command uses the 'dig' utility to query DNS A records for a given domain.

Examples:
 echo "sqrx.com" | go run main.go dig
 cat subs.txt | go run main.go dig`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create a scanner to read from stdin
		scanner := bufio.NewScanner(os.Stdin)

		// Read domains from stdin line by line
		for scanner.Scan() {
			domain := strings.TrimSpace(scanner.Text())
			if domain != "" {
				runDig(domain)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
	},
}

// runDig executes the dig command and prints the output
func runDig(domain string) {
	cmd := exec.Command("dig", "@1.1.1.1", domain, "A", "+short")

	// Capture the output
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running dig for %s: %v\n", domain, err)
		return
	}

	// Print the output directly, trimmed of any extra whitespace
	result := strings.TrimSpace(string(output))
	if result != "" {
		fmt.Println(result)
	}
}

func init() {
	rootCmd.AddCommand(digCmd)
}
