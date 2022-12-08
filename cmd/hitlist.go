/*
Copyright © 2022 Erik Rye <erik@sixint.io>
*/
package cmd

import (
	"fmt"
	"github.com/gigaryte/apple-bssid-enumerator/common"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// hitlistCmd represents the hitlist command
var hitlistCmd = &cobra.Command{
	Use:   "hitlist",
	Short: "Look up individual MACs from a file",
	Long: `hitlist uses a list of individual MACs,
newline separated, to query the Apple location services
API for.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hitlist called")
		common.ReadMACFile()
		log.Debugln("Read", len(constants.MACs), "MACs")
		common.RunMACQueries()
	},
}

func init() {
	rootCmd.AddCommand(hitlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hitlistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hitlistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
