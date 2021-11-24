## shootingobstruct-proc

* 用途: 射擊/礙航資料轉換&合併
* 資料集:
	* 格式: csv
	* 射擊通報:
		* **注意: 已上架之open data含大量格式錯誤, 暫無法直接使用**
		* 國防部軍備局規格鑑測中心兵器試驗場實彈射擊公告: https://data.gov.tw/dataset/136905
		* 空軍司令部射擊公告: https://data.gov.tw/dataset/138550
		* 國防部海軍司令部-實彈射擊公告: https://data.gov.tw/dataset/136419
		* 陸軍射擊公告: https://data.gov.tw/dataset/137325
	* 礙航通報:
		* 暫無上架open data
* 語言: golang
* 輸入格式: 多個csv
* 輸出格式: geojson
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build shootingobstruct-proc.go # 編譯
```

```
go run shootingobstruct-proc.go # 直接執行 & 由 ".csv" 資料夾轉資料
```

```
go run shootingobstruct-proc.go -i="" -urls="urls.txt" # 由 urls.txt 裡的網址下載抓最新資料
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
