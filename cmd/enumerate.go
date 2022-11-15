/*
Copyright © 2022 Erik Rye <erik@sixint.io>
*/
package cmd

import "C"
import (
	"github.com/gigaryte/apple-bssid-enumerator/common"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// enumerateCmd represents the enumerate command
var enumerateCmd = &cobra.Command{
	Use:   "enumerate",
	Short: "Scans a fixed number (specified by the user) of addresses in each OUI and quits",
	Run: func(cmd *cobra.Command, args []string) {
		constants.Enumerate = true
		common.ReadOUIFile()
		log.Debugln("Read", len(constants.OUIs), "OUIs")
		common.InitOUIInfo()
		common.RunQueries()

	},
}

func init() {
	rootCmd.AddCommand(enumerateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// enumerateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// enumerateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
