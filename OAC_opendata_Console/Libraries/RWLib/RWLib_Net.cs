using System;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using OAC_opendata_Console.Model;

namespace OAC_opendata_Console.Libraries.RWLib
{
    class RWLib_Net
    {

        /// <summary>
        ///  檢查網址連線狀態是否正常回應
        /// </summary>
        /// <param name="url">檢測網址</param>
        /// <returns></returns>
        public bool ChekUrlRequestStatus(string url)
        {
            bool status = false;

            try
            {
                HttpWebRequest myHttpWebRequest = (HttpWebRequest)WebRequest.Create(url);
                HttpWebResponse myHttpWebResponse = (HttpWebResponse)myHttpWebRequest.GetResponse();
                if (myHttpWebResponse.StatusCode == HttpStatusCode.OK)
                    status = true;
                myHttpWebResponse.Close();
            }
            catch (WebException e)
            {
                Console.WriteLine("\r\nWebException : {0}", e.Status);
            }
            catch (Exception e)
            {
                Console.WriteLine("\nException : {0}", e.Message);
            }

            return status;
        }


        /// <summary>
        /// 下載檔案
        /// </summary>
        /// <param name="sourceFileUrl">下載網址</param>
        /// <param name="desFileNamePath">檔案存放路徑</param>
        /// <returns></returns>
        public bool DownloadFile(string sourceFileUrl, string desFileNamePath)
        {
            bool flag = false;
            FileStream FStream = null;
            Stream myStream = null;
            try
            {
                // 每次都刪掉重新抓取
                if (File.Exists(desFileNamePath))
                    File.Delete(desFileNamePath);

                // 檔案不儲存建立一個檔案
                FStream = new FileStream(desFileNamePath, FileMode.Create);

                // 開啟網路連線
                HttpWebRequest myRequest = (HttpWebRequest)HttpWebRequest.Create(sourceFileUrl);
                myStream = myRequest.GetResponse().GetResponseStream();
                byte[] btContent = new byte[512];
                int intSize = 0;
                intSize = myStream.Read(btContent, 0, 512);
                while (intSize > 0)
                {
                    FStream.Write(btContent, 0, intSize);
                    intSize = myStream.Read(btContent, 0, 512);
                }
                flag = true; // 下載成功
            }
            catch (Exception ex)
            {
                Console.Write("下載檔案時異常：" + ex.Message);
            }
            finally
            {
                if (myStream != null)
                {
                    myStream.Close();
                    myStream.Dispose();
                }
                if (FStream != null)
                {
                    FStream.Close();
                    FStream.Dispose();
                }
            }
            return flag;
        }


    }
}
