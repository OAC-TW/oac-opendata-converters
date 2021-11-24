package main

/*
* 射擊/礙航資料 open data
* 將多個csv轉成json/geojson

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
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	inDir   = flag.String("i", "./csv/", "input path of CSV files")
	outFile = flag.String("o", "shootingobstruct.json", "path to save output file")

	outJson = flag.Bool("json", false, "output json format")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	// a list point to utf8 encoding csv files!!!
	urls = flag.String("urls", "urls.txt", "url list in file")
	UA   = flag.String("ua", "OAC bot", "User-Agent")

	verbosity = flag.Int("v", 3, "verbosity for app")
)

func main() {
	flag.Parse()

	if *inDir != "" {
		transDir(*inDir, *outFile)
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

	// parse url list
	url, err := parseUrls(*urls)
	if url == nil {
		Vln(2, "[urls]parse error", err)
		return
	}
	err = transUrls(url, dialFunc, *outFile)
	if err != nil {
		Vln(2, "[json]err", err)
	}
	Vln(3, "[json]ok")
}

func parseUrls(listFp string) ([]string, error) {
	fd, err := os.OpenFile(listFp, os.O_RDONLY, 0400)
	if err != nil {
		Vln(2, "[open]err", listFp, err)
		return nil, err
	}
	defer fd.Close()

	lines := make([]string, 0, 128)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		url := scanner.Text()
		urlT := strings.TrimSpace(url)
		if urlT == "" || strings.HasPrefix(urlT, "#") {
			continue
		}
		lines = append(lines, url)
	}

	if err := scanner.Err(); err != nil {
		return lines, err
	}

	return lines, nil
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

func transDir(inDir string, outFp string) error {
	dirFd, err := os.Open(inDir)
	if err != nil {
		return err
	}
	list, err := dirFd.Readdir(-1)
	dirFd.Close()
	if err != nil {
		return err
	}

	allRows := make([]*Row, 0, 1024)
	for idx, f := range list {
		if f.IsDir() {
			continue
		}

		fname := f.Name()
		if strings.ToLower(filepath.Ext(fname)) != ".csv" {
			continue // skip not csv
		}
		fp := filepath.Join(inDir, fname)
		fd, err := os.OpenFile(fp, os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", err)
			continue
		}

		rows, err := parseCSV(fd)
		if err != nil {
			Vln(2, "[parse]err", idx, fname, err)
			return err
		}
		Vln(3, "[data]", idx, fname, len(rows))
		fd.Close()

		allRows = append(allRows, rows...)
	}

	if *outJson {
		err = writeJson(allRows, outFp)
	} else {
		err = writeJson(NewGeojson(allRows), outFp)
	}
	if err != nil {
		return err
	}

	return nil
}

func transUrls(urls []string, dialFunc func(network, address string) (net.Conn, error), outFp string) error {
	allRows := make([]*Row, 0, 1024)
	for idx, url := range urls {
		fd, err := getUrlFd(url, dialFunc)
		if err != nil {
			Vln(2, "[get]err", url, err)
			continue
		}
		// defer fd.Close()
		Vln(3, "[get]start download...", url)

		rows, err := parseCSV(fd)
		if err != nil {
			Vln(2, "[parse]err", idx, url, err)
			return err
		}
		Vln(3, "[data]", idx, url, len(rows))
		fd.Close()

		allRows = append(allRows, rows...)
	}

	var err error
	if *outJson {
		err = writeJson(allRows, outFp)
	} else {
		err = writeJson(NewGeojson(allRows), outFp)
	}
	if err != nil {
		return err
	}

	return nil
}

// csv stream to one json
func transFd(fd io.Reader, outFp string) error {
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
	Date   string `json:"date,omitempty"`   // 時間
	Coords string `json:"coords,omitempty"` // 範圍

	x map[string]interface{} // other props
}

func (row *Row) MarshalJSON() ([]byte, error) {
	row.x["date"] = row.Date
	row.x["coords"] = row.Coords
	return json.Marshal(row.x)
}

type Feature Row

func (row *Feature) MarshalJSON() ([]byte, error) {
	type Geometry struct {
		Type   string          `json:"type"`        // Point
		Coords json.RawMessage `json:"coordinates"` // [[[[lon0, lat0], [lon1, lat1]]]]
	}
	aux := &struct {
		Type  string    `json:"type"` // Feature
		Geo   *Geometry `json:"geometry,omitempty"`
		Props *Row      `json:"properties,omitempty"`
	}{
		Type: "Feature",
		Geo: &Geometry{
			Type:   "MultiPolygon",
			Coords: json.RawMessage(row.Coords), //[]float64{row.Px, row.Py},
		},
		Props: (*Row)(row),
	}
	return json.Marshal(aux)
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

	colNames := []string{
		"Date",
		"Coords",
	}
	lut := make(map[string]int, 8)
	idx2name := make(map[int]string, 8)

	rows := make([]*Row, 0, 128)
	// return rows, nil
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

		if len(rec) < 2 {
			return rows, errors.New("data column change or not match schema!!!")
		}

		idx++
		if idx == 1 {
			// find index
			for _, colName := range colNames {
				if idx := findField(rec, colName); idx >= 0 {
					lut[colName] = idx
				}
			}
			for idx, colName := range rec {
				idx2name[idx] = colName
			}
			Vln(6, "[lut]", len(lut), lut, idx2name)
			continue
			// break
		}

		row := &Row{
			Date:   getString(rec, lut, "Date"),   // 時間
			Coords: getString(rec, lut, "Coords"), // 範圍
			x:      make(map[string]interface{}, 8),
		}

		// fill other props
		for idx, val := range rec {
			name, ok := idx2name[idx]
			if !ok {
				continue
			}
			switch name {
			case "Date":
			case "Coords":
			default:
				row.x[name] = val
			}
		}

		rows = append(rows, row)
	}
	return rows, nil
}

func getString(rec []string, lut map[string]int, name string) string {
	idx, ok := lut[name]
	if !ok {
		return ""
	}
	return rec[idx]
}

func findField(arr []string, name string) int {
	for i, v := range arr {
		if v == name {
			return i
		}
	}
	return -1
}

var TZ8 = time.FixedZone("UTC+8 Time", int((8 * time.Hour).Seconds()))

func parseTime(v string) time.Time {
	tmp := strings.SplitN(v, "/", 3)
	if len(tmp) != 3 {
		return time.Time{}
	}
	yyyy, mm, dd := 0, 0, 0
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

func parseIcons(v string, logInfo string) string {
	if v == "" {
		return v
	}
	list := strings.Split(v, ",")
	ids := make([]int, 0, len(list))
	for _, idStr := range list {
		s, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 32)
		if err != nil {
			Vln(3, "[parseIcon]err", logInfo, err)
			continue
		}
		ids = append(ids, int(s))
	}
	return strings.Trim(strings.Replace(fmt.Sprint(ids), " ", ",", -1), "[]")
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
