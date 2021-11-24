package main

/*
* 海面風波流
* 暫無open data
* 將json轉成json/geojson
* 依照測站只保留最新一筆資料
* 注意:
*   1. 來源資料編碼為big5
*   2. 剛跨月可能沒有任何資料
* url:
https://goocean.namr.gov.tw/OAC/?sta=O&aw=1&key=OAC&ff=txt
https://goocean.namr.gov.tw/OAC/?sta=B&aw=1&key=OAC&ff=txt
https://goocean.namr.gov.tw/OAC/?sta=C&aw=1&key=OAC&ff=txt
https://goocean.namr.gov.tw/OAC/?sta=R&aw=1&key=OAC&ff=txt
https://goocean.namr.gov.tw/OAC/?sta=Y&aw=1&key=OAC&ff=txt

潮境	O	121.808028,25.144056
鼻頭角	Y	121.907278,25.124667
蜜月灣	B	121.929139,24.948833
南灣	R	120.760000,21.945000
東吉嶼	C	119.648611,23.236111


```txt
站名：潮境資料浮標站
資料時間：202109
資料格式：觀測時間　YYYYMMDDHH，
　　　　　最大波高　#### (公分)　對應週期　##.# (秒)　尖峰週期　##.# (秒)，
　　　　　示性波高　#### (公分)　平均週期　##.# (秒)　主波向　### (度)，
　　　　　三秒陣風　##.# (公尺/秒)　平均風速　##.# (公尺/秒)　平均風向　### (度)，
　　　　　測站氣壓　####.# (百帕)　平均氣溫　##.# (度)　表面水溫　##.# (度)
          表層海流流速　##.# (公分/秒)　表層流向　### (度-去向)
##########,  ####,  ##.#,  ##.#,  ####,  ##.#,   ###,  ##.#,  ##.#,   ###,####.#,  ##.#,  ##.#,  ##.#,   ###
************************************************************************************************************

2021090100,      ,      ,   5.5,    59,   4.6,    78,   5.4,   3.6,   137,1010.8,  27.9,  27.4,  21.1,    37
2021090101,      ,      ,   4.5,    50,   4.2,    67,   6.1,   3.5,   127,1010.6,  27.8,  27.4,   8.3,     2
2021090102,      ,      ,   5.4,    51,   4.3,    56,   6.2,   3.6,   158,1010.2,  27.8,  27.4,   4.4,     1
2021090103,      ,      ,   4.8,    47,   4.2,    67,   6.4,   4.1,   157,1009.8,  27.7,  27.5,   2.6,   347
2021090104,      ,      ,   4.3,    44,   4.2,    90,   4.8,   2.0,   183,1009.0,  27.5,  27.5,   0.3,   108
2021090105,      ,      ,   4.4,    40,   4.2,    90,   8.1,   3.8,   176,1010.0,  27.5,  27.5,   2.6,   296

...

2021091817,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,
2021091818,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,
2021091819,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,
2021091820,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,      ,

...

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
					121.808028,
					25.144056
				]
			},
			"properties": {
				"station": "O",
				"stationName": "潮境",
				"lat": 25.144056,
				"lon": 121.808028,
				"time": "2021-09-14T15:00:00+08:00",
				"peakperiod": 7.5,
				"hs": 28,
				"avgperiod": 5.1,
				"wavedir": 78,
				"gust": 5.1,
				"ws": 4,
				"wd": 105,
				"pressure": 1008.7,
				"avgtemp": 28.6,
				"wstemp": 28.2,
				"flowrate": 21.2,
				"flowdir": 232
			}
		}
	]
}
```

*/

import (
	"bufio"
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
	inFile    = flag.String("i", "seasurfacestorm_raw.txt", "input txt file")
	outFile   = flag.String("o", "seasurfacestorm.json", "path to save output file")
	cacheFile = flag.String("c", "cache.json", "path to save cache file")

	outJson = flag.Bool("json", false, "output json format")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	url = flag.String("u", "https://goocean.namr.gov.tw/OAC/?sta=%v&aw=1&key=OAC&ff=txt", "url")
	UA  = flag.String("ua", "OAC bot", "User-Agent")

	verbosity = flag.Int("v", 3, "verbosity for app")
)

type StationInfo struct {
	id   string
	name string
	lon  float64
	lat  float64
}

