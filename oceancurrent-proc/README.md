## oceancurrent-proc

* 用途: 中央氣象局 橫向流速、直向流速、流速、流向、海表溫度、海高、海表鹽度
* 資料集:
	* 名稱: 海流模式-海流數值模式預報資料-第000小時
	* 編號: M-B0071-000
	* 網址: https://opendata.cwb.gov.tw/dataset/climate/M-B0071-000
	* 格式: XML
	* 資料集描述: 海流數值模式預報資料-提供本局海流數值預報模式表層資料，包含分析場(00Z)及72小時逐時預報，範圍為東經110~126度、北緯7~36度，解析度為0.1*0.1度
* 語言: golang
* 輸入格式: 已有的XML檔或直接取得最新的XML檔
* 輸出格式: json
* 補充: **需要中央氣象局open data的API授權碼才可下載資料**
* 自動抓取最新資料後, 同時透過Webhook更新線上站台的資料
* [ ] (TODO)第000~072小時參數化
* 可藉由socks5 proxy避開網路限制


### 編譯/執行

```
go build . # 編譯
./oceancurrent-proc -auth 'CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' # 執行 & 抓最新資料
```


```
go run oceancurrent-proc.go -auth 'CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX' # 直接執行 & 抓最新資料
```

```
go run oceancurrent-proc.go -auth '' -i 'M-B0071-000.20200812-1530.xml' -o 'M-B0071-000.20200812-1530.grid.json' # 直接執行 & 由現有檔案轉換
```

### 參數

```
  -auth string
    	氣象局token (default "CWB-XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX")
  -hook string
    	web hook URL (例: "http://127.0.0.1:8080/api/push/89HuRzqCRlRGIrhSifYN")
  -i string
    	input XML file (default "M-B0071-000.xml")
  -o string
    	output file (default "M-B0071-000.grid.json")
  -timeout int
    	connect timeout in Seconds (default 10)
  -u string
    	資料集下載url (default "https://opendata.cwb.gov.tw/fileapi/v1/opendataapi/M-B0071-000?Authorization=%v&downloadType=WEB&format=XML")
  -ua string
    	User-Agent (default "OAC bot")
  -v int
    	verbosity for app (default 3)
  -x string
    	socks5 proxy addr (例: "127.0.0.1:5005")
```

### sample檔案

* `sample/`
	* `M-B0071-000.20200812-1530.xml.zip` zip壓縮後的原始輸入檔, 請解壓縮後再餵入轉換程式
	* `M-B0071-000.20200812-1530.grid.json` 轉換後的檔案


