## seasurfacestorm-proc

* 用途: 海面風波流 爬蟲
* 資料集:
	* 名稱: 
	* 編號: 
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


### 編譯/執行

```
go build . # 編譯
./seasurfacestorm-proc # 執行 & 抓最新資料
```


```
go run seasurfacestorm-proc.go # 直接執行 & 抓最新資料
```

### 參數

```
  -i string
    	input txt file (default "seasurfacestorm_raw.txt")
  -json
    	output json format (default geojson)
  -o string
    	path to save output file (default "seasurfacestorm.json")
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
	* `seasurfacestorm_raw.txt` 原始輸入檔
	* `output/` 轉換後的檔案
		* `seasurfacestorm.geojson` geojson格式
		* `seasurfacestorm.json` json格式

