# OpenData處理轉換程式 for 海域遊憩活動一站式資訊平臺
抓取氣象局等各方的open data, 並轉換格式, 透過[Leaflet.js](https://leafletjs.com/)加上修改後的plugin [leaflet-velocity](https://github.com/cs8425/leaflet-velocity)直接於網頁呈現.


## 項目

* `oceancurrent-proc/`
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
	* 補充: 需要中央氣象局open data的API授權碼才可下載資料
	* 自動抓取最新資料後, 同時透過Webhook更新線上站台的資料
	* [ ] (TODO)第000~072小時參數化
	* 可藉由socks5 proxy避開網路限制

* `oceanwave-proc/`
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
	* 自動抓取最新資料並移除過時資料
	* 解壓縮/轉換時CPU核心可能會吃滿3核(可由指令參數調整)

## demo

* 啟動簡易的web server, 將根目錄指向本專案
	* 例如: `http://127.0.0.1:8080/`
* 開啟瀏覽器, 連至web server, 檢視各項目目錄底下的`demo.html`
	* 例如: `http://127.0.0.1:8080/oceancurrent-proc/demo.html`

## TODO

* [ ]將一些共通的結構/function整合成一份, 用import的方式引入


