/*
Copyright © 2022 Erik Rye <erik@sixint.io>
*/
package cmd

import (
	"fmt"
	"github.com/gigaryte/apple-bssid-enumerator/constants"

	"github.com/spf13/cobra"
)

// iterateCmd represents the iterate command
var iterateCmd = &cobra.Command{
	Use:   "iterate",
	Short: "Iteratively discover allocated IEEE WiFi OUI space",
	Long: `iterate repeatedly iterates over a (potentially large) number
of IEEE OUIs, searching for portions of the space used for WiFi access points.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("iterate called")
	},
}

func init() {
	rootCmd.AddCommand(iterateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	iterateCmd.PersistentFlags().Float64VarP(&constants.Threshold, "threshold", "t", 0.10, "Threshold for continuing to query database for this OUI in next round")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// iterateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
