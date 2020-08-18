using NetCDF;
using OAC_opendata_Console.Model;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace OAC_opendata_Console.Libraries.RWLib
{
    class RWLib_NetCDF
    {
        /// <summary>
        /// 取得 Variables 中指定 ComponentName 的 Attributes 值
        /// </summary>
        /// <param name="curVar"></param>
        /// <param name="componentName"></param>
        /// <returns></returns>
        public object GetNcAttValueByVariables(NcVar curVar, string componentName)
        {
            object returnValue = null;
            foreach (NcAtt curAtt in curVar.Attributes)
            {
                if (curAtt.ComponentName.Equals(componentName))
                {
                    switch (curAtt.GetType().ToString())
                    {
                        case "NetCDF.NcAttString":
                            returnValue = ((NcAttString)curAtt).Value;
                            break;
                        case "NetCDF.NcAttDouble":
                            returnValue = ((NcAttDouble)curAtt).Value;
                            break;
                        case "NetCDF.NcAttInt":
                            returnValue = ((NcAttInt)curAtt).Value;
                            break;
                        case "NetCDF.NcAttFloat":
                            returnValue = ((NcAttFloat)curAtt).Value;
                            break;
                    }
                }
            }
            return returnValue;
        }


        /// <summary>
        /// 讀取 nc 檔存入 OcmData 物件
        /// </summary>
        /// <param name="ncFilePath"></param>
        /// <returns></returns>
        public OcmData GetOcmNetCdfData(string ncFilePath)
        {
            OcmData _ocm_data = new OcmData();
            OcmHeader _ocmHeader = new OcmHeader();
            List<object> _data = new List<object>();

            NcFile ncFile = new NcFile(ncFilePath);

            //OCM 的 nc 檔
            if (ncFile.DimensionCount == 4)
            {
                float min_Intensity = 0;
                float max_Intensity = 0;
                foreach (NcVar curVar in ncFile.Variables)
                {
                    // 取得 header 屬性
                    switch (curVar.ComponentName)
                    {
                        case "time":
                            _ocmHeader.refTime = (string)this.GetNcAttValueByVariables(curVar, "units");
                            break;

                        case "lat":
                            _ocmHeader.la1 = (int)(((double[])GetNcAttValueByVariables(curVar, "maximum"))[0]);
                            _ocmHeader.la2 = (int)(((double[])GetNcAttValueByVariables(curVar, "minimum"))[0]);
                            _ocmHeader.dy = ((double[])GetNcAttValueByVariables(curVar, "resolution"))[0];
                            _ocmHeader.ny = (int)curVar.ElementCount;
                            break;

                        case "lon":
                            _ocmHeader.lo1 = (int)(((double[])GetNcAttValueByVariables(curVar, "minimum"))[0]);
                            _ocmHeader.lo2 = (int)(((double[])GetNcAttValueByVariables(curVar, "maximum"))[0]);
                            _ocmHeader.dx = ((double[])GetNcAttValueByVariables(curVar, "resolution"))[0];
                            _ocmHeader.nx = (int)curVar.ElementCount;
                            break;

                        case "UCURR":
                        case "VCURR":
                            _ocmHeader.parameterCategory = 2;
                            break;
                        case "SALT":
                        case "SST":
                        case "WL":
                            _ocmHeader.parameterCategory = 0;
                            break;
                    }
                    switch (curVar.ComponentName)
                    {
                        case "UCURR":
                            _ocmHeader.parameterNumber = 2;
                            break;
                        case "VCURR":
                            _ocmHeader.parameterNumber = 3;
                            break;
                        case "SALT":
                        case "SST":
                        case "WL":
                            _ocmHeader.parameterNumber = 0;
                            break;

                    }

                    // 填入 data 
                    curVar.ReadData();
                    switch (curVar.ComponentName)
                    {
                        case "UCURR":
                        case "VCURR":
                        case "SALT":
                        case "SST":
                        case "WL":
                            float[] _datafloat = ((NcVarTyped<float>)curVar).Data;
                            for (int i = 0; i < _datafloat.Length; i++)
                            {
                                if (_datafloat[i].Equals(-999))
                                    _data.Add("");
                                else
                                {
                                    float _df = float.Parse(_datafloat[i].ToString("f5"));
                                    _data.Add(_df);
                                    min_Intensity = (_df < min_Intensity) ? _df : min_Intensity;
                                    max_Intensity = (_df > max_Intensity) ? _df : max_Intensity;
                                }

                            }
                            //curVar.Attributes.AsQueryable().Where(o => o.ComponentName.Equals(""){ obj})
                            break;
                    }
                }

                _ocm_data.minimum = min_Intensity;
                _ocm_data.maximum = max_Intensity;
                _ocm_data.average = (min_Intensity + max_Intensity) / 2;
                _ocm_data.header = _ocmHeader;
                _ocm_data.data = _data;

            }

            return _ocm_data;

        }


        /// <summary>
        /// 取得 nc 檔的資訊內容
        /// </summary>
        /// <param name="ncFilePath"></param>
        /// <returns></returns>
        public string GetNcFileinfo(string ncFilePath)
        {
            StringBuilder sb = new StringBuilder();

            NcFile ncFile = new NcFile(ncFilePath);
            sb.AppendLine($"開始讀取 nc 檔 - {ncFilePath}\n");
            sb.AppendLine(String.Format("Global attributes ({0})", ncFile.AttributeCount));
            //foreach (NcComponent curAtt in ncFile.Attributes)
            //    sb.AppendLine(curAtt.ToString());             
            sb.AppendLine(string.Join(Environment.NewLine, ncFile.Attributes.ToList()));


            sb.AppendLine("\n----------------------------");
            sb.AppendLine(String.Format("Dimensions ({0})\n", ncFile.DimensionCount));
            foreach (NcDim curDim in ncFile.Dimensions)
                if (curDim.IsUnlimited)
                    sb.AppendLine(String.Format("{0}[{1}] (unlimited)", curDim.ComponentName, curDim.Size));
                else
                    sb.AppendLine(String.Format("{0}[{1}]", curDim.ComponentName, curDim.Size));

            sb.AppendLine("\n----------------------------");
            sb.AppendLine(String.Format("Variables ({0})", ncFile.VariableCount));
            foreach (NcVar curVar in ncFile.Variables)
            {
                sb.AppendLine("\n........................................");
                sb.AppendLine(String.Format("  NcType(Name)  >>  {0}({1})", curVar.NcType.ToString(), curVar.ComponentName));
                sb.AppendLine("........................................\n");


                sb.AppendLine(String.Format("  Elements: {0:###,###,###,###,###}", curVar.ElementCount));
                sb.AppendLine(String.Format("  Size: {0:###,###,###,###,###} (bytes)", curVar.Size));

                sb.AppendLine("\n  ........................................");
                sb.AppendLine(String.Format("  Attributes ({0})", curVar.Attributes.Count));
                for (int i = 0; i < curVar.Attributes.Count; i++)
                    sb.AppendLine(String.Format("    Attributes[{0}][Name](ToString)  >>  [{1}]({2})", i, curVar.Attributes[i].ComponentName, curVar.Attributes[i].ToString()));

                sb.AppendLine("\n  ........................................");
                sb.AppendLine(String.Format("\n  Dimensions ({0})", curVar.Dimensions.Count));
                for (int i = 0; i < curVar.Dimensions.Count; i++)
                    sb.AppendLine(String.Format("    Dimensions[i][Name](Size)  >>  [{0}][{1}]({2})", i, curVar.Dimensions[i].ComponentName, curVar.Dimensions[i].Size));


            }
            sb.AppendLine("----------------------------");

            return sb.ToString();
        }
          

    }
}
