package main

/*
* 中央氣象局open data
* 波浪預報模式資料-臺灣海域預報資料
* https://opendata.cwb.gov.tw/dataset/climate/F-A0020-001
* 將2D經緯度資料轉為1D-array
* 經緯度 7, 119 >> 7, 126; 7.1, 119 >> 7.1, 126; .... ; 36, 126
*/

import (
	"fmt"
	"flag"
	"log"
	"time"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"bytes"
	"regexp"

	"archive/zip"
	"path/filepath"
	"sort"

	"encoding/json"
	"encoding/xml"
	"math"
)

var (
	inFile = flag.String("i", "F-A0020-001.zip", "input XML in zip file")
	//outFile = flag.String("o", "20072318.000.grid.json", "output file")


	proxyAddr = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	cpu = flag.Int("cpu", 0, "CPU count limit, 0 == auto")

	token = flag.String("auth", "CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX", "token") // 氣象局open data的API授權碼
//	token = flag.String("auth", "", "token")
	url = flag.String("u", "https://opendata.cwb.gov.tw/fileapi/v1/opendataapi/F-A0020-001?Authorization=%v&downloadType=WEB&format=ZIP", "url")
	UA = flag.String("ua", "OAC bot", "User-Agent")

	outDir = flag.String("dir", "json/", "path to save output file")

	verbosity = flag.Int("v", 3, "verbosity for app")

	xmlRx = regexp.MustCompile(`([0-9]{8,8})-([dhirst]{1,3})\.([0-9]{3,3})\.xml`) // name in zip
	jsonRx = regexp.MustCompile(`([0-9]{8,8})\.([0-9]{3,3})\.grid\.json`) // name for old output
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*cpu) // simple cpu core count limit

	if *token == "" {
		//transFile(*inFile, *outFile)

		/*fdDir, err := os.OpenFile(*inFile + "-dir.000.xml", os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", "dir", err)
			return
		}
		defer fdDir.Close()

		fdHs, err := os.OpenFile(*inFile + "-hs.000.xml", os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", "hs", err)
			return
		}
		defer fdHs.Close()

		fdT, err := os.OpenFile(*inFile + "-t.000.xml", os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", "t", err)
			return
		}
		defer fdT.Close()

		of, err := os.OpenFile(*outFile, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			Vln(2, "[open]err", *outFile, err)
			return
		}
		defer of.Close()

		transFd(fdDir, fdHs, fdT, of)*/

		fd, err := os.OpenFile(*inFile, os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", err)
			return
		}
		defer fd.Close()

		readZipAndExtract(fd, *outDir)
		return
	}

	aurl := fmt.Sprintf(*url, *token)
	dialFunc := func(network, address string) (net.Conn, error) {
		return net.DialTimeout("tcp", address, time.Duration(*connTimeout) * time.Second)
	}
	if *proxyAddr != "" {
		dialFunc = func(network, address string) (net.Conn, error) {
			if network != "tcp" {
				return nil, errors.New("only support tcp")
			}
			return makeConnection(address, *proxyAddr, time.Duration(*connTimeout) * time.Second)
		}
	}
	
	fd, err := getUrlFd(aurl, dialFunc)
	if err != nil {
		Vln(2, "[get]err", aurl, err)
		return
	}
	defer fd.Close()

	Vln(3, "[get]start download...", aurl)

	err = readZipAndExtract(fd, *outDir)
	if err != nil {
		Vln(2, "[json]err", err)
	}
	Vln(3, "[json]ok")
}

