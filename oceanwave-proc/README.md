## oceanwave-proc

* 用途: 中央氣象局 浪高(hs)、週期(t)、波向(dir) 爬蟲
* 資料集:
	* 名稱: 波浪預報模式資料-臺灣海域預報資料
	* 編號: F-A0020-001
	* 網址: https://opendata.cwb.gov.tw/dataset/climate/F-A0020-001
	* 格式: ZIP (複數xml檔打包)
	* 資料集描述: 臺灣海域波浪預報逐三小時數值模式資料-包含浪高(hs)、週期(t)、波向(dir)
* 語言: golang
* 輸入格式: 已有的ZIP檔或直接取得最新的資料檔
* 輸出格式: 數個json, 包括一個index.json
* 補充: 需要中央氣象局open data的API授權碼才可下載資料
* 可藉由socks5 proxy避開網路限制
* 自動抓取最新資料, 並移除輸出資料夾內過時的資料
* 解壓縮/轉換時CPU核心可能會吃滿3核(可由指令參數調整)


### 編譯/執行

```
go build . # 編譯
./oceanwave-proc -auth 'CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' # 執行 & 抓最新資料
```


```
go run oceanwave-proc.go -auth 'CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' # 直接執行 & 抓最新資料
```

```
go run oceanwave-proc.go -auth '' -i 'F-A0020-001-20200618-1420.zip' # 直接執行 & 由現有檔案轉換
```

### 參數

```
  -auth string
    	氣象局token (default "CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
  -cpu int
    	CPU count limit, 0 == auto
  -dir string
    	path to save output file (default "json/")
  -i string
    	input XML in zip file (default "F-A0020-001.zip")
  -timeout int
    	connect timeout in Seconds (default 10)
  -u string
    	資料集下載url (default "https://opendata.cwb.gov.tw/fileapi/v1/opendataapi/F-A0020-001?Authorization=%v&downloadType=WEB&format=ZIP")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: 127.0.0.1:5005)

```

### sample檔案

* `sample/`
	* `F-A0020-001-20200618-1420.zip` 原始輸入檔
	* `json/` 轉換後的檔案
		* `index.json` 索引檔, 提供時間(UTC+0跟UTC+8)跟檔名
		* `[0-9]{8}.[0-9]{3}.grid.json` 輸出檔案


