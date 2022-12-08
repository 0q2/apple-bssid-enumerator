package wloc

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/gigaryte/apple-bssid-enumerator/constants"
	pb "github.com/gigaryte/apple-bssid-enumerator/proto"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"io"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	URL  = "https://gs-loc.apple.com/clls/wloc"
	HOST = "gs-loc.apple.com"
)

var (
	mu      sync.Mutex
	NBSSIDs int
)

type Wloc struct {
	Header     []byte
	Locale     string
	Identifier string
	Version    string
	Footer     []byte
	Message    []byte
	WlocHeader []byte
}

func InitWloc() *Wloc {
	wl := &Wloc{}
	wl.Header = []byte{0x00, 0x01}
	wl.Locale = "en_US"
	wl.Identifier = "com.apple.locationd"
	wl.Version = "8.4.1.12H321"
	wl.Footer = []byte{0x00, 0x00}

	wl.SerializeHeader()
	return wl
}

func (wl *Wloc) SerializeHeader() {

	var out []byte

	/* Header */
	out = append(out, wl.Header...)
	/* Locale string */
	out = append(out, []byte{0x00, byte(len([]rune(wl.Locale)))}...)
	out = append(out, []byte(wl.Locale)...)
	/* Identifier string */
	out = append(out, []byte{0x00, byte(len([]rune(wl.Identifier)))}...)
	out = append(out, []byte(wl.Identifier)...)
	/* Version string */
	out = append(out, []byte{0x00, byte(len([]rune(wl.Version)))}...)
	out = append(out, []byte(wl.Version)...)
	/* Footer */
	out = append(out, wl.Footer...)
	/* Second header/footer -- unclear why */
	out = append(out, wl.Header...)
	out = append(out, wl.Footer...)

	wl.WlocHeader = out
}

func (wl *Wloc) Query() {

	msgLen := uint16(len(wl.Message))
	lenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(lenBuf, msgLen)
	buf := append(wl.WlocHeader, lenBuf...)
	buf = append(buf, wl.Message...)

	body := bytes.NewBuffer(buf)
	req, err := http.NewRequest("POST", URL, body)
	if err != nil {
		log.Errorln("Error in creating the request; skipping:", err)
		return
	}
	req.Host = HOST

	req.Header = http.Header{
		"Content-Type":    {"application/x-www-form-urlencoded"},
		"Accept":          {"*/*"},
		"Accept-Charset":  {"utf-8"},
		"Accept-Encoding": {"gzip, deflate"},
		"Accept-Language": {"en-us"},
		"User-Agent":      {"locationd/1753.17 CFNetwork/711.1.12 Darwin/14.0.0"},
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}

	res, err := client.Do(req)
	if err != nil {
		//Something went wrong with the request; log it and return
		log.Errorln("Error doing the client request: ", err)
		return
	}
	defer res.Body.Close()

	var reader io.ReadCloser
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(res.Body)
		defer reader.Close()
	default:
		reader = res.Body
	}

	resp, err := io.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	//There's just a 10 byte header we're going to remove here
	resp = resp[10:]
	wifi := pb.WiFiLocation{}
	err = proto.Unmarshal(resp, &wifi)
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range wifi.Wifi {
		bssid := padBSSID(w.GetBssid())
		oui := bssid[:8]
		lat := float64(w.Location.GetLat()) * math.Pow10(-8)
		lon := float64(w.Location.GetLon()) * math.Pow10(-8)

		if constants.Iterate {
			mu.Lock()
			if _, ok := constants.BSSIDMap[oui]; !ok {
				constants.BSSIDMap[oui] = make(map[string]bool)
			}
			//Make sure it's a valid geo before adding it to the hit list
			if lat != -180 && lon != -180 {
				constants.BSSIDMap[oui][bssid] = true
			}
			mu.Unlock()
		}

		if constants.OutfilePtr != nil {
			constants.OutfilePtr.WriteString(fmt.Sprintf("%v %v %v,%v\n", time.Now().Unix(),
				bssid, lat, lon))
		} else {
			fmt.Printf("%v %v %v,%v\n", time.Now().Unix(),
				bssid, lat, lon)
		}
	}

}

func padBSSID(s string) string {
	var ret []string
	for _, e := range strings.Split(s, ":") {
		if len(e) < 2 {
			e = "0" + e
		}
		ret = append(ret, e)
	}

	return strings.Join(ret, ":")
}