func getUrlFd(url string, dialFunc func(network, addr string) (net.Conn, error)) (io.ReadCloser, error) {
	var netTransport = &http.Transport{
		Dial: dialFunc,
		TLSHandshakeTimeout: time.Duration(*connTimeout) * time.Second,
	}

	var netClient = &http.Client{
		Timeout: time.Second * 180,
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


// one xml to one json
func transFile(inFp string, outFp string) error {
	fd, err := os.OpenFile(inFp, os.O_RDONLY, 0400)
	if err != nil {
		Vln(2, "[open]err", *inFile, err)
		return err
	}
	defer fd.Close()

	grid, err := parseXML(fd)
	if err != nil {
		Vln(2, "[parse]err", err)
		return err
	}
	Vln(3, "[grid]", grid.Nx, grid.Ny, grid.Lo1, grid.La1, grid.Lo2, grid.La2)

	of, err := os.OpenFile(outFp, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		Vln(2, "[open]err", *outFile, err)
		return err
	}
	defer of.Close()

	enc := json.NewEncoder(of)
	err = enc.Encode(grid)
	if err != nil {
		Vln(2, "[json]err", err)
		return err
	}
	return nil
}

// (dir + hs + t) xml stream to json stream
func transFd(fdDir io.Reader, fdHs io.Reader, fdT io.Reader, fdOut io.Writer) (*VectorGrid, error) {
	var wg sync.WaitGroup

	fds := []io.Reader{fdDir, fdHs, fdT}
	retCh := make(chan *VectorGrid, 1)
	for _, fd := range fds {
		wg.Add(1)
		go func(fd io.Reader) {
			defer wg.Done()

			grid, err := parseXML(fd)
			if err != nil {
				Vln(2, "[parse]err", err)
				return
			}
			Vln(3, "[grid]000", grid.Nx, grid.Ny, grid.Lo1, grid.La1, grid.Lo2, grid.La2)
			retCh <- grid
		}(fd)
	}

	var gridDir, gridHs, gridT *VectorGrid
	endCh := make(chan struct{})
	go func() {
		for grid := range retCh {
			_, ok := grid.Data["浪向"]
			if ok {
				gridDir = grid
				continue
			}
			_, ok = grid.Data["浪高"]
			if ok {
				gridHs = grid
				continue
			}
			_, ok = grid.Data["週期"]
			if ok {
				gridT = grid
				continue
			}
		}
		close(endCh)
	}()

	wg.Wait()
	close(retCh)

	// wait all done
	<- endCh

	Vln(3, "[grid]Dir", gridDir.Nx, gridDir.Ny, gridDir.Lo1, gridDir.La1, gridDir.Lo2, gridDir.La2)
	Vln(3, "[grid]HS", gridHs.Nx, gridHs.Ny, gridHs.Lo1, gridHs.La1, gridHs.Lo2, gridHs.La2)
	Vln(3, "[grid]T", gridT.Nx, gridT.Ny, gridT.Lo1, gridT.La1, gridT.Lo2, gridT.La2)

	grid := gridDir
	grid.Desc = grid.Desc + ";" + gridHs.Desc
	grid.Desc = grid.Desc + ";" + gridT.Desc
	grid.Data["浪高"] = gridHs.Data["浪高"]
	grid.Data["週期"] = gridT.Data["週期"]

	grid.DataRange["浪高"] = gridHs.DataRange["浪高"]
	grid.DataRange["週期"] = gridT.DataRange["週期"]

	enc := json.NewEncoder(fdOut)
	err := enc.Encode(grid)
	if err != nil {
		Vln(2, "[json]err", err)
		return nil, err
	}
	return grid, nil
}


func readZipAndExtract(conn io.ReadCloser, dirOut string) error {
	// copy to in-memory buffer
	var b bytes.Buffer
	_, err := io.Copy(&b, conn)
	if err != nil {
		return err
	}
	Vln(3, "[get]download end")

	// list old file for clean up
	oldFiles, err := readDir(dirOut)
	if err != nil {
		Vln(2, "[proc]list old data", err)
		return err
	}

	// unzip & output
	list, err := unzip(b.Bytes(), dirOut)
	if err != nil {
		return err
	}

	// update index.json
	err = updateIndex(filepath.Join(dirOut, "index.json"), list)
	if err != nil {
		Vln(2, "[proc]update index", err)
		return err
	}

	for _, f := range list {
		k := f.Name
		if oldFiles[k] {
			delete(oldFiles, k)
		}
	}

	// remove old file for clean up
	err = removeFiles(dirOut, oldFiles)
	if err != nil {
		Vln(2, "[proc]clean up old data", err)
		return err
	}

	return nil
}

type IndexFile struct {
	TimeUTC time.Time `json:"timeUTC"`
	Time08  time.Time `json:"time08"`
	Name string `json:"name"`

	DataRange map[string][]jsonFloat `json:"drange"`

	fileDir *zip.File
	fileHs *zip.File
	fileT *zip.File
}

func updateIndex(outFp string, list []*IndexFile) error {
	Vln(6, "[idx]count", len(list))
	for _, item := range list {
		Vln(6, "[idx]", item)
	}
	buf, err := json.Marshal(list)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outFp, buf, 0644)
}

type sortByTime []*IndexFile
func (s sortByTime) Len() int      { return len(s) }
func (s sortByTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByTime) Less(i, j int) bool { return s[i].TimeUTC.Before(s[j].TimeUTC) }

func unzip(zbuff []byte, out string) ([]*IndexFile, error) {
	zr := bytes.NewReader(zbuff)
	r, err := zip.NewReader(zr, int64(len(zbuff)))
	if err != nil {
		return nil, err
	}

	loc := time.FixedZone("UTC+8", +8*60*60)
	now := time.Now().UTC()

	listSeq := make([]*IndexFile, 0, 294)
	list := make(map[string]*IndexFile, 294)
	for _, f := range r.File {
		base := filepath.Base(f.Name)
		Vln(3, "[zip][file]", f.Name, f.CompressedSize64, f.UncompressedSize64)

		switch filepath.Ext(base) {
		case ".xml":
		default: // skip any other
			continue
		}

		// check name & time
		reOut := xmlRx.FindStringSubmatch(base)
		Vln(7, "[zip][rx]", f.Name, reOut)
		if len(reOut) != 4 { // ["20072318-dir.000.xml" "20072318", "dir", "000"]
			continue // skip
		}
		datetimeStr := reOut[1]
		typeStr := reOut[2]
		offsetStr := reOut[3]

		var yyyy, MM, DD, HH, offset int
		if s, err := strconv.ParseInt(datetimeStr[0:2], 10, 32); err == nil {
			yyyy = 2000 + int(s)
		}
		if s, err := strconv.ParseInt(datetimeStr[2:4], 10, 32); err == nil {
			MM = int(s)
		}
		if s, err := strconv.ParseInt(datetimeStr[4:6], 10, 32); err == nil {
			DD = int(s)
		}
		if s, err := strconv.ParseInt(datetimeStr[6:8], 10, 32); err == nil {
			HH = int(s)
		}
		if s, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int(s)
		}
		t0 := time.Date(yyyy, time.Month(MM), DD, HH, 0, 0, 0, time.UTC).Add(time.Duration(offset) * time.Hour)
		t8 := t0.In(loc)

		nameJson := fmt.Sprintf("%v.%v.grid.json", datetimeStr, offsetStr)
		item, ok := list[nameJson]
		if !ok {
			item = &IndexFile{
				TimeUTC: t0,
				Time08: t8,
				Name: nameJson,
			}
			list[nameJson] = item
			listSeq = append(listSeq, item)
		}

		switch typeStr {
		case "dir":
			item.fileDir = f
		case "hs":
			item.fileHs = f
		case "t":
			item.fileT = f
		}
	}
	Vln(6, "[zip][xml]count", len(list), len(listSeq))

	// sort
	sort.Sort(sortByTime(listSeq))

	// remove old data
	for i, f := range listSeq {
		if f.TimeUTC.After(now) {
			i = i - 1
			if i < 0 {
				i = 0
			}
			for _, fs := range listSeq[:i] {
				delete(list, fs.Name)
			}
			listSeq = listSeq[i:]
			break
		}
	}

	// write file
	for _, f := range list {
		grid, err := unzipAndTransXML(f, out)
		if err != nil {
			continue // skip
			//return nil, err
		}
		f.DataRange = grid.DataRange
	}
	return listSeq, nil
}

