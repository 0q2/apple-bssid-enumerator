package constants

import (
	"os"
)

var (
	NPerOUI        uint
	Threshold      float64
	OUIFile        string
	Outfile        string
	OUIs           []string
	MACs           []string
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
