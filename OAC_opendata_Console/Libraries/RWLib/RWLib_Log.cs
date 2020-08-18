using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace OAC_opendata_Console.Libraries.RWLib
{
    class RWLib_Log
    {
        private string _logFilePath;
        public RWLib_Log(string logFileFolderPath, string logName = "")
        {
            if (!Directory.Exists(Path.GetDirectoryName(logFileFolderPath)))
                Directory.CreateDirectory(Path.GetDirectoryName(logFileFolderPath));

            this._logFilePath = $"{logFileFolderPath}/{(logName.Equals("") ? "" : logName + "_")}{DateTime.Now.ToString("yyyyMMdd")}.txt";
        }

        public void log(string logMsg)
        {
            string datetime = DateTime.Now.ToString("yyyy/MM/dd HH:mm:ss");
            string _logMsg = $"[{datetime}]  {logMsg}";

            Console.WriteLine(_logMsg);
            using (StreamWriter sw = File.AppendText($"{this._logFilePath}"))
            {
                sw.WriteLine(_logMsg);
            }
        }

    }
}
