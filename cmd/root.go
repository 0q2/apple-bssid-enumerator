/*
Copyright © 2022 Erik Rye
*/
package cmd

import (
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apple-bssid-enumerator",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

// Initializes the logger to whatever level we're running at
func initLogging() {
	if verbose {
		log.SetLevel(log.DebugLevel)
		log.Debug("Logging at DebugLevel")
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	cobra.OnInitialize(initLogging)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().BoolVarP(&constants.SingleResponse, "single", "s", false, "Query WLOC for only single BSSID (no nearby)")
	rootCmd.PersistentFlags().IntVarP(&constants.NBSSIDs, "nBSSIDs", "N", 100, "Number of BSSIDs to include in each request to WLOC")
	rootCmd.PersistentFlags().IntVarP(&constants.NWorkers, "nWorkers", "w", 10, "Number of worker threads")
	rootCmd.PersistentFlags().StringVarP(&constants.OUIFile, "OUIfile", "f", "", "File of OUIs to query for")
	rootCmd.PersistentFlags().UintVarP(&constants.NPerOUI, "nPerOUI", "n", 16, "Exponent of number of BSSIDs to search for each OUI (2^x)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}
