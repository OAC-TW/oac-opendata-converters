## radiation-proc

* 用途: 海域輻射 爬蟲
* 資料集:
	* 名稱: 台灣海域輻射調查資料
	* 編號: 136060
	* 網址: https://data.gov.tw/dataset/136060
	* 格式: csv
	* 資料集描述: 取樣日期、經度、緯度、深度(M)、銫-137活度、銫-137單位、氚活度、氚單位、類別、地點編號、地點、取樣單位/人、備註
* 語言: golang
* 輸入格式: 已有的csv檔或直接取得最新的資料檔
* 輸出格式: geojson或json (依參數)
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build . # 編譯
./radiation-proc -i "" # 執行 & 抓最新資料
```


```
go run radiation-proc.go -i "" # 直接執行 & 抓最新資料
```

```
go run radiation-proc.go -i 'radiation_raw.csv' # 直接執行 & 由現有檔案轉換
```

### 參數

```
  -i string
    	input CSV file (default "radiation_raw.csv")
  -json
    	output json format (default geojson)
  -o string
    	path to save output file (default "radiation.json")
  -keep int
    	keep latest N days data (default 180)
  -timeout int
    	connect timeout in Seconds (default 10)
  -u string
    	url (default "https://www.aec.gov.tw/share/file/information/gdhyTsQBfnL9N3h7Xf2iYg__.csv")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: 127.0.0.1:5005)
```

### sample檔案

* `sample/`
	* `radiation_raw.csv` 原始輸入檔
	* `json/` 轉換後的檔案
		* `radiation.geojson` geojson格式
		* `radiation.json` json格式