func unzipAndTransXML(f *IndexFile, outDir string) (*VectorGrid, error) {
	rcDir, err := f.fileDir.Open()
	if err != nil {
		return nil, err
	}
	defer rcDir.Close()

	rcHs, err := f.fileHs.Open()
	if err != nil {
		return nil, err
	}
	defer rcHs.Close()

	rcT, err := f.fileT.Open()
	if err != nil {
		return nil, err
	}
	defer rcT.Close()

	outFp := filepath.Join(outDir, f.Name)
	oFd, err := os.OpenFile(outFp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		Vln(2, "[zip]open output fail", err)
		return nil, err
	}
	defer oFd.Close()

	grid, err := transFd(rcDir, rcHs, rcT, oFd)
	if err != nil {
		return nil, err
	}
	return grid, nil
}

func readDir(dirname string) (map[string]bool, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	Vln(2, "[cache]old data", dirname, len(list), list)

	// filter out non-json
	out := make(map[string]bool, len(list))
	for _, fi := range list {
		if jsonRx.MatchString(fi.Name()) {
			out[fi.Name()] = true
		}
	}

	return out, nil
}

func removeFiles(basePath string, list map[string]bool) error {
	Vln(6, "[clean]old data", len(list), list)
	for name, _ := range list {
		fp := filepath.Join(basePath, name)
		err := os.Remove(fp)
		if err != nil {
			Vln(2, "[clean]remove file fail", fp, err)
			//return err
		}
	}
	return nil
}


