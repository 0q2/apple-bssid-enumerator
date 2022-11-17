/*
Copyright © 2022 Erik Rye <erik@sixint.io>
*/
package cmd

import (
	"github.com/gigaryte/apple-bssid-enumerator/common"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	"github.com/gigaryte/apple-bssid-enumerator/cperm"
	log "github.com/sirupsen/logrus"
	"math"

	"github.com/spf13/cobra"
)

// iterateCmd represents the iterate command
var iterateCmd = &cobra.Command{
	Use:   "iterate",
	Short: "Iteratively discover allocated IEEE WiFi OUI space",
	Long: `iterate repeatedly iterates over a (potentially large) number
of IEEE OUIs, searching for portions of the space used for WiFi access points.`,

	Run: func(cmd *cobra.Command, args []string) {
		constants.Iterate = true
		common.ReadOUIFile()
		log.Debugln("Read", len(constants.OUIs), "OUIs")

		//Do NIteration number of iterations
		for i := 0; i < constants.NIterations; i++ {

			log.Infof("There are %v OUIs to probe\n", len(constants.OUIs))

			common.InitOUIInfo()

			log.Infof("len(cperm.OUIInfos) at check: %v\n", len(cperm.OUIInfos))
			if len(cperm.OUIInfos) == 0 {
				log.Infof("No OUIs remain to be probed; exiting\n")
				break
			}

			log.Debugf("Running queries with %v BSSIDs per OUI", math.Pow(2, float64(constants.NPerOUI)))
			common.RunQueries()
			common.DetermineNextOUIs()
			//This is a 2^x exponent, so ++ doubles it
			constants.NPerOUI++
		}

	},
}

func calculateIterations() {
	//If this is 0, need to figure out how many iterations we actually need to do
	if constants.NIterations == 0 {
		//24 is the maximum, since there are 2^24 MACs in an OUI
		constants.NIterations = int(24 - constants.NPerOUI)
		log.Infof("User selected -i/--iterations = 0; max iterations is actually %v\n", constants.NIterations)
	}
}

func init() {
	rootCmd.AddCommand(iterateCmd)
	cobra.OnInitialize(calculateIterations)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	iterateCmd.PersistentFlags().Float64VarP(&constants.Threshold, "threshold", "t", 0.10, "Threshold for continuing to query database for this OUI in next round")
	iterateCmd.PersistentFlags().IntVarP(&constants.NIterations, "iterations", "i", 0, "Max number of doubling iterations (select 0 to continue until 2^24 or all OUI are exhauasted)")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// iterateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