func main() {
	flag.Parse()

	if *inFile != "" {
		fd, err := os.OpenFile(*inFile, os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", err)
			return
		}
		defer fd.Close()

		pts, err := transFd(fd, nil, nil)
		if err != nil {
			Vln(2, "[parse]err", err)
			return
		}
		Vln(2, "[save]", len(pts), pts)
		if *outJson {
			err = writeJson(pts, *outFile)
		} else {
			err = writeJson(NewGeojson(pts), *outFile)
		}
		if err != nil {
			Vln(2, "[json]err", err)
		}
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

	urlList := []*StationInfo{
		{"O", "潮境", 121.808028, 25.144056},
		{"Y", "鼻頭角", 121.907278, 25.124667},
		{"B", "蜜月灣", 121.929139, 24.948833},
		{"R", "南灣", 120.760000, 21.945000},
		{"C", "東吉嶼", 119.648611, 23.236111},
	}

	// load cache
	cache := NewCache()
	cache.Load(*cacheFile)

	pts := make([]*Row, 0, len(urlList))
	for i, obj := range urlList {
		aurl := fmt.Sprintf(*url, obj.id)
		Vln(3, "[get]start download...", i, obj.id, obj.name, aurl)

		fd, err := getUrlFd(aurl, dialFunc)
		if err != nil {
			Vln(2, "[get]err", i, obj.id, obj.name, aurl, err)
			return
		}
		Vln(3, "[get]download.end ..", i, obj.id, obj.name, aurl)

		vals, err := transFd(fd, cache, obj)
		if err != nil {
			Vln(2, "[json]err", err)
		}
		if vals[0] != nil {
			pts = append(pts, vals...)
		}
	}
	Vln(3, "[json]all ok", len(pts))

	var err error
	if *outJson {
		err = writeJson(pts, *outFile)
	} else {
		err = writeJson(NewGeojson(pts), *outFile)
	}
	if err != nil {
		Vln(2, "[json]err", err)
	}

	// flush cache
	cache.Flush(*cacheFile)
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

func writeJson(obj interface{}, outFp string) error {
	fdOut, err := os.OpenFile(outFp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Vln(2, "[open]err", outFp, err)
		return err
	}
	defer fdOut.Close()

	enc := json.NewEncoder(fdOut)
	enc.SetIndent("", "\t")
	err = enc.Encode(obj)
	if err != nil {
		Vln(2, "[json]err", err)
		return err
	}
	return nil
}

// data stream to one json
func transFd(fd io.ReadCloser, cache *Cache, info *StationInfo) ([]*Row, error) {
	defer fd.Close()

	rows, err := parseData(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return nil, err
	}
	sort.Stable(sort.Reverse(sortByTime(rows)))
	Vln(3, "[data]out", len(rows))

	if len(rows) > 0 {
		if info != nil {
			r := rows[0]
			r.Station, r.StationName, r.Lat, r.Lon = info.id, info.name, info.lat, info.lon

			// update cache
			if cache != nil {
				cache.Update(r)
			}
		}
		return rows[0:1], nil
	}

	// return cache
	if info != nil && cache != nil {
		r := (*cache)[info.id]
		return []*Row{r}, nil
	}
	return nil, nil
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
type Cache map[string]*Row

func (c *Cache) Load(fp string) error {
	fd, err := os.OpenFile(*inFile, os.O_RDONLY, 0400)
	if err != nil {
		return err
	}
	defer fd.Close()

	dec := json.NewDecoder(fd)
	err = dec.Decode(c)
	if err != nil && err != io.EOF { // skip EOF
		return err
	}
	return nil
}

func (c *Cache) Flush(fp string) error {
	return writeJson(c, fp)
}

func (c *Cache) Update(row *Row) {
	(*c)[row.Station] = row
}

func NewCache() *Cache {
	c := Cache(make(map[string]*Row, 8))
	return &c
}

type Row struct {
	Station     string  `json:"station,omitempty"`     // eg: O = 潮境
	StationName string  `json:"stationName,omitempty"` // eg: 潮境
	Lat         float64 `json:"lat"`                   // 緯度
	Lon         float64 `json:"lon"`                   // 經度

	Time           Time8      `json:"time"`                 // 時間
	WaveHeight     *jsonFloat `json:"waveheight,omitempty"` // 最大波高(cm)
	Period         *jsonFloat `json:"period,omitempty"`     // 對應週期(s)
	PeakPeriod     *jsonFloat `json:"peakperiod,omitempty"` // 尖峰週期(s)
	Hs             *jsonFloat `json:"hs,omitempty"`         // 示性波高(cm)
	AvgPeriod      *jsonFloat `json:"avgperiod,omitempty"`  // 平均週期(s)
	WaveDirectoin  *jsonFloat `json:"wavedir,omitempty"`    // 主波向(deg)
	Gust           *jsonFloat `json:"gust,omitempty"`       // 三秒陣風(m/s)
	AvgWS          *jsonFloat `json:"ws,omitempty"`         // 平均風速(m/s)
	WD             *jsonFloat `json:"wd,omitempty"`         // 平均風向(deg)
	Pressure       *jsonFloat `json:"pressure,omitempty"`   // 測站氣壓(hPa)
	AvgTemperature *jsonFloat `json:"avgtemp,omitempty"`    // 平均氣溫(deg)

	// 表層海流
	WaterSurfaceTemperature *jsonFloat `json:"wstemp,omitempty"`   // 表面水溫(deg)
	FlowRate                *jsonFloat `json:"flowrate,omitempty"` // 流速(cm/s)
	FlowDirectoin           *jsonFloat `json:"flowdir,omitempty"`  // 表層流向(deg-去向)
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

func (s sortByTime) Len() int      { return len(s) }
func (s sortByTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTime) Less(i, j int) bool {
	return time.Time(s[i].Time).Before(time.Time(s[j].Time))
}

var TZ8 = time.FixedZone("UTC+8 Time", int((8 * time.Hour).Seconds()))

type Time8 time.Time

func (v *Time8) UnmarshalJSON(in []byte) error {
	tIn := strings.Trim(string(in), `"`)
	if tIn == "null" {
		v = nil
		return nil
	}
	t, err := time.ParseInLocation("2006-01-02T15:04:05.000", tIn, TZ8)
	if err != nil {
		t, err = time.ParseInLocation("2006-01-02T15:04:05", tIn, TZ8)
		if err != nil {
			return err
		}
	}
	*v = Time8(t)
	return nil
}
func (v *Time8) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(*v))
}

type jsonFloat float64

func (value jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return []byte("\"--\""), nil
	}
	return []byte(fmt.Sprintf("%v", value)), nil
}

func NewJsonFloat(v float64) *jsonFloat {
	val := new(jsonFloat)
	*val = jsonFloat(v)
	return val
}

func parseData(r io.Reader) ([]*Row, error) {
	rows := make([]*Row, 0, 744) // 24 * 31
	startDataLine := false
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !startDataLine {
			if strings.Count(line, "*") > 80 { // TODO: not hardcode
				startDataLine = true
			}
			continue
		}
		if line == "" {
			continue
		}

		fields := strings.SplitN(line, ",", 15)
		Vln(5, "[data]", line, len(fields))
		if len(fields) != 15 {
			return rows, errors.New("data schema may change!!!!")
		}
		vals, ok := TrimAndCheck(fields[1:])
		if !ok {
			break
		}
		row := &Row{
			Time:                    parseTime(fields[0]),
			WaveHeight:              vals[0],
			Period:                  vals[1],
			PeakPeriod:              vals[2],
			Hs:                      vals[3],
			AvgPeriod:               vals[4],
			WaveDirectoin:           vals[5],
			Gust:                    vals[6],
			AvgWS:                   vals[7],
			WD:                      vals[8],
			Pressure:                vals[9],
			AvgTemperature:          vals[10],
			WaterSurfaceTemperature: vals[11],
			FlowRate:                vals[12],
			FlowDirectoin:           vals[13],
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func TrimAndCheck(fields []string) ([]*jsonFloat, bool) {
	hasData := false
	out := make([]*jsonFloat, 0, len(fields))
	for _, v := range fields {
		if f64, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			out = append(out, NewJsonFloat(f64))
			hasData = true
		} else {
			out = append(out, nil)
		}
	}
	return out, hasData
}

func parseTime(data string) Time8 {
	t, err := time.ParseInLocation("2006010215", data, TZ8)
	if err != nil {
		Vln(3, "[parseTime]err", data, err)
	}
	return Time8(t)
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
