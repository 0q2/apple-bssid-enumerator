package constants

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var (
	NPerOUI        uint
	Threshold      float64
	OUIFile        string
	Outfile        string
	OUIs           []string
	NWorkers       int
	Total          int64
	Finished       int64
	NBSSIDs        int
	SingleResponse bool
	Enumerate      bool
	Iterate        bool
	NIterations    int
	BSSIDMap       = make(map[string]map[string]bool)
	OutfilePtr     *os.File
)

const (
	KEYLEN  = 16
	PERMEND = -5
)

func init() {
	var err error
	if Outfile != "" {
		OutfilePtr, err = os.OpenFile(Outfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
}
