## sightseeingspot-proc

* 用途: 海洋遊憩景點 爬蟲
* 資料集:
	* 基隆市政府: https://data.gov.tw/dataset/143682
	* 苗栗縣政府: https://data.gov.tw/dataset/144515
	* 金門縣政府:
		* https://data.gov.tw/dataset/143630
		* https://data.gov.tw/dataset/143629
		* https://data.gov.tw/dataset/143628
		* https://data.gov.tw/dataset/143627 
		* https://data.gov.tw/dataset/143626
	* 彰化縣政府: 
	* 格式: xlsx (待改善), ods (待改善), csv (big5, 待改善), csv (utf8)
* 語言: golang
* 輸入格式: 多個csv
* 輸出格式: geojson或json (依參數)
* 可藉由socks5 proxy避開網路限制
* 注意:
	* 目前資料來源的格式無法自動轉換, 須手動轉檔成csv後再使用


### 編譯/執行

```
go build sightseeingspot-proc.go # 編譯
```

```
go run sightseeingspot-proc.go # 直接執行 & 由 ".csv" 資料夾轉資料
```

```
go run sightseeingspot-proc.go -i="" -urls="urls.txt" # 由 urls.txt 裡的網址下載抓最新資料
```

### 參數

```
  -i string
    	input path of CSV files (default "./csv/")
  -json
    	output json format (default json)
  -o string
    	path to save output file (default "sightseeingspot.json")
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