// ==== proc XML ====
type VectorGrid struct {
	// 原點 經度, 緯度
	Lo1 float32 `json:"lo1"`
	La1 float32 `json:"la1"`

	// 終點 經度, 緯度
	Lo2 float32 `json:"lo2"`
	La2 float32 `json:"la2"`

	Nx int `json:"nx"` // 經度格數
	Ny int `json:"ny"` // 緯度格數

	Time string `json:"time"` // just copy now
	Desc string `json:"Description"`  // just copy

	DataRange map[string][]jsonFloat `json:"drange"`

	Data map[string][]jsonFloat `json:"d"`
}

type jsonFloat float32
func (value jsonFloat) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(value)) {
		return []byte("\"\""), nil
	}
	return []byte(fmt.Sprintf("%v", value)), nil
}

func NewVectorGrid() *VectorGrid {
	vg := &VectorGrid{}
	vg.Data = make(map[string][]jsonFloat, 2)
	vg.DataRange = make(map[string][]jsonFloat, 2)
	return vg
}

func parseXML(r io.Reader) (*VectorGrid, error) {
	grid := NewVectorGrid()

	ps := &procState{}
	xs := NewXMLState()
	decoder := xml.NewDecoder(r)
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return grid, nil
			}
			return grid, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			stelm := xml.StartElement(t)
			//fmt.Println("start: ", stelm.Name.Local)
			xs.StartTag(stelm)

		case xml.EndElement:
			endelm := xml.EndElement(t)
			//Vln(5, "end: ", endelm.Name.Local)
			xs.EndTag(endelm)

		case xml.CharData:
			data := xml.CharData(t)
			ps.FillTag(xs, data, grid)

			//str := string(data)
			//Vln(5, "[val]", xs.GetPath(), str)
		}
	}

	return grid, nil
}

