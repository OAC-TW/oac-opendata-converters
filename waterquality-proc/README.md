## waterquality-proc

* 用途: 海域水質 爬蟲
* 資料集:
	* 名稱: 
	* 編號: 
	* 網址: https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnWaterQuality.ashx?code=D0B4699202A04DAD8A3153E48B6F937E
	* 格式: json
	* 資料集描述: 
* 語言: golang
* 輸入格式: 已有的json檔或直接取得最新的資料檔
* 輸出格式: geojson或json (依參數)
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build . # 編譯
./waterquality-proc -i "" # 執行 & 抓最新資料
```


```
go run waterquality-proc.go -i "" # 直接執行 & 抓最新資料
```

```
go run waterquality-proc.go -i 'waterquality_raw.json' # 直接執行 & 由現有檔案轉換
```

### 參數

```
  -i string
    	input json file (default "waterquality_raw.json")
  -json
    	output json format (default geojson)
  -o string
    	path to save output file (default "waterquality.json")
  -timeout int
    	connect timeout in Seconds (default 10)
  -u string
    	url (default "https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnWaterQuality.ashx?code=D0B4699202A04DAD8A3153E48B6F937E")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: 127.0.0.1:5005)
```

### sample檔案

* `sample/`
	* `waterquality_raw.json` 原始輸入檔
	* `output/` 轉換後的檔案
		* `waterquality.geojson` geojson格式
		* `waterquality.json` json格式

