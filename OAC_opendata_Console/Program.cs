using NetCDF;
using Newtonsoft.Json;
using OAC_opendata_Console.Model;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using OAC_opendata_Console.Libraries.RWLib;
using System.Xml;
using System.Text.RegularExpressions;

namespace OAC_opendata_Console
{
    class Program
    { 

        static void Main(string[] args)
        {
            var param = ParseString(Environment.CommandLine);
            if (param.Count == 0)
            {
                Console.WriteLine("================================");
                Console.WriteLine("請搭配輸入指令使用本程式");
                Console.WriteLine("--------------------------------");
                Console.WriteLine("指令說明: \n");
                Console.WriteLine("--------------------------------");
                Console.WriteLine(" -callEvent (必填)處理資料集代碼 \n");
                Console.WriteLine("--------------------------------");
                Console.WriteLine(" Ex.  -callEvent = isoheXMLtoJson \n");
                Console.WriteLine("  isoheXMLtoJson   : 運研所海氣象資料下載轉檔 - 商港觀測資料 ");
                Console.WriteLine("  ocmUVncToJson    : OPeNDAP OCM 資料下載轉檔 - 海流 UV ");
                Console.WriteLine("  ocmSALTncToJson  : OPeNDAP OCM 資料下載轉檔 - SALT 海表鹽度 ");
                Console.WriteLine("  ocmSSTncToJson   : OPeNDAP OCM 資料下載轉檔 - SST 海表溫度 ");
                Console.WriteLine("  ocmWLncToJson    : OPeNDAP OCM 資料下載轉檔 - WL 海面高 ");
                Console.WriteLine("  delFolder        : 刪除資料夾 (提供刪除下載資料原始檔存放的資料夾)");
                Console.WriteLine(" ");
                Console.WriteLine("--------------------------------");
                Console.WriteLine(" -targetPath (必填)資料輸出存放資料夾路徑 \n");
                Console.WriteLine("--------------------------------");
                Console.WriteLine(" Ex.  -targetPath = \"C:\\OpenDataTrans\\isoheStation\" \n");
                Console.WriteLine(" ");
                Console.WriteLine("--------------------------------");
                Console.WriteLine(" -dataTime (選填)指定要下載的 OPeNDAP OCM 資料集時間(不指定時會自動判斷是抓今日或昨日有資料的源) \n");
                Console.WriteLine("--------------------------------");
                Console.WriteLine($" Ex.  -dataTime = {DateTime.Now.ToString("yyyyMMdd")}         (指定抓 OPeNDAP 上 YYYYMMDD 的資料源) \n");
                Console.WriteLine($" Ex.  -dataTime = onlyNow          (只會抓取當小時最新的 1 筆資料)  ");
                Console.WriteLine(" ");
                Console.WriteLine("================================");

                //不關閉視窗讓使用者看說明
                Console.ReadKey();

            }
            else
            {
                Console.WriteLine("================================");
                Console.WriteLine("您目前輸入的參數");
                Console.WriteLine("--------------------------------");
                foreach (KeyValuePair<string, string> kv in param)
                {
                    Console.WriteLine("Key={0}  Value={1}",kv.Key, kv.Value);
                }
                Console.WriteLine("--------------------------------");

                // 必填參數
                if (param.ContainsKey("callEvent") && param.ContainsKey("targetPath"))
                {
                    Console.WriteLine("程式執行開始");

                    //自訂處理工具
                    RWLib_NetCDF rwLibNetcdf = new RWLib_NetCDF();
                    RWLib_Date rwLibDate = new RWLib_Date();
                    RWLib_Net rwLibNet = new RWLib_Net();
                    RWLib_FileIO rwLibFio = new RWLib_FileIO();


                    // 指令 - 作業種類代碼
                    string callEvent = param["callEvent"];
                    // 指令 - 作業目錄位置
                    string argFileOrOutFolderPath = param["targetPath"];
                    Console.WriteLine(callEvent);
                    Console.WriteLine(argFileOrOutFolderPath);

                    #region  運研所海氣象資料下載轉檔 - 商港觀測資料

                    if (callEvent.Equals("isoheXMLtoJson"))
                    {

                        RWLib_Log rwLibisoheLog = new RWLib_Log(@$"{argFileOrOutFolderPath}\log\", "isoheStations");

                        //https://isohe.ihmt.gov.tw/
                        rwLibisoheLog.log("運研所海氣象資料轉檔 --- Begin");

                        // XML 下載存放路徑
                        string isoheXMLsavePath = @$"{argFileOrOutFolderPath}\xml\";
                        rwLibFio.DelAndCreateFolder(isoheXMLsavePath);
                        // 產出 JSON 檔案路徑
                        string outJsonFolderPath = @$"{argFileOrOutFolderPath}\json\";
                        string outJsonFilePath = @$"{outJsonFolderPath}isoheStations.json";

                        // 商港觀測站 OpenData (裡面有很多觀測站沒有 OpenData 所以自己建要爬的資料集)
                        // https://isohe.ihmt.gov.tw/station/OpenData/XML/GetSAstationXML.aspx
                        // 有 OpenData 的資料集  (代碼,港名,opendata資料集編號)
                        string isoheStations = "KL,基隆港,127851|SA,蘇澳港,127855|HL,花蓮港,127852|TC,臺中港,127831|KH,高雄港,127853|TP,臺北港,127836|BD,布袋港,127840|AP,安平港,127846|NG,馬祖南竿,127847";
                        //isoheStations = "KL,基隆港,127851"; // 測試只單抓一個商港

                        // 準備下載資訊的資料集
                        List<isoheStation> isoheStationList = new List<isoheStation>();
                        foreach (string sta in isoheStations.Split("|"))
                        {
                            isoheStation staObj = new isoheStation();
                            staObj.Code = sta.Split(",")[0];
                            staObj.Name = sta.Split(",")[1];
                            staObj.XmlSourceUrl = $"https://isohe.ihmt.gov.tw/station/OpenData/XML/Get{sta.Split(",")[0]}stationXML.aspx";
                            staObj.OpendataLinkUrl = $"https://data.gov.tw/dataset/{sta.Split(",")[2]}";
                            isoheStationList.Add(staObj);
                        }

                        List<Station> _stationList = new List<Station>();
                        foreach (isoheStation sta in isoheStationList)
                        {
                            rwLibisoheLog.log($">>>> " + sta.XmlSourceUrl);
                            string xmlFilePath = $"{isoheXMLsavePath}{sta.Code}_Station.xml";
                            //下載資料
                            var success = rwLibNet.DownloadFile(sta.XmlSourceUrl, xmlFilePath);
                            rwLibisoheLog.log($"{sta.Code}_Station  下載狀態 " + success);
                            if (success)
                            {
                                //讀取 XML
                                XmlDocument doc = new XmlDocument();
                                doc.Load(xmlFilePath);

                                List<HistoryDataItem> hisDataList = new List<HistoryDataItem>();
                                List<TideDataItem> tideDataList = new List<TideDataItem>();
                                List<WindDataItem> windDataList = new List<WindDataItem>();
                                StationLocation hisStLoc = new StationLocation();
                                StationLocation tideStLoc = new StationLocation();
                                StationLocation windStLoc = new StationLocation();
                                // 塞入資料到物件
                                foreach (XmlNode xn in doc.ChildNodes.Item(1).ChildNodes)
                                {
                                    switch (xn.Name)
                                    {
                                        case "History":
                                            HistoryDataItem _hisItem = new HistoryDataItem();
                                            double _hisLat = xn.SelectSingleNode("Latitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Latitude").InnerText);
                                            double _hisLng = xn.SelectSingleNode("Longitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Longitude").InnerText);
                                            hisStLoc.Latitude = (hisStLoc.Latitude == 0.0) ? _hisLat : ((_hisLat + hisStLoc.Latitude) / 2);
                                            hisStLoc.Longitude = (hisStLoc.Longitude == 0.0) ? _hisLng : ((_hisLng + hisStLoc.Longitude) / 2);

                                            //這邊的節點項目有可能會不存在！！ 線上資料集沒說明值的內容若是 -999.99 是什麼意思
                                            _hisItem.Date_Time = xn.SelectSingleNode("Date_Time") == null ? "" : xn.SelectSingleNode("Date_Time").InnerText;
                                            _hisItem.HS = xn.SelectSingleNode("HS") == null ? 0.0 : double.Parse(xn.SelectSingleNode("HS").InnerText);
                                            _hisItem.TP = xn.SelectSingleNode("TP") == null ? 0.0 : double.Parse(xn.SelectSingleNode("TP").InnerText);
                                            _hisItem.MDIR = xn.SelectSingleNode("MDIR") == null ? 0.0 : double.Parse(xn.SelectSingleNode("MDIR").InnerText);
                                            _hisItem.Tmean = xn.SelectSingleNode("Tmean") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Tmean").InnerText);
                                            _hisItem.Velocity = xn.SelectSingleNode("Velocity") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Velocity").InnerText);
                                            _hisItem.Vmdir = xn.SelectSingleNode("Vmdir") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Vmdir").InnerText);

                                            hisDataList.Add(_hisItem);
                                            break;
                                        case "WindData":
                                            WindDataItem _windItem = new WindDataItem();

                                            double _windLat = xn.SelectSingleNode("Latitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Latitude").InnerText);
                                            double _windLng = xn.SelectSingleNode("Longitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Longitude").InnerText);
                                            windStLoc.Latitude = (windStLoc.Latitude == 0.0) ? _windLat : ((_windLat + windStLoc.Latitude) / 2);
                                            windStLoc.Longitude = (windStLoc.Longitude == 0.0) ? _windLng : ((_windLng + windStLoc.Longitude) / 2);

                                            _windItem.Date_Time = xn.SelectSingleNode("Date_Time") == null ? "" : xn.SelectSingleNode("Date_Time").InnerText;
                                            _windItem.WD_AVG = xn.SelectSingleNode("WD_AVG") == null ? 0.0 : double.Parse(xn.SelectSingleNode("WD_AVG").InnerText);
                                            _windItem.WS_AVG = xn.SelectSingleNode("WS_AVG") == null ? 0.0 : double.Parse(xn.SelectSingleNode("WS_AVG").InnerText);

                                            windDataList.Add(_windItem);
                                            break;
                                        case "TideData":
                                            TideDataItem _tideItem = new TideDataItem();

                                            double _tideLat = xn.SelectSingleNode("Latitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Latitude").InnerText);
                                            double _tideLng = xn.SelectSingleNode("Longitude") == null ? 0.0 : double.Parse(xn.SelectSingleNode("Longitude").InnerText);

                                            tideStLoc.Latitude = (tideStLoc.Latitude == 0.0) ? _tideLat : ((_tideLat + tideStLoc.Latitude) / 2);
                                            tideStLoc.Longitude = (tideStLoc.Longitude == 0.0) ? _tideLng : ((_tideLng + tideStLoc.Longitude) / 2);

                                            _tideItem.Date_Time = xn.SelectSingleNode("Date_Time") == null ? "" : xn.SelectSingleNode("Date_Time").InnerText;
                                            _tideItem.TideValue = xn.SelectSingleNode("TideValue") == null ? 0.0 : double.Parse(xn.SelectSingleNode("TideValue").InnerText);

                                            tideDataList.Add(_tideItem);
                                            break;

                                    }
                                }

                                Station _station = new Station()
                                {
                                    Name = sta.Name,
                                    Code = sta.Code,
                                    OpendataLinkUrl = sta.OpendataLinkUrl,
                                    XmlSourceUrl = sta.XmlSourceUrl,
                                    DataSet = new StationDataSet()
                                    {
                                        HistoryData = new HistoryData()
                                        {
                                            location = hisStLoc,
                                            Data = hisDataList
                                        }
                                        ,
                                        TideData = new TideData()
                                        {
                                            location = tideStLoc,
                                            Data = tideDataList
                                        }
                                        ,
                                        WindData = new WindData()
                                        {
                                            location = windStLoc,
                                            Data = windDataList
                                        }
                                    }
                                };
                                _stationList.Add(_station);

                            }

                        }
                        // 組合各港觀測資料
                        StationList stationList = new StationList()
                        {
                            FieldsDesc = new FieldsDesc()
                            {
                                StationFieldsDesc = new StationFieldsDesc(),
                                HistoryDataFieldsDesc = new HistoryDataFieldsDesc(),
                                TideDataFieldsDesc = new TideDataFieldsDesc(),
                                WindDataFieldsDesc = new WindDataFieldsDesc()
                            }
                            ,
                            Stations = _stationList
                        };
                        // 判斷成功轉出的測試數量與組合的數量相同，則轉出 JSON 檔
                        if (_stationList.Count == isoheStationList.Count)
                        {
                            rwLibFio.DelAndCreateFolder(outJsonFolderPath);

                            string json = JsonConvert.SerializeObject(stationList);
                            rwLibisoheLog.log($"產出 JSON 檔 {outJsonFilePath}");
                            System.IO.File.WriteAllText(outJsonFilePath, json);
                        }
                        else
                        {
                            rwLibisoheLog.log($"XML 檔下載數量 [{_stationList.Count}] 不同於測站數量 [{ isoheStationList.Count}]， 不進行 JSON 檔更新");
                        }

