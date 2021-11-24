## sandybeach-proc

* 用途: 海灘水質 爬蟲
* 資料集:
	* 名稱: 
	* 編號: 
	* 網址: https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnSandyBeach.ashx?code=F36490AE812D4E4791C14975551DD2E9
	* 格式: json
	* 資料集描述: 
* 語言: golang
* 輸入格式: 已有的json檔或直接取得最新的資料檔
* 輸出格式: geojson或json (依參數)
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build . # 編譯
./sandybeach-proc -i "" # 執行 & 抓最新資料
```


```
go run sandybeach-proc.go -i "" # 直接執行 & 抓最新資料
```

```
go run sandybeach-proc.go -i 'sandybeach_raw.json' # 直接執行 & 由現有檔案轉換
```

### 參數

```
  -i string
    	input json file (default "sandybeach_raw.json")
  -json
    	output json format (default geojson)
  -o string
    	path to save output file (default "sandybeach.json")
  -timeout int
    	connect timeout in Seconds (default 10)
  -u string
    	url (default "https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnSandyBeach.ashx?code=F36490AE812D4E4791C14975551DD2E9")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: 127.0.0.1:5005)
```

### sample檔案

* `sample/`
	* `sandybeach_raw.json` 原始輸入檔
	* `output/` 轉換後的檔案
		* `sandybeach.geojson` geojson格式
		* `sandybeach.json` json格式


