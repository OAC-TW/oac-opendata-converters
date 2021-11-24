package main

/*
* 海域水質
* 暫無open data
* https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnWaterQuality.ashx?code=D0B4699202A04DAD8A3153E48B6F937E
* 將json轉成json/geojson
* 依照測點名稱只保留最新一筆資料

```json
[
  {
    "VALUE_ID": 8077,
    "TYPE": "馬祖沿海海域",
    "STATION_NAME": "南竿鄉北部沿海",
    "LOCATION": "連江縣",
    "LAT": 26.173927,
    "LON": 119.924698,
    "BODY_LEVEL": "乙",
    "SAMPLE_DATE": null,
    "STATION_ID": "5933",
    "DEPTH": "1",
    "TEM_AIR": "ND",
    "TEM_WATER": "ND",
    "SALINITY": "ND",
    "PH": "ND",
    "DO_TIT": "",
    "DO_ELE": "ND",
    "DO_S": "ND",
    "SS": "ND",
    "CHL_a": "ND",
    "NH3_N": "ND",
    "NO3_N": "ND",
    "MI3PO4": "ND",
    "NO2_N": "ND",
    "SiO2": "ND",
    "Cd": "ND",
    "Cr": "ND",
    "Cu": "ND",
    "Zn": "ND",
    "Pb": "ND",
    "Hg": "ND",
    "Memo": "無法採樣",
    "UPDATE_TIME": "2021-08-17T12:00:00"
  },
  {
    "VALUE_ID": 734,
    "TYPE": "東石布袋沿海海域",
    "STATION_NAME": "東石港外海一",
    "LOCATION": "嘉義縣",
    "LAT": null,
    "LON": null,
    "BODY_LEVEL": "甲",
    "SAMPLE_DATE": "2002/03/01",
    "STATION_ID": "5178",
    "DEPTH": "1",
    "TEM_AIR": "",
    "TEM_WATER": "22.9",
    "SALINITY": "34.1",
    "PH": "8.2",
    "DO_TIT": "7.2",
    "DO_ELE": "",
    "DO_S": "",
    "SS": "14.1",
    "CHL_a": "6.5",
    "NH3_N": "",
    "NO3_N": "ND",
    "MI3PO4": "0.037",
    "NO2_N": "0.005",
    "SiO2": "0.21",
    "Cd": "ND",
    "Cr": "ND",
    "Cu": "0.0021",
    "Zn": "0.0097 ",
    "Pb": "ND ",
    "Hg": "ND",
    "Memo": "",
    "UPDATE_TIME": "2020-10-23T08:59:01.917"
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
					119.924698,
					26.173927
				]
			},
			"properties": {
				"VALUE_ID": 8125,
				"TYPE": "臺東沿海海域",
				"LOCATION": "",
				"STATION_NAME": "綠島",
				"lat": 26.173927,
				"lon": 119.924698,
				"BODY_LEVEL": "甲",
				"SAMPLE_DATE": "2021/05/12",
				"STATION_ID": "5211",
				"DEPTH": "1",
				"TEM_AIR": 28.7,
				"TEM_WATER": 27.8,
				"SALINITY": 34.5,
				"PH": 8.27,
				"DO_TIT": "ND",
				"DO_ELE": 6.8,
				"DO_S": 106.2,
				"SS": "ND",
				"CHL_a": 0.2,
				"NH3_N": "ND",
				"NO3_N": "ND",
				"MI3PO4": "ND",
				"NO2_N": "ND",
				"SiO2": "ND",
				"Cd": "ND",
				"Cr": "ND",
				"Cu": "ND",
				"Zn": 0.002,
				"Pb": "ND",
				"Hg": "ND",
				"Memo": "",
				"UPDATE_TIME": "2021-08-17T12:00:00+08:00"
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
	inFile  = flag.String("i", "waterquality_raw.json", "input json file")
	outFile = flag.String("o", "waterquality.json", "path to save output file")

	outJson = flag.Bool("json", false, "output json format")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	url = flag.String("u", "https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnWaterQuality.ashx?code=D0B4699202A04DAD8A3153E48B6F937E", "url")
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

// data stream to one json
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
	ID           int               `json:"VALUE_ID"`     // useless
	TYPE         string            `json:"TYPE"`         // 採樣分區
	LOCATION     string            `json:"LOCATION"`     // 縣市
	STATION_NAME string            `json:"STATION_NAME"` // 測站
	LAT          jsonNullableFloat `json:"lat"`          // 緯度
	LON          jsonNullableFloat `json:"lon"`          // 經度
	BODY_LEVEL   string            `json:"BODY_LEVEL"`   // 海域標準
	SAMPLE_DATE  string            `json:"SAMPLE_DATE"`  // 採樣日期
	STATION_ID   string            `json:"STATION_ID"`   // 測站編號
	DEPTH        string            `json:"DEPTH"`        // 深度(m)?
	TEM_AIR      *jsonFloat        `json:"TEM_AIR"`      // 氣溫
	TEM_WATER    *jsonFloat        `json:"TEM_WATER"`    // 水溫
	SALINITY     *jsonFloat        `json:"SALINITY"`     // 鹽度
	PH           *jsonFloat        `json:"PH"`           // PH
	DO_TIT       *jsonFloat        `json:"DO_TIT"`       // 溶氧(滴定定量法)
	DO_ELE       *jsonFloat        `json:"DO_ELE"`       // 溶氧(電極法)
	DO_S         *jsonFloat        `json:"DO_S"`         // 溶氧飽和度
	SS           *jsonFloat        `json:"SS"`           // 懸浮固體
	CHL_a        *jsonFloat        `json:"CHL_a"`        // 葉綠素a
	NH3_N        *jsonFloat        `json:"NH3_N"`        // 氨氮
	NO3_N        *jsonFloat        `json:"NO3_N"`        // 硝酸鹽氮
	MI3PO4       *jsonFloat        `json:"MI3PO4"`       // 正磷酸鹽
	NO2_N        *jsonFloat        `json:"NO2_N"`        // 亞硝酸鹽氮
	SiO2         *jsonFloat        `json:"SiO2"`         // 矽酸鹽
	Cd           *jsonFloat        `json:"Cd"`           // 鎘
	Cr           *jsonFloat        `json:"Cr"`           // 鉻
	Cu           *jsonFloat        `json:"Cu"`           // 銅
	Zn           *jsonFloat        `json:"Zn"`           // 鋅
	Pb           *jsonFloat        `json:"Pb"`           // 鉛
	Hg           *jsonFloat        `json:"Hg"`           // 汞
	Memo         string            `json:"Memo"`         // 備註
	UPDATE_TIME  Time8             `json:"UPDATE_TIME"`  // 更新日期
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
			Coords: []float64{*row.LON.V, *row.LAT.V},
		},
		Props: (*Row)(row),
	}
	return json.Marshal(aux)
}

type sortByTime []*Row

func (s sortByTime) Len() int      { return len(s) }
func (s sortByTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTime) Less(i, j int) bool {
	return time.Time(s[i].UPDATE_TIME).Before(time.Time(s[j].UPDATE_TIME))
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

func (value *jsonFloat) UnmarshalJSON(in []byte) error {
	valS := strings.TrimSpace(strings.Trim(string(in), `"`))
	switch valS {
	case "", "ND":
		*value = *NewJsonFloat(math.NaN())
		return nil
	}
	v, err := strconv.ParseFloat(valS, 64)
	if err != nil {
		return err
	}
	*value = *NewJsonFloat(v)
	return nil
}
func (value jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return []byte("\"ND\""), nil
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

type jsonNullableFloat struct {
	V *float64
}

func NewjsonNullableFloat(v float64) *jsonNullableFloat {
	val := new(float64)
	*val = v
	return &jsonNullableFloat{
		V: val,
	}
}

func NewNullFloat() *jsonNullableFloat {
	return &jsonNullableFloat{
		V: nil,
	}
}

func (value *jsonNullableFloat) UnmarshalJSON(in []byte) error {
	valS := string(in)
	switch valS {
	case "null":
		*value = jsonNullableFloat{V: nil}
		return nil
	}
	v, err := strconv.ParseFloat(valS, 64)
	if err != nil {
		return err
	}
	*value = *NewjsonNullableFloat(v)
	return nil
}
func (value *jsonNullableFloat) MarshalJSON() ([]byte, error) {
	if value == nil {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("%v", *value.V)), nil
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
		ok := exist[row.STATION_ID]
		if ok {
			continue
		}
		exist[row.STATION_ID] = true
		out = append(out, row)
	}

	for _, row := range out {
		if row.LAT.V != nil && row.LON.V != nil {
			continue
		}
		lat, lon, ok := lookUPLatLon(rows, row.STATION_ID)
		if !ok {
			// TODO: look up table
			panic("No lat lon data!!")
		}
		row.LAT = *NewjsonNullableFloat(lat)
		row.LON = *NewjsonNullableFloat(lon)
	}
	return out
}

func lookUPLatLon(rows []*Row, stationID string) (float64, float64, bool) {
	for _, row := range rows {
		if row.LAT.V != nil && row.LON.V != nil {
			return *row.LAT.V, *row.LON.V, true
		}
	}
	return 0, 0, false
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
