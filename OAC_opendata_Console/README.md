# OpenData 處理轉換程式 for 海域遊憩活動一站式資訊平臺

抓取氣象局等各方的open data, 並轉換格式提供一站式資訊平臺使用。

## 項目

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

	* 框架語言: .NET Core 3.1 (C#)
	* 轉檔輸出格式:  JSON
	* [ ] (TODO)NWW3 波浪模式

## 程式打包方式

使用套件管理員將程式打包為單一執行檔(.exe)
```bash         
PM>  dotnet publish -r win10-x64 -p:PublishSingleFile=true
```
將本程式需使用的 NetCDF 套件 *.dll 檔複製至與單一執行檔(.exe)同一目錄下

`打包後的 *.dll 檔會輸出在打包目錄的 ..\ThirdParty\NetCDF\4.6.2\x64\ 中`


## 程式參數說明  
直接執行編譯後的 .exe 檔也可檢視參數說明
* `-callEvent` (必填)處理資料集代碼 
    * `isoheXMLtoJson` : 運研所海氣象商港觀測資料
    * `ocmUVncToJson` : OPeNDAP OCM - 海流 UV 
    *  `ocmSALTncToJson` : OPeNDAP OCM - SALT 海表鹽度 
    *  `ocmSSTncToJson` : OPeNDAP OCM - SST 海表溫度 
    *  `ocmWLncToJson` : OPeNDAP OCM - WL 海面高
    *  `delFolder` : 刪除資料夾
* `-targetPath` (必填)資料輸出存放資料夾路徑 
* `-dataTime` (選填)指定要下載的 OPeNDAP OCM 資料集時間

## 程式執行範例
1. 執行下載轉檔「交通部運輸研究所 - 商港海象觀測資料」
```bash 
# 下載 XML 檔轉成 Json 檔輸出至指定資料夾 (會自動建立json資料夾存放轉檔成果)
C:\>OCA_opendata_Console.exe -callEvent=isoheXMLtoJson -targetPath="C:\OpenData\isoheStation"
        
# 若不想留存原始資料，可使用下列指令刪除下載存放的 XML 檔資料夾
C:\>OCA_opendata_Console.exe -callEvent=delFolder -targetPath="C:\OpenData\isoheStation\xml"
```
2. 執行下載轉檔「OPeNDAP OCM - 海流 UV 時序資料」
```bash 
# 下載 NetCDF 檔轉成 Json 檔輸出至指定的資料夾 (會自動建立json資料夾存放轉檔成果)
C:\>OCA_opendata_Console.exe -callEvent=ocmUVncToJson -targetPath="C:\OpenData\OcmUV"
        
# 若不想留存原始資料，可使用下列指令刪除下載存放的 nc 檔資料夾
C:\>OCA_opendata_Console.exe -callEvent=delFolder -targetPath="C:\OpenData\OcmUV\nc"
```
        
3. 執行下載轉檔「OPeNDAP OCM - 海流 UV 特定日期資料集的資料」
```bash 
C:\>OCA_opendata_Console.exe -callEvent=ocmUVncToJson -targetPath="C:\OpenDataTmp\OcmUV_Now" -dataTime=20200801
```
        
4. 執行下載轉檔「OPeNDAP OCM - 海流 UV 當時時間所在小時(hour)的資料」
```bash 
C:\>OCA_opendata_Console.exe -callEvent=ocmUVncToJson -targetPath="C:\OpenDataTmp\OcmUV_Now" -dataTime=onlyNow
```


