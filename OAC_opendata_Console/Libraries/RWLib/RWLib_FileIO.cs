using System;
using System.Collections.Generic;
using System.IO;
using System.Text;
using System.Threading;

namespace OAC_opendata_Console.Libraries.RWLib
{
    class RWLib_FileIO
    {

        /// <summary>
        /// 清空或刪除資料夾
        /// </summary>
        /// <param name="folderPath"></param>
        public void DelAndCreateFolder(string folderPath)
        {
            if (Directory.Exists(Path.GetDirectoryName(folderPath)))
                Directory.Delete(Path.GetDirectoryName(folderPath), true);
            Directory.CreateDirectory(Path.GetDirectoryName(folderPath));
            Thread.Sleep(1000);
        }


    }
}
