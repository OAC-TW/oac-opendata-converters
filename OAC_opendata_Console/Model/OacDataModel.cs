using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Xml;
using System.Xml.Serialization;

namespace OAC_opendata_Console.Model
{

    #region OCM 海象資料


    /// <summary>
    /// 提供 OCM NetCDF 轉檔後存的 Json 格式
    /// </summary>
    public class OcmDataSet
    {
        public List<OcmData> dataset { get; set; }

    }
    public class OcmData
    {
        public OcmHeader header { get; set; }
        public List<object> data { get; set; }
        /// <summary>
        /// 存放本資料集中 data 的最大最小值
        /// </summary>
        public float minimum { get; set; } = 0;
        public float maximum { get; set; } = 0;
        public float average { get; set; } = 0;


    }
    public class OcmHeader
    {
        public int parameterCategory { get; set; }
        public int parameterNumber { get; set; }
        public int scanMode { get; set; }
        public int nx { get; set; }
        public int ny { get; set; }
        public int lo1 { get; set; }
        public int la1 { get; set; }
        public int lo2 { get; set; }
        public int la2 { get; set; }
        //resolution
        public double dx { get; set; }
        public double dy { get; set; }
        public string refTime { get; set; }


    }

    /// <summary>
    /// 下載資料集的日期索引檔
    /// </summary>
    public class OcmDataSetIndex
    {
        public string ApiUrl { get; set; }

        public List<OcmDataIndex> DataList { get; set; }


    }
    /// <summary>
    /// 單一筆資料索引內容
    /// </summary>
    public class OcmDataIndex
    {
        /// <summary>
        /// 索引編號
        /// </summary>
        public int Index { get; set; }
        /// <summary>
        /// 時間集編號
        /// </summary>
        public int Time { get; set; }

        /// <summary>
        /// 資料日期時間
        /// </summary>
        public string timeUTC { get; set; }
        public string time08 { get; set; }



        /// <summary>
        /// nc檔下載來源網址
        /// </summary>
        public string ncDataUrl { set; get; }

        /// <summary>
        /// nc檔存放位址
        /// </summary>
        public string ncFilePath { set; get; }

        /// <summary>
        /// 轉出的 json 檔名
        /// </summary>
        public string name { get; set; }

        /// <summary>
        /// 該筆資料集的最小、最大值
        /// </summary>
        public float min_Intensity { get; set; }
        public float max_Intensity { get; set; }

    }

    // UTC、+8 時區日期物件
    public class DateUtc_08
    {
        public string timeUTC { get; set; }
        public string time08 { get; set; }
        public DateTime dt08 { get; set; }
        public DateTime dtUTC { get; set; }

    }



    #endregion

    #region  運研所海氣象資料轉檔 - 商港觀測資料


    public class isoheStation
    {

        public string Name { get; set; }
        public string Code { get; set; }
        public string XmlSourceUrl { get; set; }
        public string OpendataLinkUrl { get; set; }

    }


    public class StationFieldsDesc
    {
        public string Name { get; set; } = "測站名稱";
        public string Code { get; set; } = "代碼";
        public string XmlSourceUrl { get; set; } = "XML資料來源網址";
        public string OpendataLinkUrl { get; set; } = "OpenData資料集網址";

    }

    public class StationList
    {
        public FieldsDesc FieldsDesc { get; set; }
        public List<Station> Stations { get; set; }

    }

    /// <summary>
    /// 單一類測站的坐標位置
    /// </summary>
    public class StationLocation
    {
        public double Latitude { get; set; } = 0.0;
        public double Longitude { get; set; } = 0.0;
    }

    /// <summary>
    /// 欄位描述
    /// </summary>
    public class FieldsDesc
    {
        public StationFieldsDesc StationFieldsDesc { get; set; }
        public TideDataFieldsDesc TideDataFieldsDesc { get; set; }
        public HistoryDataFieldsDesc HistoryDataFieldsDesc { get; set; }
        public WindDataFieldsDesc WindDataFieldsDesc { get; set; }

    }

    /// <summary>
    /// 單一測站資訊及資料集
    /// </summary>
    public class Station
    {
        public string Name { get; set; }
        public string Code { get; set; } = "代碼";
        public string XmlSourceUrl { get; set; } = "XML資料來源網址";
        public string OpendataLinkUrl { get; set; } = "OpenData資料集網址";
        public StationDataSet DataSet { get; set; }

    }

    public class StationDataSet
    {
        public TideData TideData { get; set; }
        public HistoryData HistoryData { get; set; }
        public WindData WindData { get; set; }

    }


    public class TideDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string TideValue { get; set; } = "潮位(公尺)";

    }
    public class TideDataItem
    {
        public string Date_Time { get; set; }
        public double TideValue { get; set; }
    }
    public class TideData
    {
        public StationLocation location { get; set; }
        public List<TideDataItem> Data { get; set; }

    }

    public class HistoryDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string HS { get; set; } = "波高(公尺)";
        public string TP { get; set; } = "尖峰週期(秒)";
        public string MDIR { get; set; } = "波向(度)";
        public string Tmean { get; set; } = "平均週期(秒)";
        public string Velocity { get; set; } = "流速(公尺/秒)";
        public string Vmdir { get; set; } = "流向(度)";

    }
    public class HistoryDataItem
    {
        public string Date_Time { get; set; }
        public double HS { get; set; }
        public double TP { get; set; }
        public double MDIR { get; set; }
        public double Tmean { get; set; }
        public double? Velocity { get; set; }
        public double Vmdir { get; set; }
    }
    public class HistoryData
    {
        public StationLocation location { get; set; }
        public List<HistoryDataItem> Data { get; set; }

    }

    public class WindDataFieldsDesc
    {
        public string Date_Time { get; set; } = "觀測時間";
        public string WS_AVG { get; set; } = "平均風速(公尺/秒)";
        public string WD_AVG { get; set; } = "平均風向(度)";

    }
    public class WindDataItem
    {
        public string Date_Time { get; set; }
        public double WS_AVG { get; set; }
        public double WD_AVG { get; set; }
    }
    public class WindData
    {
        public StationLocation location { get; set; }
        public List<WindDataItem> Data { get; set; }

    }



    #endregion

}
