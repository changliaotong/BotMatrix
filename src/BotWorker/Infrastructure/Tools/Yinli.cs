using System.Globalization;

namespace sz84.Infrastructure.Tools
{
    /// <summary>
    /// 农历阴历
    /// </summary>
    public class Yinli
    {
        public static DateTime dateMin = new ChineseLunisolarCalendar().MinSupportedDateTime;
        public static DateTime dateMax = new ChineseLunisolarCalendar().MaxSupportedDateTime;

        public static readonly List<string> tianGans = ["甲", "乙", " 丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"];
        public static readonly List<string> diZhis = ["子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"];
        public static readonly List<string> animals = ["鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"];
        public static readonly List<string> months = ["一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "十二"];
        public static readonly List<string> yinliMonths = ["正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"];
        public static readonly List<string> yinliMonths2 = ["正", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖", "拾", "冬", "腊"];
        public static readonly List<string> yinliDays = [ "初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十", "十一", "十二", "十三", "十四", "十五",
                                                             "十六", "十七", "十八", "十九", "二十", "廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十" ];
        public static readonly List<string> dayOfWeeks = ["日", "一", "二", "三", "四", "五", "六"];
        public static readonly List<string> dayOfWeeks2 = ["一", "二", "三", "四", "五", "六", "日"];

        private static readonly string[] xingZuos = { "白羊座", "金牛座", "双子座", "巨蟹座", "狮子座", "处女座", "天秤座", "天蝎座", "射手座", "摩羯座", "水瓶座", "双鱼座" };

        /// <summary>
        /// 农历年数字
        /// </summary>
        public int Year { get; set; }

        /// <summary>
        /// 农历月数字
        /// </summary>
        public int Month { get; set; }

        /// <summary>
        /// 农历日数字
        /// </summary>
        public int Day { get; set; }

        /// <summary>
        /// 星期几
        /// </summary>
        public DayOfWeek DayOfWeek { get; set; }

        /// <summary>
        /// 闰月数字
        /// </summary>
        public int LeapMonth { get; set; }

        /// <summary>
        /// 干支年
        /// </summary>
        public string? GanzhiYearName { get; set; }

        /// <summary>
        /// 农历年
        /// </summary>
        public string? YearName { get; set; }

        /// <summary>
        /// 农历月
        /// </summary>
        public string? MonthName { get; set; }

        /// <summary>
        /// 农历日
        /// </summary>
        public string? DayName { get; set; }

        /// <summary>
        /// 属相
        /// </summary>
        public string? Animal => GetAnimal(Year);

        private DateTime _date = DateTime.Now;

        #region 基本常量
        private const int MinYear = 1900;
        private const int MaxYear = 2050;
        private const int GanZhiStartYear = 1864; //干支计算起始年
        private static DateTime GanZhiStartDay = new DateTime(1899, 12, 22);//起始日
        private const string HZNum = "零一二三四五六七八九";
        private const int AnimalStartYear = 1900; //1900年为鼠年
        private static DateTime ChineseConstellationReferDay = new DateTime(2007, 9, 13);//28星宿参考值,本日为角
        #endregion

        #region 内部结构
        private struct SolarHolidayStruct
        {
            public int Month;
            public int Day;
            public int Recess; //假期长度
            public string HolidayName;
            public SolarHolidayStruct(int month, int day, int recess, string name)
            {
                Month = month;
                Day = day;
                Recess = recess;
                HolidayName = name;
            }
        }

        private struct LunarHolidayStruct
        {
            public int Month;
            public int Day;
            public int Recess;
            public string HolidayName;

            public LunarHolidayStruct(int month, int day, int recess, string name)
            {
                Month = month;
                Day = day;
                Recess = recess;
                HolidayName = name;
            }
        }

        private struct WeekHolidayStruct
        {
            public int Month;
            public int WeekAtMonth;
            public int WeekDay;
            public string HolidayName;

            public WeekHolidayStruct(int month, int weekAtMonth, int weekDay, string name)
            {
                Month = month;
                WeekAtMonth = weekAtMonth;
                WeekDay = weekDay;
                HolidayName = name;
            }
        }
        #endregion

        #region 二十四节气
        private static string[] _lunarHolidayName =
                    {
                    "小寒", "大寒", "立春", "雨水",
                    "惊蛰", "春分", "清明", "谷雨",
                    "立夏", "小满", "芒种", "夏至",
                    "小暑", "大暑", "立秋", "处暑",
                    "白露", "秋分", "寒露", "霜降",
                    "立冬", "小雪", "大雪", "冬至"
                    };
        #endregion

        #region 二十八星宿
        private static string[] _chineseConstellationName =
            {
                  //四        五      六         日        一      二      三  
                "角木蛟","亢金龙","女土蝠","房日兔","心月狐","尾火虎","箕水豹",
                "斗木獬","牛金牛","氐土貉","虚日鼠","危月燕","室火猪","壁水獝",
                "奎木狼","娄金狗","胃土彘","昴日鸡","毕月乌","觜火猴","参水猿",
                "井木犴","鬼金羊","柳土獐","星日马","张月鹿","翼火蛇","轸水蚓"
            };
        #endregion

        #region 节气数据
        private static string[] SolarTerm = new string[] { "小寒", "大寒", "立春", "雨水", "惊蛰", "春分", "清明", "谷雨", "立夏", "小满", "芒种", "夏至", "小暑", "大暑", "立秋", "处暑", "白露", "秋分", "寒露", "霜降", "立冬", "小雪", "大雪", "冬至" };
        private static int[] sTermInfo = new int[] { 0, 21208, 42467, 63836, 85337, 107014, 128867, 150921, 173149, 195551, 218072, 240693, 263343, 285989, 308563, 331033, 353350, 375494, 397447, 419210, 440795, 462224, 483532, 504758 };
        #endregion

        #region 按公历计算的节日
        private static SolarHolidayStruct[] sHolidayInfo = new SolarHolidayStruct[]{
            new SolarHolidayStruct(1, 1, 1, "元旦"),
            new SolarHolidayStruct(2, 2, 0, "世界湿地日"),
            new SolarHolidayStruct(2, 10, 0, "国际气象节"),
            new SolarHolidayStruct(2, 14, 0, "情人节"),
            new SolarHolidayStruct(3, 1, 0, "国际海豹日"),
            new SolarHolidayStruct(3, 5, 0, "学雷锋纪念日"),
            new SolarHolidayStruct(3, 8, 0, "妇女节"),
            new SolarHolidayStruct(3, 12, 0, "植树节 孙中山逝世纪念日"),
            new SolarHolidayStruct(3, 14, 0, "国际警察日"),
            new SolarHolidayStruct(3, 15, 0, "消费者权益日"),
            new SolarHolidayStruct(3, 17, 0, "中国国医节 国际航海日"),
            new SolarHolidayStruct(3, 21, 0, "世界森林日 消除种族歧视国际日 世界儿歌日"),
            new SolarHolidayStruct(3, 22, 0, "世界水日"),
            new SolarHolidayStruct(3, 24, 0, "世界防治结核病日"),
            new SolarHolidayStruct(4, 1, 0, "愚人节"),
            new SolarHolidayStruct(4, 7, 0, "世界卫生日"),
            new SolarHolidayStruct(4, 22, 0, "世界地球日"),
            new SolarHolidayStruct(5, 1, 1, "劳动节"),
            new SolarHolidayStruct(5, 2, 1, "劳动节假日"),
            new SolarHolidayStruct(5, 3, 1, "劳动节假日"),
            new SolarHolidayStruct(5, 4, 0, "青年节"),
            new SolarHolidayStruct(5, 8, 0, "世界红十字日"),
            new SolarHolidayStruct(5, 12, 0, "国际护士节"),
            new SolarHolidayStruct(5, 31, 0, "世界无烟日"),
            new SolarHolidayStruct(6, 1, 0, "国际儿童节"),
            new SolarHolidayStruct(6, 5, 0, "世界环境保护日"),
            new SolarHolidayStruct(6, 26, 0, "国际禁毒日"),
            new SolarHolidayStruct(7, 1, 0, "建党节 香港回归纪念 世界建筑日"),
            new SolarHolidayStruct(7, 11, 0, "世界人口日"),
            new SolarHolidayStruct(8, 1, 0, "建军节"),
            new SolarHolidayStruct(8, 8, 0, "中国男子节 父亲节"),
            new SolarHolidayStruct(8, 15, 0, "抗日战争胜利纪念"),
            new SolarHolidayStruct(9, 9, 0, "  逝世纪念"),
            new SolarHolidayStruct(9, 10, 0, "教师节"),
            new SolarHolidayStruct(9, 18, 0, "九·一八事变纪念日"),
            new SolarHolidayStruct(9, 20, 0, "国际爱牙日"),
            new SolarHolidayStruct(9, 27, 0, "世界旅游日"),
            new SolarHolidayStruct(9, 28, 0, "孔子诞辰"),
            new SolarHolidayStruct(10, 1, 1, "国庆节 国际音乐日"),
            new SolarHolidayStruct(10, 2, 1, "国庆节假日"),
            new SolarHolidayStruct(10, 3, 1, "国庆节假日"),
            new SolarHolidayStruct(10, 6, 0, "老人节"),
            new SolarHolidayStruct(10, 24, 0, "联合国日"),
            new SolarHolidayStruct(11, 10, 0, "世界青年节"),
            new SolarHolidayStruct(11, 12, 0, "孙中山诞辰纪念"),
            new SolarHolidayStruct(12, 1, 0, "世界艾滋病日"),
            new SolarHolidayStruct(12, 3, 0, "世界残疾人日"),
            new SolarHolidayStruct(12, 20, 0, "澳门回归纪念"),
            new SolarHolidayStruct(12, 24, 0, "平安夜"),
            new SolarHolidayStruct(12, 25, 0, "圣诞节"),
            new SolarHolidayStruct(12, 26, 0, " 诞辰纪念")
           };
        #endregion

        #region 按农历计算的节日
        private static LunarHolidayStruct[] lHolidayInfo = new LunarHolidayStruct[]{
            new LunarHolidayStruct(1, 1, 1, "春节"),
            new LunarHolidayStruct(1, 15, 0, "元宵节"),
            new LunarHolidayStruct(5, 5, 0, "端午节"),
            new LunarHolidayStruct(7, 7, 0, "七夕情人节"),
            new LunarHolidayStruct(7, 15, 0, "中元节 盂兰盆节"),
            new LunarHolidayStruct(8, 15, 0, "中秋节"),
            new LunarHolidayStruct(9, 9, 0, "重阳节"),
            new LunarHolidayStruct(12, 8, 0, "腊八节"),
            new LunarHolidayStruct(12, 23, 0, "北方小年(扫房)"),
            new LunarHolidayStruct(12, 24, 0, "南方小年(掸尘)"),
            //new LunarHolidayStruct(12, 30, 0, "除夕")  //注意除夕需要其它方法进行计算
        };
        #endregion

        #region 按某月第几个星期几
        private static WeekHolidayStruct[] wHolidayInfo = new WeekHolidayStruct[]{
            new WeekHolidayStruct(5, 2, 1, "母亲节"),
            new WeekHolidayStruct(5, 3, 1, "全国助残日"),
            new WeekHolidayStruct(6, 3, 1, "父亲节"),
            new WeekHolidayStruct(9, 3, 3, "国际和平日"),
            new WeekHolidayStruct(9, 4, 1, "国际聋人节"),
            new WeekHolidayStruct(10, 1, 2, "国际住房日"),
            new WeekHolidayStruct(10, 1, 4, "国际减轻自然灾害日"),
            new WeekHolidayStruct(11, 4, 5, "感恩节")
        };
        #endregion

        public Yinli(DateTime dt)
        {
            if (dt >= dateMin && dt <= dateMax)
            {
                ChineseLunisolarCalendar chineseDate = new();
                Year = chineseDate.GetYear(dt);
                Month = chineseDate.GetMonth(dt);
                Day = chineseDate.GetDayOfMonth(dt);
                DayOfWeek = dt.DayOfWeek;
                LeapMonth = chineseDate.GetLeapMonth(Year);//获取第几个月是闰月,等于0表示本年无闰月
                GanzhiYearName = GetYear(Year);
                MonthName = $"{(LeapMonth > 0 && Month == LeapMonth ? "闰" : "")}{(LeapMonth > 0 && Month >= LeapMonth ? yinliMonths[Month - 2] : yinliMonths[Month - 1])}";
                DayName = yinliDays[Day - 1];
            }
            else
            {
                Year = 0;
                Month = 0;
                Day = 0;
                DayOfWeek = dt.DayOfWeek;
                LeapMonth = 0;
                GanzhiYearName = "";
                MonthName = "";
                DayName = "";
            }
        }

        public static string GetAnimal(int nlYear)
        {
            return animals[(nlYear - 4) % 12];
        }

        public static string GetTiangan(int nlYear)
        {
            return tianGans[(nlYear - 4) % 60 % 10];
        }

        public static string GetDizhi(int nlYear)
        {
            return diZhis[(nlYear - 4) % 60 % 12];
        }

        public static string GetYear(int nlYear)
        {
            return $"{GetTiangan(nlYear)}{GetDizhi(nlYear)}{GetAnimal(nlYear)}";
        }


        #region ChineseTwentyFourDay
        /// <summary>
        /// 定气法计算二十四节气,二十四节气是按地球公转来计算的，并非是阴历计算的
        /// </summary>
        /// <remarks>
        /// 节气的定法有两种。古代历法采用的称为"恒气"，即按时间把一年等分为24份，
        /// 每一节气平均得15天有余，所以又称"平气"。现代农历采用的称为"定气"，即
        /// 按地球在轨道上的位置为标准，一周360°，两节气之间相隔15°。由于冬至时地
        /// 球位于近日点附近，运动速度较快，因而太阳在黄道上移动15°的时间不到15天。
        /// 夏至前后的情况正好相反，太阳在黄道上移动较慢，一个节气达16天之多。采用
        /// 定气时可以保证春、秋两分必然在昼夜平分的那两天。
        /// </remarks>
        public string ChineseTwentyFourDay
        {
            get
            {
                DateTime baseDateAndTime = new(1900, 1, 6, 2, 5, 0); //#1/6/1900 2:05:00 AM#
                DateTime newDate;
                double num;
                int y;
                string tempStr = "";

                y = _date.Year;

                for (int i = 1; i <= 24; i++)
                {
                    num = 525948.76 * (y - 1900) + sTermInfo[i - 1];

                    newDate = baseDateAndTime.AddMinutes(num);//按分钟计算
                    if (newDate.DayOfYear == _date.DayOfYear)
                    {
                        tempStr = SolarTerm[i - 1];
                        break;
                    }
                }
                return tempStr;
            }
        }

        //当前日期前一个最近节气
        public string ChineseTwentyFourPrevDay
        {
            get
            {
                DateTime baseDateAndTime = new(1900, 1, 6, 2, 5, 0); //#1/6/1900 2:05:00 AM#
                DateTime newDate;
                double num;
                int y;
                string tempStr = "";

                y = _date.Year;

                for (int i = 24; i >= 1; i--)
                {
                    num = 525948.76 * (y - 1900) + sTermInfo[i - 1];

                    newDate = baseDateAndTime.AddMinutes(num);//按分钟计算

                    if (newDate.DayOfYear < _date.DayOfYear)
                    {
                        tempStr = string.Format("{0}[{1}]", SolarTerm[i - 1], newDate.ToString("yyyy-MM-dd"));
                        break;
                    }
                }

                return tempStr;
            }

        }

        //当前日期后一个最近节气
        public string ChineseTwentyFourNextDay
        {
            get
            {
                DateTime baseDateAndTime = new(1900, 1, 6, 2, 5, 0); //#1/6/1900 2:05:00 AM#
                DateTime newDate;
                double num;
                int y;
                string tempStr = "";

                y = _date.Year;

                for (int i = 1; i <= 24; i++)
                {
                    num = 525948.76 * (y - 1900) + sTermInfo[i - 1];

                    newDate = baseDateAndTime.AddMinutes(num);//按分钟计算

                    if (newDate.DayOfYear > _date.DayOfYear)
                    {
                        tempStr = string.Format("{0}[{1}]", SolarTerm[i - 1], newDate.ToString("yyyy-MM-dd"));
                        break;
                    }
                }
                return tempStr;
            }

        }
        #endregion

    }
}