type procState struct {
	st int
	valName string

	lat float32
	lon float32
	lat0 float32
	lon0 float32
	lat1 float32
	lon1 float32

	latStr string
	lonStr string
	latIdx map[string]bool
	lonIdx map[string]bool
	buf map[string]map[string]map[string]jsonFloat // type >> lat >> lon
}
func (ps *procState) FillTag(xs *XmlState, data []byte, grid *VectorGrid) {
	switch ps.st {
	case 0:
		path := xs.GetPath()
		switch path {
		case "cwbopendata/dataset/datasetInfo/datasetDescription":
			str := string(data)
			Vln(3, "[desc]", path, str)
			grid.Desc = str
		case "cwbopendata/dataset/datasetInfo/parameterSet/parameter/parameterName":
			str := string(data)
			Vln(4, "[parmName]", path, str)
			ps.valName = str
		case "cwbopendata/dataset/datasetInfo/parameterSet/parameter/parameterValue":
			str := string(data)
			Vln(4, "[parmVal]", path, str)
			if v, err := strconv.ParseUint(str, 10, 32); err == nil {
				switch ps.valName {
				case "經度格點數":
					grid.Nx = int(v)
				case "緯度格點數":
					grid.Ny = int(v)
				default:
				}
			}
		case "cwbopendata/dataset/time/datetime", "cwbopendata/dataset/time/dataTime":
			str := string(data)
			Vln(3, "[time]", path, str)
			grid.Time = str
		case "cwbopendata/dataset/location":
			ps.st = 1 // start parse grids
			ps.lat0 = 9999
			ps.lon0 = 9999
			ps.lat1 = -9999
			ps.lon1 = -9999

			// make 2D array
			if grid.Nx > 0 && grid.Ny > 0 {
				ps.buf = make(map[string]map[string]map[string]jsonFloat)
				ps.latIdx = make(map[string]bool, grid.Ny)
				ps.lonIdx = make(map[string]bool, grid.Nx)
			}
		}
	case 1:
		tag := xs.LastPath()
		switch tag {
		case "lat": // 緯度
			str := string(data)
			if v, err := strconv.ParseFloat(str, 32); err == nil {
				ps.lat = float32(v)
				ps.latStr = str
				if ps.lat < ps.lat0 {
					ps.lat0 = ps.lat
				}
				if ps.lat > ps.lat1 {
					ps.lat1 = ps.lat
				}
			}
		case "lon": // 經度
			str := string(data)
			if v, err := strconv.ParseFloat(str, 32); err == nil {
				ps.lon = float32(v)
				ps.lonStr = str
				if ps.lon < ps.lon0 {
					ps.lon0 = ps.lon
				}
				if ps.lon > ps.lon1 {
					ps.lon1 = ps.lon
				}
			}
		case "elementName":
			str := string(data)
			switch str {
			case "浪向", "浪高", "週期":
				ps.valName = str
			default:
				ps.valName = ""
			}
		case "value":
			if ps.valName == "" {
				break
			}
			v, err := strconv.ParseFloat(string(data), 32)
			if err != nil {
				break
			}

			arr, ok := grid.Data[ps.valName]
			if !ok {
				arr = make([]jsonFloat, 0, grid.Ny)
			}
			//arr = append(arr, jsonFloat(v))
			grid.Data[ps.valName] = arr

			if !math.IsNaN(v) {
				minMax, ok := grid.DataRange[ps.valName]
				if !ok {
					minMax = []jsonFloat{jsonFloat(v), jsonFloat(v)}
					grid.DataRange[ps.valName] = minMax
				}
				if v < float64(minMax[0]) {
					minMax[0] = jsonFloat(v)
				}
				if v > float64(minMax[1]) {
					minMax[1] = jsonFloat(v)
				}
			}

			if ps.buf == nil {
				break
			}

			arr2d, ok := ps.buf[ps.valName]
			if !ok {
				arr2d = make(map[string]map[string]jsonFloat)
				ps.buf[ps.valName] = arr2d
			}
			rows, ok := arr2d[ps.latStr]
			if !ok {
				rows = make(map[string]jsonFloat)
				arr2d[ps.latStr] = rows
			}
			rows[ps.lonStr] = jsonFloat(v)

			ps.latIdx[ps.latStr] = true
			ps.lonIdx[ps.lonStr] = true

		case "cwbopendata": // end dataset
			ps.st = 2
			//grid.Nx = len(grid.Data["X"]) / grid.Ny

			// min >> max
			grid.Lo1 = ps.lon0
			grid.Lo2 = ps.lon1

			// max >> min
			grid.La1 = ps.lat1
			grid.La2 = ps.lat0

			if grid.Nx > 0 && grid.Ny > 0 {
				for k, arr2d := range ps.buf {
					grid.Data[k] = transTo1D(arr2d, ps.latIdx, ps.lonIdx)
				}
				grid.Nx = len(ps.lonIdx)
				grid.Ny = len(ps.latIdx)
			} else {
				for k, arr := range grid.Data {
					grid.Data[k] = transT(arr, grid.Ny)
				}
			}

			Vln(3, "[grid]", ps.lat, ps.lon, len(grid.Data["浪向"]), len(grid.Data["浪高"]), len(grid.Data["週期"]))
		}
	}
}

