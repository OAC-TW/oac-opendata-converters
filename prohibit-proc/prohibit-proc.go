package main

/*
* 禁止海域 open data
* 將多個geojson合併成一個geojson

 */

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
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
	inDir   = flag.String("i", "./geojson/", "input path of geojson files")
	outFile = flag.String("o", "prohibit.json", "path to save output file")

	proxyAddr   = flag.String("x", "", "socks5 proxy addr (127.0.0.1:5005)")
	connTimeout = flag.Int("timeout", 10, "connect timeout in Seconds")

	// a list point to geojson files!!!
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

	allRows := make([]json.RawMessage, 0, 1024)
	for idx, f := range list {
		if f.IsDir() {
			continue
		}

		fname := f.Name()
		switch strings.ToLower(filepath.Ext(fname)) {
		case ".json", ".geojson":
		default:
			continue // skip not json/geojson
		}
		fp := filepath.Join(inDir, fname)
		fd, err := os.OpenFile(fp, os.O_RDONLY, 0400)
		if err != nil {
			Vln(2, "[open]err", err)
			continue
		}

		rows, err := parseData(fd)
		if err != nil {
			Vln(2, "[parse]err", idx, fname, err)
			return err
		}
		Vln(3, "[data]", idx, fname, len(rows))
		fd.Close()

		allRows = append(allRows, rows...)
	}

	err = writeJson(NewGeojson(allRows), outFp)
	if err != nil {
		return err
	}

	return nil
}

func transUrls(urls []string, dialFunc func(network, address string) (net.Conn, error), outFp string) error {
	allRows := make([]json.RawMessage, 0, 1024)
	for idx, url := range urls {
		fd, err := getUrlFd(url, dialFunc)
		if err != nil {
			Vln(2, "[get]err", url, err)
			continue
		}
		// defer fd.Close()
		Vln(3, "[get]start download...", url)

		rows, err := parseData(fd)
		if err != nil {
			Vln(2, "[parse]err", idx, url, err)
			return err
		}
		Vln(3, "[data]", idx, url, len(rows))
		fd.Close()

		allRows = append(allRows, rows...)
	}

	err := writeJson(NewGeojson(allRows), outFp)
	if err != nil {
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
	// enc.SetIndent("", "\t")
	err = enc.Encode(obj)
	if err != nil {
		Vln(2, "[json]err", err)
		return err
	}
	return nil
}

type Geojson struct {
	Type     string            `json:"type"`     // FeatureCollection
	Features []json.RawMessage `json:"features"` //json.RawMessage
}

func NewGeojson(rows []json.RawMessage) *Geojson {
	return &Geojson{
		Type:     "FeatureCollection",
		Features: rows,
	}
}

func parseData(fd io.ReadCloser) ([]json.RawMessage, error) {
	geojson := NewGeojson(nil)
	dec := json.NewDecoder(fd)
	err := dec.Decode(&geojson)
	if err != nil {
		return nil, err
	}
	return geojson.Features, nil
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
