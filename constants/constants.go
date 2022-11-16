package constants

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
)

const (
	KEYLEN  = 16
	PERMEND = -5
)