type sortByNumberString []string
func (s sortByNumberString) Len() int      { return len(s) }
func (s sortByNumberString) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortByNumberString) Less(i, j int) bool {
	li := len(s[i])
	lj := len(s[j])
	si := s[i]
	sj := s[j]
	if li < lj {
		si = perpend(si, lj)
	}
	if li > lj {
		sj = perpend(sj, li)
	}
	return si < sj
}
func perpend(str string, size int) string {
	buf := make([]byte, 0, size)
	sz := len(str)
	for n := size - sz; n>0; n-- {
		buf = append(buf, '0')
	}
	buf = append(buf, []byte(str)...)
	return string(buf)
}
func transTo1D(arr2d map[string]map[string]jsonFloat, yAxis map[string]bool, xAxis map[string]bool) []jsonFloat {
	ny := len(yAxis)
	nx := len(xAxis)
	latS := make([]string, 0, ny) // == Ny
	lonS := make([]string, 0, nx) // == Nx
	for str, _ := range yAxis {
		latS = append(latS, str)
	}
	for str, _ := range xAxis {
		lonS = append(lonS, str)
	}
	sort.Sort(sortByNumberString(latS))
	sort.Sort(sortByNumberString(lonS))

	Vln(3, "[transTo1D]", ny, nx, len(latS), len(lonS))

	out := make([]jsonFloat, 0, ny * nx)
	for _, lat := range latS {
		row, ok := arr2d[lat]
		if !ok { // empty
			out = append(out, make([]jsonFloat, nx)...)
			continue
		}
		for _, lon := range lonS {
			v, ok := row[lon]
			if !ok { // empty
				v = jsonFloat(math.NaN())
			}
			out = append(out, v)
		}
	}
	return out
}

func transT(in []jsonFloat, stride int) []jsonFloat {
	sz := len(in)
	stride2 := sz / stride
	out := make([]jsonFloat, sz, sz)
	for i, v := range in {
		a := i / stride
		b := i % stride
		idx := a + b * stride2
		out[idx] = v
	}
	return out
}

type XmlState struct {
	Path []string
}

func NewXMLState() *XmlState {
	xs := &XmlState{}
	xs.Path = make([]string, 0, 32)
	return xs
}

func (xs *XmlState) StartTag(t xml.StartElement) {
	xs.Path = append(xs.Path, string(t.Name.Local))
}

func (xs *XmlState) EndTag(t xml.EndElement) {
	sz := len(xs.Path)
	xs.Path = xs.Path[:sz-1]
}

func (xs *XmlState) GetPath() string {
	return strings.Join(xs.Path, "/")
}

func (xs *XmlState) LastPath() string {
	sz := len(xs.Path) - 1
	if sz < 0 {
		return ""
	}
	return xs.Path[sz]
}

func (xs *XmlState) PathLevel() int {
	return len(xs.Path)
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

