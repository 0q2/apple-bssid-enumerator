package common

import "C"
import (
	"bufio"
	"fmt"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	"github.com/gigaryte/apple-bssid-enumerator/cperm"
	pb "github.com/gigaryte/apple-bssid-enumerator/proto"
	"github.com/gigaryte/apple-bssid-enumerator/wloc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"sync"
)

func ReadOUIFile() {

	if constants.OUIFile == "" {
		log.Fatal("Error: -f/--OUIfile is required")
	}

	f, err := os.Open(constants.OUIFile)
	if err != nil {
		log.Fatal("Error opening OUI file: ", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		constants.OUIs = append(constants.OUIs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func InitOUIInfo() {

	var wg sync.WaitGroup
	ouiChan := make(chan string, len(constants.OUIs))

	//Fill ouiChan for consumers to take from it
	for _, oui := range constants.OUIs {
		ouiChan <- oui
	}

	//Fire up nWorkers to initialize all the OUIInfos
	for i := 0; i < constants.NWorkers; i++ {
		wg.Add(1)
		//Launch goroutine to init the OUIInfo
		go func(ch chan string) {
			for oui := range ouiChan {
				oi := &cperm.OUIInfo{
					OUI: oui,
				}
				oi.InitCPerm()
				cperm.OUIInfos = append(cperm.OUIInfos, oi)
			}
			wg.Done()
		}(ouiChan)
	}

	//Wait here till all the worker threads finish initializing the OUIInfos
	close(ouiChan)
	wg.Wait()
}

func serialize(macs []string) []byte {
	wifi := pb.WiFiLocation{}
	var zero int32 = 0
	var zero64 int64 = 0
	var one int32 = 1
	if constants.SingleResponse {
		wifi.Single = &one
	} else {
		wifi.Single = &zero
	}
	wifi.Unk1 = &zero64

	for _, mac := range macs {
		//Need this b/c variable shadowing
		localMAC := mac
		thisGeo := &pb.BSSIDGeo{}
		thisGeo.Bssid = &localMAC
		wifi.Wifi = append(wifi.Wifi, thisGeo)
	}

	out, err := proto.Marshal(&wifi)
	if err != nil {
		log.Fatal(err)
	}

	return out
}

func consume(jobs chan []byte) {

	wi := wloc.InitWloc()
	for j := range jobs {
		wi.Message = j
		wi.Query()
	}

}

func RunQueries() {

	jobs := make(chan []byte)

	//Start NWorkers workers to do the querying
	for i := 0; i < constants.NWorkers; i++ {
		go consume(jobs)
	}

	//Use this to know when we've exhausted all OUIs to query
	constants.Total = int64(len(cperm.OUIInfos))

	//Keeps the macs to run for the next query
	var lookupMACs []string
	for {
		ouiIdx := rand.Intn(len(cperm.OUIInfos))
		thisOUI := cperm.OUIInfos[ouiIdx]

		mac, ret := thisOUI.NextMAC()

		//Check if we're done; if ret == PERM_END, val is meaningless
		if ret == constants.PERMEND {
			log.Debugf("Got PERM_END for %v\n", thisOUI.OUI)
			thisOUI.CPermDestroy()
			//remove from the slice
			cperm.OUIInfos[ouiIdx] = cperm.OUIInfos[len(cperm.OUIInfos)-1]
			cperm.OUIInfos = cperm.OUIInfos[:len(cperm.OUIInfos)-1]
			constants.Finished++

			//If Finished == Total, this is the last network; all the others are done
			if constants.Finished == constants.Total {
				//Lookup the remaining ones
				if len(lookupMACs) > 0 {
					jobs <- serialize(lookupMACs)
				}
				log.Debugf("Finished final OUI; all %v complete", constants.Total)
				break
			}
		} else {
			lookupMACs = append(lookupMACs, mac)
			if len(lookupMACs) == constants.NBSSIDs {
				fmt.Println(lookupMACs)
				jobs <- serialize(lookupMACs)
				lookupMACs = []string{}
			}
		}
	}

}
