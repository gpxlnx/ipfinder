package banner

import (
	"fmt"
)

// prints the version message
const version = "v0.0.1"

func PrintVersion() {
	fmt.Printf("Current ipfinder version %s\n", version)
}

// Prints the Colorful banner
func PrintBanner() {
	banner := `
    _         ____ _             __           
   (_)____   / __/(_)____   ____/ /___   _____
  / // __ \ / /_ / // __ \ / __  // _ \ / ___/
 / // /_/ // __// // / / // /_/ //  __// /    
/_// .___//_/  /_//_/ /_/ \__,_/ \___//_/     
  /_/`
	fmt.Printf("%s\n%50s\n\n", banner, "Current ipfinder version "+version)
}
