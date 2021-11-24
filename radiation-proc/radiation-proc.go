package main

/*
* 行政院原子能委員會輻射偵測中心 海域輻射 open data
* 台灣海域輻射調查資料
* https://data.gov.tw/dataset/136060
* 將csv轉成json/geojson
* 民國日期轉換為西元日期
* 經緯度 度分秒 轉為10進制小數


注意:
當前的經緯度跟欄位標示是相反的!!!
此程式有手動修正
若未來資料有修正
請自行移除該段code(關鍵字"workaround")

```csv
取樣日期,經度,緯度,深度(M),銫-137活度,銫-137單位,氚活度,氚單位,類別,地點編號,地點,取樣單位/人,備註
106/03/14,"22°20'18.03""","120°53'55.88""",,－,貝克/公斤-乾重,,,沉積物,R-9,大武漁港,偵測中心,岸沙
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
					120.89885555555556,
					22.338341666666665
				]
			},
			"properties": {
				"time": "2017-03-14T00:00:00+08:00",
				"lon": 120.89885555555556,
				"lat": 22.338341666666665,
				"cs137": "--",
				"cs137unit": "貝克/公斤-乾重",
				"category": "沉積物",
				"locationid": "R-9",
				"location": "大武漁港",
				"sampler": "偵測中心",
				"memo": "岸沙"
			}
		}
	]
}
```

*/

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	inFile  = flag.String("i", "radiation_raw.csv", "input CSV file")
	outFile = flag.String("o", "radiation.json", "path to save output file")

	keepTime = flag.Int("keep", 30*6, "keep latest N days data")

	outJson = flag.Bool("json", false, "output json format")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	url = flag.String("u", "https://www.aec.gov.tw/share/file/information/gdhyTsQBfnL9N3h7Xf2iYg__.csv", "url")
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

		transFd(fd, *outFile, *keepTime)
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

	err = transFd(fd, *outFile, *keepTime)
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
func transFd(fd io.Reader, outFp string, keep int) error {

	fdOut, err := os.OpenFile(outFp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Vln(2, "[open]err", outFp, err)
		return err
	}
	defer fdOut.Close()

	rows, err := parseCSV(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return err
	}
	Vln(3, "[data]", len(rows))

	if keep > 0 {
		t0 := time.Now().AddDate(0, 0, -keep)
		for i, row := range rows {
			if row.Time.Before(t0) {
				rows = rows[:i]
				break
			}
		}
		Vln(3, "[data]", len(rows))
	}

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

// ==== proc CSV ====
type Row struct {
	Time time.Time `json:"time"`        // 取樣日期
	Lon  float64   `json:"lon"`         // 經度
	Lat  float64   `json:"lat"`         // 緯度
	Z    string    `json:"Z,omitempty"` // 深度(M)

	Cs137     *jsonFloat `json:"cs137,omitempty"`     // 銫-137活度
	Cs137Unit string     `json:"cs137unit,omitempty"` // 銫-137單位

	H3     *jsonFloat `json:"h3,omitempty"`     // 氚活度
	H3Unit string     `json:"h3unit,omitempty"` // 氚單位

	Category   string `json:"category,omitempty"`   // 類別
	LocationID string `json:"locationid,omitempty"` // 地點編號
	Location   string `json:"location,omitempty"`   // 地點
	Sampler    string `json:"sampler,omitempty"`    // 取樣單位/人
	Memo       string `json:"memo"`                 // 備註
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
			Coords: []float64{row.Lon, row.Lat},
		},
		Props: (*Row)(row),
	}
	return json.Marshal(aux)
}

type sortByTime []*Row

func (s sortByTime) Len() int           { return len(s) }
func (s sortByTime) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortByTime) Less(i, j int) bool { return s[i].Time.Before(s[j].Time) }

type jsonFloat float64

func (value jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return []byte("\"--\""), nil
	}
	return []byte(fmt.Sprintf("%v", value)), nil
}
func (value *jsonFloat) Set(v float64) {
	*value = jsonFloat(v)
}

func NewJsonFloat(v float64) *jsonFloat {
	val := new(jsonFloat)
	*val = jsonFloat(v)
	return val
}

func parseCSV(r0 io.Reader) ([]*Row, error) {
	r := bufio.NewReader(r0)
	b, err := r.Peek(3)
	if err != nil {
		return nil, err
	}
	if b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		r.Discard(3)
	}

	rows := make([]*Row, 0, 128)
	//	return rows, nil
	cr := csv.NewReader(r)
	idx := 0
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return rows, err
		}
		Vln(7, "[row]", len(rec), rec)
		if len(rec) < 13 {
			return rows, errors.New("data column change!!!")
		}

		idx++
		if idx == 1 {
			continue
		}

		row := &Row{
			Z:          rec[3],  // 深度(M)
			Cs137Unit:  rec[5],  // 銫-137單位
			H3Unit:     rec[7],  // 氚單位
			Category:   rec[8],  // 類別
			LocationID: rec[9],  // 地點編號
			Location:   rec[10], // 地點
			Sampler:    rec[11], // 取樣單位/人
			Memo:       rec[12], // 備註
		}

		row.Time = parseTime(rec[0])
		row.Lon = parseLatLon(rec[1])
		row.Lat = parseLatLon(rec[2])

		// workaround: fix for bad column define / value
		row.Lon, row.Lat = row.Lat, row.Lon

		row.Cs137 = parseVal(rec[4])
		row.H3 = parseVal(rec[6])

		rows = append(rows, row)
	}
	sort.Stable(sort.Reverse(sortByTime(rows)))
	return rows, nil
}

var TZ8 = time.FixedZone("UTC+8 Time", int((8 * time.Hour).Seconds()))

func parseTime(v string) time.Time {
	tmp := strings.SplitN(v, "/", 3)
	if len(tmp) != 3 {
		return time.Time{}
	}
	yyyy, mm, dd := 1911, 00, 00
	if s, err := strconv.ParseUint(tmp[0], 10, 32); err == nil {
		yyyy += int(s)
	}
	if s, err := strconv.ParseUint(tmp[1], 10, 32); err == nil {
		mm += int(s)
	}
	if s, err := strconv.ParseUint(tmp[2], 10, 32); err == nil {
		dd += int(s)
	}
	return time.Date(yyyy, time.Month(mm), dd, 0, 0, 0, 0, TZ8)
}
func parseLatLon(v string) float64 {
	val := 0.0
	tmp := strings.SplitN(v, "°", 2)
	if s, err := strconv.ParseFloat(tmp[0], 64); err == nil {
		val += s
	}
	if len(tmp) != 2 {
		return val
	}

	tmp = strings.SplitN(tmp[1], "'", 2)
	if s, err := strconv.ParseFloat(tmp[0], 64); err == nil {
		val += s / 60.0
	}
	if len(tmp) != 2 {
		return val
	}

	tmp = strings.SplitN(tmp[1], `"`, 2)
	if s, err := strconv.ParseFloat(tmp[0], 64); err == nil {
		val += (s / 60.0) / 60.0
	}
	return val
}
func parseVal(v string) *jsonFloat {
	switch v {
	case "－", "--":
		return NewJsonFloat(math.NaN())
	case "":
		return nil
	}

	s, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil // TODO: error handle
	}
	return NewJsonFloat(s)
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
