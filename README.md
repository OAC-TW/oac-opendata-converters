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
	
* `OAC_opendata_Console/`
	* 用途: 提供將下列 OpenData 轉換為 一站式平臺使用之資料格式
		* ######  交通部運輸研究所 - 商港海象觀測資料
            * 資料集描述: 商港海氣象資訊（風力、潮位、波浪、海流） 風力：觀測時間、平均風速、平均風向、緯度、經度 潮位：觀測時間、潮位、緯度、經度 波浪：觀測時間、波高、尖峰週期、波向、平均週期、緯度、經度 海流：觀測時間、流速、流向、緯度、經度
            * 資料格式: XML
            * 資料集網址
                * 臺北商港：https://data.gov.tw/dataset/127836
                * 基隆商港：https://data.gov.tw/dataset/127851
                * 蘇澳商港：https://data.gov.tw/dataset/127855
                * 臺中商港：https://data.gov.tw/dataset/127831
                * 布袋商港：https://data.gov.tw/dataset/127840
                * 安平商港：https://data.gov.tw/dataset/127846
                * 高雄商港：https://data.gov.tw/dataset/127853
                * 花蓮商港：https://data.gov.tw/dataset/127852
                * 馬祖(南竿): https://data.gov.tw/dataset/127847

    	* ######  中央氣象局  OCM 海流模式資料
            * 資料集描述: 提供 OCM 預報模式的海流資訊 (海流、海表鹽度、海表溫度、海面高)。
            * 開放資料來源網址 https://ocean.cwb.gov.tw/V2/data_interface/datasets
    		* 資料格式: NetCDF
            * OPeNDAP OCM 資料集網址: http://med.cwb.gov.tw/opendap/OCM/contents.html

    	* ######  中央氣象局  OPENDATA 資料
            * 資料集描述: 颱風消息與警報-災害性天氣資訊-颱風警報CAP檔。
            * 開放資料來源網址 https://opendata.cwb.gov.tw/dataset/warning/W-C0034-001
            * 資料集描述: 颱風消息與警報-颱風消息KML壓縮檔。
            * 開放資料來源網址 https://opendata.cwb.gov.tw/dataset/warning/W-C0034-002
            * 資料集描述: 潮汐預報-(未來1個月潮汐預報，鄉鎮、大潮小潮、滿潮乾潮、時間、潮高)。
            * 開放資料來源網址 https://opendata.cwb.gov.tw/dataset/forecast/F-A0021-001
    		* 資料集描述: 臺灣各鄉鎮市區預報資料-鄉鎮天氣預報。
            * 開放資料來源網址 https://opendata.cwb.gov.tw/dataset/forecast/F-D0047-095 
            * 資料集描述: 臺灣鄰近及遠洋漁業海面天氣預報資料-海面天氣預報(中文版)。
            * 開放資料來源網址 https://opendata.cwb.gov.tw/dataset/forecast/F-A0012-001

	* 框架語言: .NET Core 3.1 (C#)
	* 轉檔輸出格式:  JSON
	* [ ] (TODO)NWW3 波浪模式

* `radiation-proc/`
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
	* TODO:
		* [x] 只輸出最近N天內的資料
		* [ ] 只輸出最新的N筆

* `sandybeach-proc/`
	* 用途: 海灘水質 爬蟲
	* 資料集:
		* 名稱: ????
		* 編號: ????
		* 網址: https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnSandyBeach.ashx?code=F36490AE812D4E4791C14975551DD2E9
		* 格式: json
		* 資料集描述: 
	* 語言: golang
	* 輸入格式: 已有的json檔或直接取得最新的資料檔
	* 輸出格式: geojson或json (依參數)
	* 可藉由socks5 proxy避開網路限制

* `waterquality-proc/`
	* 用途: 海域水質 爬蟲
	* 資料集:
		* 名稱: ????
		* 編號: ????
		* 網址: https://iocean.oca.gov.tw/OCA_OceanConservation/Service/GeneratorFromJosnWaterQuality.ashx?code=D0B4699202A04DAD8A3153E48B6F937E
		* 格式: json
		* 資料集描述: 
	* 語言: golang
	* 輸入格式: 已有的json檔或直接取得最新的資料檔
	* 輸出格式: geojson或json (依參數)
	* 可藉由socks5 proxy避開網路限制

* `seasurfacestorm-proc/`
	* 用途: 海面風波流 爬蟲
	* 資料集:
		* 名稱: ????
		* 編號: ????
		* 網址:
			* https://goocean.namr.gov.tw/OAC/?sta=O&aw=1&key=OAC&ff=txt
			* https://goocean.namr.gov.tw/OAC/?sta=B&aw=1&key=OAC&ff=txt
			* https://goocean.namr.gov.tw/OAC/?sta=C&aw=1&key=OAC&ff=txt
			* https://goocean.namr.gov.tw/OAC/?sta=R&aw=1&key=OAC&ff=txt
			* https://goocean.namr.gov.tw/OAC/?sta=Y&aw=1&key=OAC&ff=txt
		* 格式: txt
		* 資料集描述: 
	* 語言: golang
	* 輸入格式: 直接取得最新的資料檔
	* 輸出格式: geojson或json (依參數)
	* 可藉由socks5 proxy避開網路限制
	* 注意:
		* **來源資料編碼為big5**
		* 剛跨月可能沒有任何資料
		* 因資料來源的特性, 會保存最後一筆資料

* `sightseeingspot-proc/`
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

* `prohibit-proc/`
	* 用途: 禁止海域資料合併
	* 資料集:
		* 尚未上架open data
		* 格式: geojson
	* 語言: golang
	* 輸入格式: 多個geojson
	* 輸出格式: geojson
	* 可藉由socks5 proxy避開網路限制

* `shootingobstruct-proc/`
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

## demo

* 啟動簡易的web server, 將根目錄指向本專案
	* 例如: `http://127.0.0.1:8080/`
* 開啟瀏覽器, 連至web server, 檢視各項目目錄底下的`demo.html`
	* 例如: `http://127.0.0.1:8080/oceancurrent-proc/demo.html`

## TODO

* [ ]將一些共通的結構/function整合成一份, 用import的方式引入


