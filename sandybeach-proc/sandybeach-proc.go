package main

/*
* 海灘水質
* 暫無open data
* https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnSandyBeach.ashx?code=F36490AE812D4E4791C14975551DD2E9
* 將json轉成json/geojson
* 依照測點名稱只保留最新一筆資料

```json
[
  {
    "ID": 1,
    "ESCHERICHIA_COLI": "26",
    "ENTEROCOCCUS": "小於10",
    "WATER_LEVEL": "優良",
    "SAMPLE_DATE": "2019-07-23T00:00:00",
    "STATION_NAME": "福隆海水浴場",
    "LON": 121.94455,
    "LAT": 25.0223
  },
  {
    "ID": 2,
    "ESCHERICHIA_COLI": "19",
    "ENTEROCOCCUS": "小於10",
    "WATER_LEVEL": "優良",
    "SAMPLE_DATE": "2019-07-23T00:00:00",
    "STATION_NAME": "新金山海水浴場",
    "LON": 121.64436,
    "LAT": 25.23044
  }
]
```

```json
{
	"type": "FeatureCollection",
	"features": [
		{
			"type": "Feature",
			"geometry": {
				"type": "Point",
				"coordinates": [
					121.64436,
					25.23044
				]
			},
			"properties": {
				"ID": 32,
				"ESCHERICHIA_COLI": "15",
				"ENTEROCOCCUS": "小於10",
				"WATER_LEVEL": "優良",
				"SAMPLE_DATE": "2021-07-16T00:00:00+08:00",
				"STATION_NAME": "新金山海水浴場",
				"lon": 121.64436,
				"lat": 25.23044
			}
		}
	]
}
```

*/

import (
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	inFile  = flag.String("i", "sandybeach_raw.json", "input json file")
	outFile = flag.String("o", "sandybeach.json", "path to save output file")

	outJson = flag.Bool("json", false, "output json format")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	url = flag.String("u", "https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnSandyBeach.ashx?code=F36490AE812D4E4791C14975551DD2E9", "url")
	UA  = flag.String("ua", "OAC bot", "User-Agent")

	verbosity = flag.Int("v", 3, "verbosity for app")
)

func main() {
	flag.Parse()

	if *inFile != "" {
		fd, err := os.OpenFile(*inFile, os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", err)
			return
		}
		defer fd.Close()

		transFd(fd, *outFile)
		return
	}

	dialFunc := func(network, address string) (net.Conn, error) {
		return net.DialTimeout("tcp", address, time.Duration(*connTimeout)*time.Second)
	}
	if *proxyAddr != "" {
		dialFunc = func(network, address string) (net.Conn, error) {
			if network != "tcp" {
				return nil, errors.New("only support tcp")
			}
			return makeConnection(address, *proxyAddr, time.Duration(*connTimeout)*time.Second)
		}
	}

	fd, err := getUrlFd(*url, dialFunc)
	if err != nil {
		Vln(2, "[get]err", *url, err)
		return
	}
	defer fd.Close()

	Vln(3, "[get]start download...", *url)

	err = transFd(fd, *outFile)
	if err != nil {
		Vln(2, "[json]err", err)
	}
	Vln(3, "[json]ok")
}

