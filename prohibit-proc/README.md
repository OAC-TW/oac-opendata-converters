## prohibit-proc
* 用途: 禁止海域資料合併
* 資料集:
	* 尚未上架open data
	* 格式: geojson
* 語言: golang
* 輸入格式: 多個geojson
* 輸出格式: geojson
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build prohibit-proc.go # 編譯
```


```
go run prohibit-proc.go # 直接執行 & 由 ./geojson/ 資料夾內現有檔案轉換
```


### 參數

```
  -i string
    	input path of geojson files (default "./geojson/")
  -o string
    	path to save output file (default "prohibit.json")
  -timeout int
    	connect timeout in Seconds (default 10)
  -urls string
    	url list in file (default "urls.txt")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: 127.0.0.1:5005)
```

