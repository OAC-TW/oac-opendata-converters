using OAC_opendata_Console.Model;
using System;
using System.Collections.Generic;
using System.Text;

namespace OAC_opendata_Console.Libraries.RWLib
{
    class RWLib_Date
    {

        /// <summary>
        /// nc 檔 自 1800-01-01 00:00:00 起算 time 小時的日期，取回 utc 及 +8 時區的時間
        /// </summary>
        /// <param name="time"></param>
        /// <returns></returns>
        public DateUtc_08 GetUtcAnd08DateTimeFromHours(int time)
        {
            DateTime dtinit = new DateTime(1800, 1, 1, 0, 0, 0);
            DateTime dt08 = TimeZoneInfo.ConvertTimeFromUtc(dtinit.AddHours(time), TimeZoneInfo.FindSystemTimeZoneById("Taipei Standard Time"));
            DateTime dtUtc = TimeZoneInfo.ConvertTimeFromUtc(dtinit.AddHours(time), TimeZoneInfo.Utc);

            DateUtc_08 dts = new DateUtc_08
            {
                timeUTC = dtUtc.ToString("yyyy-MM-ddTHH:mm:ssZ"),
                time08 = dt08.ToString("yyyy-MM-ddTHH:mm:ss+08:00"),
                dt08 = dt08,
                dtUTC = dtUtc
            };

            return dts;

        }


    }
}
