package common

import "C"
import (
	"bufio"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	"github.com/gigaryte/apple-bssid-enumerator/cperm"
	pb "github.com/gigaryte/apple-bssid-enumerator/proto"
	"github.com/gigaryte/apple-bssid-enumerator/wloc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	mu sync.Mutex
)

func ReadMACFile() {

	if constants.OUIFile == "" {
		log.Fatal("Error: -f/--infile is required")
	}

	f, err := os.Open(constants.OUIFile)
	if err != nil {
		log.Fatal("Error opening MAC file: ", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		constants.MACs = append(constants.MACs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func ReadOUIFile() {

	if constants.OUIFile == "" {
		log.Fatal("Error: -f/--infile is required")
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
	log.Debug("Initializing OUI info")
	cperm.OUIInfos = []*cperm.OUIInfo{}

	//Fill ouiChan for consumers to take from it
	for _, oui := range constants.OUIs {
		ouiChan <- oui
	}

	//Fire up nWorkers to initialize all the OUIInfos
	for i := 0; i < constants.NWorkers; i++ {
		wg.Add(1)
		//Launch goroutine to init the OUIInfo
		go func(ch chan string, i int) {
			defer wg.Done()
			for oui := range ouiChan {
				oi := &cperm.OUIInfo{
					OUI: oui,
				}
				oi.InitCPerm()
				mu.Lock()
				cperm.OUIInfos = append(cperm.OUIInfos, oi)
				mu.Unlock()
			}
		}(ouiChan, i)
	}

	//Wait here till all the worker threads finish initializing the OUIInfos
	close(ouiChan)
	wg.Wait()
	log.Infof("At the end of InitOUIInfo, len(OUIs): %v, len(OUIInfos): %v\n",
		len(constants.OUIs), len(cperm.OUIInfos))
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

func consume(jobs chan []byte, done chan bool) {

	wi := wloc.InitWloc()
	for {
		select {
		case j := <-jobs:
			wi.Message = j
			wi.Query()
		case <-done:
			log.Debugln("Received done signal from main thread; quitting")
			return
		}
	}
}

func RunMACQueries() {
	//used to indicate that we're done and the consumer should die
	done := make(chan bool)
	//Sends the query job as a byte slice to a consumer
	jobs := make(chan []byte)

	//Start NWorkers workers to do the querying
	for i := 0; i < constants.NWorkers; i++ {
		go consume(jobs, done)
	}

	var lookupMACs []string
	for len(constants.MACs) > constants.NBSSIDs {

		lookupMACs = constants.MACs[:constants.NBSSIDs]
		jobs <- serialize(lookupMACs)
		constants.MACs = constants.MACs[constants.NBSSIDs:]
	}

	//Do the remaining MACs
	if len(constants.MACs) > 0 {
		lookupMACs = constants.MACs
		jobs <- serialize(lookupMACs)
	}

	waitSeconds := 5
	log.Infof("Waiting %v seconds for the workers to finish\n", waitSeconds)
	time.Sleep(5 * time.Second)
	for i := 0; i < constants.NWorkers; i++ {
		done <- true
	}

}

func RunQueries() {

	//used to indicate that we're done and the consumer should die
	done := make(chan bool)
	//Sends the query job as a byte slice to a consumer
	jobs := make(chan []byte)

	//Start NWorkers workers to do the querying
	for i := 0; i < constants.NWorkers; i++ {
		go consume(jobs, done)
	}

	//Use this to know when we've exhausted all OUIs to query
	constants.Total = int64(len(cperm.OUIInfos))

	//Keeps the macs to run for the next query
	var lookupMACs []string
	for len(cperm.OUIInfos) > 0 {
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
				jobs <- serialize(lookupMACs)
				lookupMACs = []string{}
			}
		}
	}

	waitSeconds := 5
	log.Infof("Waiting %v seconds for the workers to finish\n", waitSeconds)
	time.Sleep(5 * time.Second)
	for i := 0; i < constants.NWorkers; i++ {
		done <- true
	}

}

func DetermineNextOUIs() {

	var nextRoundOUIs []string
	//This is the number of BSSIDs we queried per OUI * the threshold fraction that we have to meet to
	//progress to the next round
	nHitsToProgress := int(math.Pow(2, float64(constants.NPerOUI)) * constants.Threshold)

	for oui := range constants.BSSIDMap {
		nOUIHits := len(constants.BSSIDMap[oui])
		//There were enough hits for this one to progress to next level
		if nOUIHits >= nHitsToProgress {
			log.Debugf("%v: %v >= %v; progressing\n", oui, nOUIHits, nHitsToProgress)
			nextRoundOUIs = append(nextRoundOUIs, oui)
		} else {
			log.Debugf("%v: %v < %v; not progressing\n", oui, nOUIHits, nHitsToProgress)
		}
	}

	constants.OUIs = nextRoundOUIs
}