                        rwLibisoheLog.log("運研所海氣象資料轉檔 --- End ");
                    }

                    #endregion


                    #region OPeNDAP OCM 資料下載轉檔 - 海流 UV、海表鹽度、海表溫度、海面高 

                    if (callEvent.Equals("ocmSALTncToJson")
                        || callEvent.Equals("ocmSSTncToJson")
                        || callEvent.Equals("ocmWLncToJson")
                        || callEvent.Equals("ocmUVncToJson")
                         || callEvent.Equals("ocmNcToJson")
                        )
                    {


                        #region 讀取參數設定 OCM 要抓取資料的自訂日期
                        // OPeNDAP OCM 資料來源網址
                        string cwbOpenDapOcmUrl = $"http://med.cwb.gov.tw/opendap/hyrax/OCM/";

                        // Ocm 資料是否只抓當下最新的 1 筆
                        bool onlyTransNowData = false;
                        // Ocm 資料日期路徑 
                        string argOcmDataObtainDate = DateTime.Now.ToString("yyyyMMdd");
                        if (!rwLibNet.ChekUrlRequestStatus($"{cwbOpenDapOcmUrl}{argOcmDataObtainDate}/"))
                        {
                            // 自動取得目前 OCM 的資料是今日或昨日
                            argOcmDataObtainDate = DateTime.Now.AddDays(-1).ToString("yyyyMMdd");
                            Console.WriteLine($" OPeNDAP OCM 資料最新日期為 {argOcmDataObtainDate} \n");
                        }
                        // 指定下載資料的日期
                        if (param.ContainsKey("dataTime"))
                        {
                            string parm_dataTime = param["dataTime"];
                            parm_dataTime = parm_dataTime.Split('"')[0].Trim();
                            
                            if (parm_dataTime.IndexOf("onlyNow") != -1)
                                onlyTransNowData = true;
                            else if(parm_dataTime.Length == 8)
                                argOcmDataObtainDate = parm_dataTime;
                        }
                        #endregion



                        string ncTypeCode = "";
                        switch (callEvent)
                        {
                            case "ocmSALTncToJson":
                                ncTypeCode = "SALT";
                                break;
                            case "ocmSSTncToJson":
                                ncTypeCode = "SST";
                                break;
                            case "ocmWLncToJson":
                                ncTypeCode = "WL";
                                break;
                            case "ocmUVncToJson"://OPeNDAP OCM 資料下載轉檔 - 海流 UV
                                ncTypeCode = "UCURR,VCURR";
                                break;
                        }

                        RWLib_Log rwLibOcmLog = new RWLib_Log($@"{argFileOrOutFolderPath}\log\", ncTypeCode);
                        rwLibOcmLog.log($"=========================================================================");
                        rwLibOcmLog.log($" OPeNDAP OCM 資料下載轉檔 ---- Begin \n");

                        if (!ncTypeCode.Equals(""))
                        {
                            try
                            {

                                //下載時間資料集數量 (目前 OPeNDAP OCM 資料集完整時間為 120)
                                int downloadNcTimeItemCount = 120;
                                //下載資料 存放路徑
                                string progFileSaveFolderPath = $@"{argFileOrOutFolderPath}\";
                                string nc_FolderPath = $@"{progFileSaveFolderPath}nc\";
                                string json_Temp_FolderPath = $@"{progFileSaveFolderPath}jsonTemp\";
                                string json_FolderPath = $@"{progFileSaveFolderPath}json\";
                                // 下載失敗重新嘗試次數
                                int reTryDownloadTimes = 5;

                                // 清空下載及暫存的資料夾
                                rwLibFio.DelAndCreateFolder(nc_FolderPath);
                                rwLibFio.DelAndCreateFolder(json_Temp_FolderPath);


                                // (1) 建立一 ocmIndexList 記錄 nc 下載成功數量
                                // (2) 跑各類別 ncTypeCode 的 nc 檔下載作業，將結果記錄於上述 List
                                // (3) 檢查 List 中各類別成功下載 nc 檔的數量是否大於 0 且能被 類別 ncTypeCode 數量整除
                                // (4) 進行 json 轉檔
                                // (5) 整理 index 資料，轉出 index json 檔

                                string[] _ncTypeCodes = ncTypeCode.Split(",");
                                int ncTypeCount = 0;
                                List<OcmDataIndex> ocmIndexList = new List<OcmDataIndex>();
                                foreach (string _ncTypeCode in _ncTypeCodes)
                                {
                                    rwLibOcmLog.log($"處理資料集 {_ncTypeCode} ； 整體進度 {ncTypeCount + 1}/{_ncTypeCodes.Length} \n");

                                    string ncTimeFilePath = @$"{nc_FolderPath}{_ncTypeCode}times.nc";

                                    //先確認設定日期的時間檔是否存在，有存在才進行檔案的下載作業
                                    rwLibOcmLog.log($"時間集 nc 檔連線狀態檢查 \n");
                                    string ocm_TimeNcUrl = $"{cwbOpenDapOcmUrl}{argOcmDataObtainDate}/00/9999/{_ncTypeCode}.{argOcmDataObtainDate}00.nc.nc?time[0:1:119]";
                                    rwLibOcmLog.log($"{ocm_TimeNcUrl}\n");

                                    if (rwLibNet.ChekUrlRequestStatus(ocm_TimeNcUrl))
                                    {
                                        rwLibOcmLog.log($"時間集 nc 檔連線狀態檢查 OK!! \n");
                                        rwLibOcmLog.log($"> 開始下載 {ocm_TimeNcUrl} 到 {ncTimeFilePath}");
                                        var successTimeNc = rwLibNet.DownloadFile(ocm_TimeNcUrl, ncTimeFilePath);
                                        if (successTimeNc)
                                        {
                                            rwLibOcmLog.log($">> 時間集 nc 檔下載成功 \n");

                                            // 解析 times 檔中的時段
                                            NcFile ncTimeFile = new NcFile(ncTimeFilePath);
                                            ncTimeFile.Variables[0].ReadData();
                                            int[] _times = ((NcVarTyped<int>)ncTimeFile.Variables[0]).Data;
                                            rwLibOcmLog.log($"取得時間資料集數量  [{_times.ToList().Count}]");

                                            if (_times.ToList().Count == 120)
                                            {
                                                rwLibOcmLog.log($"時間資料集數量檢核正確，開始進行 nc 檔下載 \n");
                                                int i = 0;
                                                int ncDataCount = 0;// nc檔下載數量
                                                int historyDataCount = 0;//歷史資料集數量
                                                foreach (var time in _times.ToList())
                                                {
                                                    // 建立單一時間集的 nc資料索引記錄物件檔
                                                    OcmDataIndex ocmIndex = new OcmDataIndex();
                                                    ocmIndex.Index = i;

                                                    // 取得 UTC 轉換後的時間
                                                    ocmIndex.Time = time;
                                                    DateUtc_08 dtUtc08 = rwLibDate.GetUtcAnd08DateTimeFromHours(time);
                                                    ocmIndex.time08 = dtUtc08.time08;
                                                    ocmIndex.timeUTC = dtUtc08.timeUTC;

                                                    string dataDTimeHourName = dtUtc08.dt08.ToString("yyyyMMddHH");

                                                    // 只轉目前日期小時以後的預報資料
                                                    if (DateTime.Compare(dtUtc08.dt08, DateTime.Now.AddHours(-1)) > 0)
                                                    {
                                                        //  nc 資料集下載網址
                                                        ocmIndex.ncDataUrl = $"{cwbOpenDapOcmUrl}{argOcmDataObtainDate}/00/9999/{_ncTypeCode}.{argOcmDataObtainDate}00.nc.nc?{_ncTypeCode}[{i}:1:{i}][0:1:0][0:1:1160][0:1:640]";
                                                        //  nc 資料存放位址
                                                        ocmIndex.ncFilePath = $"{nc_FolderPath}{i}_{_ncTypeCode}_{dataDTimeHourName}.nc";
                                                        //  json file name
                                                        ocmIndex.name = $"{dataDTimeHourName}_{_ncTypeCode}.json";

                                                        // 下載 nc 檔
                                                        bool successNCdata = rwLibNet.DownloadFile(ocmIndex.ncDataUrl, ocmIndex.ncFilePath);
                                                        if (successNCdata)
                                                        {
                                                            rwLibOcmLog.log($">> {i}_{_ncTypeCode} 資料集 nc 檔 下載成功\n");
                                                            ocmIndexList.Add(ocmIndex);
                                                            ncDataCount++;
                                                        }
                                                        else
                                                        {
                                                            rwLibOcmLog.log($">> {i}_{_ncTypeCode} 資料集 nc 檔 下載失敗 !!!!!!!!!!!!!!!!!!");

                                                            // 重新下載嘗試
                                                            bool successNCdata_try = false;
                                                            if (!successNCdata)
                                                            {
                                                                for (int j = 0; j < reTryDownloadTimes; j++)
                                                                {
                                                                    rwLibOcmLog.log($">> {i}_{_ncTypeCode} 資料集 nc 檔 嘗試重新下載 第 {j + 1} 次");
                                                                    successNCdata_try = rwLibNet.DownloadFile(ocmIndex.ncDataUrl, ocmIndex.ncFilePath);
                                                                    if (successNCdata_try)
                                                                    {
                                                                        rwLibOcmLog.log($">> {i}_{_ncTypeCode} 資料集 nc 檔 嘗試重新下載成功\n");
                                                                        ocmIndexList.Add(ocmIndex);
                                                                        ncDataCount++;
                                                                        break;
                                                                    }
                                                                    else
                                                                        rwLibOcmLog.log($">> {i}_{_ncTypeCode} 資料集 nc 檔 嘗試重新下載失敗 !!!!!!!!!!!!!!!!!!\n");
                                                                }
                                                            }
                                                        }


                                                        // 設置為 onlyNow 只抓最新 1 次的資料
                                                        if (onlyTransNowData)
                                                            // 當成功抓取第 1 份資料時
                                                            if (ncDataCount == 1)
                                                            {
                                                                rwLibOcmLog.log($"目前設置為 onlyNow 只抓取目前時間的資料\n");
                                                                // 重設本次作業應跑過的時間集數量
                                                                downloadNcTimeItemCount = historyDataCount + 1;
                                                                break;
                                                            }

                                                    }
                                                    else
                                                    {
                                                        rwLibOcmLog.log($"[{i}][{time}][{dataDTimeHourName}] 資料集已為歷史資料，不進行下載\n");
                                                        historyDataCount++;
                                                    }

                                                    i++;

                                                    if (i == downloadNcTimeItemCount)
                                                        break;
                                                }
                                                rwLibOcmLog.log($" nc 檔下載數量為 [{ncDataCount}]  ；  歷史 nc 檔不下載的數量為 [{historyDataCount}]\n");


                                            }
                                            else
                                            {
                                                rwLibOcmLog.log($"取得時間資料集數量  [{_times.ToList().Count}] 不為 120，停止\n");
                                            }

                                        }
                                        else
                                        {
                                            rwLibOcmLog.log($">> 時間集 nc 檔下載失敗");
                                            rwLibOcmLog.log($">> {successTimeNc} {ocm_TimeNcUrl}");
                                            rwLibOcmLog.log($">> 停止下載所有 nc 檔\n");
                                        }
                                    }
                                    else
                                    {
                                        rwLibOcmLog.log($"{argOcmDataObtainDate} 時間集 nc 檔連線狀態檢查 失敗，不執行資料下載更新作業。\n ");
                                    }

                                    ncTypeCount++;
                                }

                                Console.WriteLine($"ocmIndexList.Count  {ocmIndexList.Count}  (有 nc 檔的 time 數量)");
                                if (ocmIndexList.Count > 0
                                    && (ocmIndexList.Count % _ncTypeCodes.Length) == 0)
                                {
                                    rwLibOcmLog.log($"開始進行 JSON 轉檔 \n");
                                    int jsonCount = 0;//轉出的 json 檔數量
                                    rwLibOcmLog.log($"> JSON 轉檔輸出 Temp 路徑 {json_Temp_FolderPath}");
                                    int oneTypeNcCount = ocmIndexList.Count / _ncTypeCodes.Length;//單一類別下載的 nc 檔數量
                                    List<OcmDataIndex> _outOcmIndexList = new List<OcmDataIndex>();//整理後輸出的 index 物件

                                    #region 轉出 json 檔
                                    for (int t = 0; t < oneTypeNcCount; t++)
                                    {//time

                                        List<OcmData> _ocm_datalist = new List<OcmData>();
                                        OcmDataIndex _outOcmIndex = new OcmDataIndex();
                                        string dataDTimeHourName = "";
                                        for (int tc = 0; tc < _ncTypeCodes.Length; tc++)
                                        {//typecode
                                            int index = t + (tc * oneTypeNcCount);
                                            OcmDataIndex odi = ocmIndexList[index];
                                            OcmData _od = rwLibNetcdf.GetOcmNetCdfData(odi.ncFilePath);
                                            _ocm_datalist.Add(_od);

                                            //將最大、最小值更新至 Index 中
                                            if (_ncTypeCodes.Length == 1)
                                            {
                                                odi.max_Intensity = _od.maximum;
                                                odi.min_Intensity = _od.minimum;
                                            }
                                            _outOcmIndex = odi;
                                            dataDTimeHourName = odi.name.Split('_')[0];
                                        }

                                        //整理要輸出的資料項目內容
                                        var _outDataList = from _ocmdata in _ocm_datalist
                                                           select new
                                                           {
                                                               _ocmdata.header,
                                                               _ocmdata.data
                                                           };

                                        string json = JsonConvert.SerializeObject(_outDataList);
                                        string outJsonFileName = $"{dataDTimeHourName}_{string.Join("_", _ncTypeCodes)}.json";
                                        string outJsonFilePath = $"{json_Temp_FolderPath}{outJsonFileName}";

                                        File.WriteAllText(outJsonFilePath, json);
                                        rwLibOcmLog.log($">> 轉出 {outJsonFileName}  檔\n");
                                        jsonCount++;

                                        // 更新 index 的 json 路徑
                                        _outOcmIndex.name = outJsonFileName;
                                        _outOcmIndexList.Add(_outOcmIndex);

                                    }
                                    #endregion

                                    #region 整理  index 資料，轉出 index json 檔

                                    //輸出索引檔 ocmIndexList  轉 json 
                                    //篩選輸出於 json 中的屬性項目
                                    string jsonIndex = "";
                                    if (_ncTypeCodes.Length > 1)
                                    {
                                        jsonIndex = JsonConvert.SerializeObject(from data in _outOcmIndexList
                                                                                select new
                                                                                {
                                                                                    data.timeUTC,
                                                                                    data.time08,
                                                                                    data.name
                                                                                });
                                    }
                                    else
                                    {
                                        jsonIndex = JsonConvert.SerializeObject(from data in _outOcmIndexList
                                                                                select new
                                                                                {
                                                                                    data.timeUTC,
                                                                                    data.time08,
                                                                                    data.name,
                                                                                    data.min_Intensity,
                                                                                    data.max_Intensity,
                                                                                });
                                    }

                                    string outJsonIndexFilePath = $"{json_Temp_FolderPath}index.json";
                                    rwLibOcmLog.log($"轉出 index.json 索引檔  {outJsonIndexFilePath} \n");
                                    File.WriteAllText(outJsonIndexFilePath, jsonIndex);

                                    #endregion

                                    // 更新至正式 json 位置
                                    if (jsonCount == _outOcmIndexList.Count)
                                    {
                                        rwLibOcmLog.log($"json 轉出數量符合 nc 檔數量 [{jsonCount}] ，進行正式資料更新替換\n");

                                        //若正式輸出 json 目錄不存在則建立資料夾
                                        rwLibFio.DelAndCreateFolder(json_FolderPath);

                                        //當輸出成功時，將 json 檔移到成功的目錄中替換
                                        DirectoryInfo di = new DirectoryInfo(json_Temp_FolderPath);
                                        foreach (FileInfo jsonFI in di.GetFiles("*.json"))
                                        {
                                            string moveToPath = $"{json_FolderPath}{jsonFI.Name}";
                                            jsonFI.MoveTo(moveToPath, true);
                                            rwLibOcmLog.log($"> 移動 {jsonFI.Name} 檔案至 {moveToPath}\n");
                                        }

                                    }
                                    else
                                    {
                                        //若未轉出成功，則不替換正式目錄中的 json 資料
                                        rwLibOcmLog.log($"json 轉出數量不完整 {jsonCount}，停止更新替換正式資料\n");
                                    }



                                }



                            }
                            catch (Exception ex)
                            {
                                rwLibOcmLog.log($"try catch Exception:  {ex.Message}  {ex.StackTrace}\n\n");
                            }
                        }
                        else
                        {
                            rwLibOcmLog.log($" 無法識別的代碼 \n");
                        }

                        rwLibOcmLog.log($" OPeNDAP OCM 資料下載轉檔 ---- End\n\n");

                    }

                    #endregion


                    #region 刪除資料夾 (提供刪除下載資料原始檔存放的資料夾)

                    if (callEvent.Equals("delFolder"))
                    {
                        Console.WriteLine("刪除資料夾 ---  Begin");

                        try
                        {
                            Console.WriteLine($"刪除資料夾 {argFileOrOutFolderPath}\n");
                            // 清空或刪除資料夾
                            if (System.IO.Directory.Exists(argFileOrOutFolderPath))
                                System.IO.Directory.Delete(argFileOrOutFolderPath, true);
                        }
                        catch (Exception ex)
                        {
                            Console.WriteLine(ex.Message);
                            Console.WriteLine(ex.StackTrace);
                        }

                        Console.WriteLine("刪除資料夾 End");
                    }

                    #endregion

                     
                     

                    Console.WriteLine("程式執行結束");



                }
            
            }




        }

        // The following template implements the following notation:
        // -key1 = some value   -key2 = "some value even with '-' character "  ...
        private const string ParameterQuery = "\\-(?<key>\\w+)\\s*=\\s*(\"(?<value>[^\"]*)\"|(?<value>[^\\-]*))\\s*";

        private static Dictionary<string, string> ParseString(string value)
        {
            var regex = new Regex(ParameterQuery);
            return regex.Matches(value).Cast<Match>().ToDictionary(m => m.Groups["key"].Value.Trim(), m => m.Groups["value"].Value.Trim());
        }
    }
}