func getUrlFd(url string, dialFunc func(network, addr string) (net.Conn, error)) (io.ReadCloser, error) {
	var netTransport = &http.Transport{
		Dial:                dialFunc,
		TLSHandshakeTimeout: time.Duration(*connTimeout) * time.Second,
	}

	var netClient = &http.Client{
		Timeout:   time.Second * 180,
		Transport: netTransport,
	}

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", *UA)
	req.Close = true
	res, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// csv stream to one json
func transFd(fd io.Reader, outFp string) error {

	fdOut, err := os.OpenFile(outFp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Vln(2, "[open]err", outFp, err)
		return err
	}
	defer fdOut.Close()

	rows, err := parseData(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return err
	}
	Vln(3, "[data]out", len(rows))

	enc := json.NewEncoder(fdOut)
	enc.SetIndent("", "\t")

	if *outJson {
		err = enc.Encode(rows)
	} else {
		err = enc.Encode(NewGeojson(rows))
	}
	if err != nil {
		Vln(2, "[json]err", err)
		return err
	}
	return nil
}

type Geojson struct {
	Type     string     `json:"type"` // FeatureCollection
	Features []*Feature `json:"features"`
}

func NewGeojson(rows []*Row) *Geojson {
	features := make([]*Feature, len(rows), cap(rows))
	for i, row := range rows {
		features[i] = (*Feature)(row)
	}
	return &Geojson{
		Type:     "FeatureCollection",
		Features: features,
	}
}

// ==== proc Data ====
type Row struct {
	ID               int     `json:"ID"`               // useless
	ESCHERICHIA_COLI string  `json:"ESCHERICHIA_COLI"` // 大腸桿菌群
	ENTEROCOCCUS     string  `json:"ENTEROCOCCUS"`     // 腸球菌群
	WATER_LEVEL      string  `json:"WATER_LEVEL"`      // 水質分級
	SAMPLE_DATE      Time8   `json:"SAMPLE_DATE"`      // 採樣日期
	STATION_NAME     string  `json:"STATION_NAME"`
	LON              float64 `json:"lon"` // 經度
	LAT              float64 `json:"lat"` // 緯度
}

type Feature Row

func (row *Feature) MarshalJSON() ([]byte, error) {
	type Geometry struct {
		Type   string    `json:"type"`        // Point
		Coords []float64 `json:"coordinates"` // [lon, lat]
	}
	aux := &struct {
		Type  string    `json:"type"` // Feature
		Geo   *Geometry `json:"geometry,omitempty"`
		Props *Row      `json:"properties,omitempty"`
	}{
		Type: "Feature",
		Geo: &Geometry{
			Type:   "Point",
			Coords: []float64{row.LON, row.LAT},
		},
		Props: (*Row)(row),
	}
	return json.Marshal(aux)
}

type sortByTime []*Row

func (s sortByTime) Len() int      { return len(s) }
func (s sortByTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTime) Less(i, j int) bool {
	return time.Time(s[i].SAMPLE_DATE).Before(time.Time(s[j].SAMPLE_DATE))
}

var TZ8 = time.FixedZone("UTC+8 Time", int((8 * time.Hour).Seconds()))

type Time8 time.Time

func (v *Time8) UnmarshalJSON(in []byte) error {
	tIn := strings.Trim(string(in), `"`)
	t, err := time.ParseInLocation("2006-01-02T15:04:05", tIn, TZ8)
	// t, err := time.ParseInLocation("2006-01-02T15:04:05Z07:00", tIn+"+08:00", TZ8)
	// t, err := time.ParseInLocation(time.RFC3339, tIn+"+08:00", TZ8)
	if err != nil {
		return err
	}
	*v = Time8(t)
	return nil
}
func (v *Time8) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*v))
}

func parseData(r io.Reader) ([]*Row, error) {
	rows := make([]*Row, 0, 128)
	dec := json.NewDecoder(r)
	err := dec.Decode(&rows)
	if err != nil && err != io.EOF { // skip EOF
		return sortAndFilter(rows), err
	}
	return sortAndFilter(rows), nil
}

func sortAndFilter(rows []*Row) []*Row {
	Vln(3, "[data]raw", len(rows))
	sort.Stable(sort.Reverse(sortByTime(rows)))

	exist := make(map[string]bool, len(rows))
	out := make([]*Row, 0, len(rows))
	for _, row := range rows {
		ok := exist[row.STATION_NAME]
		if ok {
			continue
		}
		exist[row.STATION_NAME] = true
		out = append(out, row)
	}
	return out
}

// ==== proxy ====
func makeConnection(targetAddr string, socksAddr string, timeout time.Duration) (net.Conn, error) {

	host, portStr, err := net.SplitHostPort(targetAddr)
	if err != nil {
		Vln(2, "SplitHostPort err:", targetAddr, err)
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		Vln(2, "failed to parse port number:", portStr, err)
		return nil, err
	}
	if port < 1 || port > 0xffff {
		Vln(2, "port number out of range:", portStr, err)
		return nil, err
	}

	socksReq := []byte{0x05, 0x01, 0x00, 0x03}
	socksReq = append(socksReq, byte(len(host)))
	socksReq = append(socksReq, host...)
	socksReq = append(socksReq, byte(port>>8), byte(port))

	conn, err := net.DialTimeout("tcp", socksAddr, timeout)
	if err != nil {
		Vln(2, "connect to ", socksAddr, err)
		return nil, err
	}

	var b [10]byte

	// send request
	conn.Write([]byte{0x05, 0x01, 0x00})

	// read reply
	_, err = conn.Read(b[:2])
	if err != nil {
		return nil, err
	}

	// send server addr
	conn.Write(socksReq)

	// read reply
	n, err := conn.Read(b[:10])
	if n < 10 {
		Vln(2, "Dial err replay:", targetAddr, "via", socksAddr, n)
		return nil, err
	}
	if err != nil || b[1] != 0x00 {
		Vln(2, "Dial err:", targetAddr, "via", socksAddr, n, b[1], err)
		return nil, err
	}

	return conn, nil
}

// ==== log ====
func Vf(level int, format string, v ...interface{}) {
	if level <= *verbosity {
		log.Printf(format, v...)
	}
}
func Vln(level int, v ...interface{}) {
	if level <= *verbosity {
		log.Println(v...)
	}
}
