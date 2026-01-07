using System.Text.RegularExpressions;
using BotWorker.Common;
using BotWorker.Core.MetaDatas;
using BotWorker.Core.Database;

namespace BotWorker.Bots.Models.Buses
{
    public class Search : MetaData<Search>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus_search";
        public override string KeyField => "search_id";

        //输入关键字，返回查询结果应跳转到的URL
        public static string GetKeywordUrl(string keyword, ref string answer, bool isMobile)
        {
            string client_id = "";
            string client_ip = "";
            return GetKeywordURL(client_id, client_ip, keyword, ref answer, isMobile);
        }


        public static string GetKeywordUrl(string keyword)
        {
            string answer = "";
            return GetKeywordUrl(keyword, ref answer);
        }

        public static string GetKeywordUrl(string keyword, ref string answer)
        {
            return GetKeywordUrl(keyword, ref answer, false);
        }

        public static string GetKeywordUrl(string keyword, bool isMobile)
        {
            string answer = "";
            return GetKeywordUrl(keyword, ref answer, isMobile);
        }


        public static string GetKeywordUrlBBS(string client_id, string client_ip, string keyword)
        {
            string answer = "BBS";
            string url = GetKeywordURL(client_id, client_ip, keyword, ref answer, false);
            if (answer != "")
                return answer + $"\n{url}" + url;
            else
                return "";
        }

        public static string GetKeywordURL(string client_id, string client_ip, string keyword, ref string answer, bool isMobile)
        {
            string key1 = "";
            string key2 = "";
            string url = "";

            keyword = keyword.Replace("'", "");
            //keyword = keyword.AsJianti();
            if (keyword == "")
            {
                if (isMobile)
                    url = "/m/bus";
                else
                    url = "/bus/all.aspx";
                return url;
            }

            AddSearchHis(client_id.ToString(), client_ip, keyword);
            keyword = RemoveNouseKey(keyword);

            answer = "";

            getKey1Key2(keyword.Trim(), ref key1, ref key2);
            if (key1 == "")
            {
                if (isMobile)
                    url = "/m/bus";
                else
                    url = "/bus/all.aspx";
                return url;
            }

            //只有一个有效关键字
            if (key2 == "")
            {
                //判断是否有该车次	
                string bus_name = GetBusNameKeyword(keyword);
                int c_bus = Bus.GetBusCount(bus_name);
                if (c_bus == 1)
                {
                    string bus_id = Bus.GetBusIDByKey(bus_name);
                    answer = bus_name;
                    if (isMobile)
                        url = "/m/bus/view.aspx?bus_id=" + bus_id;
                    else
                        url = "/bus/" + bus_id + ".shtml";
                }
                else if (c_bus > 1)
                {
                    //多个符合条件线路的情况
                    answer = "有多个符合“" + bus_name + " 的公交线路：";
                    if (isMobile)
                        url = "/m/bus/list.aspx?keyword=" + bus_name;
                    else
                        url = "/bus/k" + bus_name + ".shtml";
                }

                if (url != "") return url;

                key1 = Place.GetShortPlaceName(key1);
                int place_id = Place.GetPlaceID(key1);
                if (place_id != 0)
                {
                    //是否有添加公交站点呢？
                    answer = "关于" + key1 + " 的详细资料：";
                    if (Place.GetStopCountByPlace(place_id.ToString()) > 0)
                    {
                        if (isMobile)
                            url = "/m/bus/list.aspx?place_id=" + place_id.ToString();
                        else
                            url = "/bus/p" + place_id.ToString() + ".shtml";
                    }
                    else
                    {
                        //是否设置了转向？
                        string place_url = Place.GetPlaceUrl(place_id.ToString());
                        if (place_url != "")
                            url = place_url;
                        else
                        {
                            if (isMobile)
                                url = "/m/place/view.aspx?place_id=" + place_id;
                            else
                                url = "/place/" + place_id + ".shtml";
                        }
                    }
                }
                else
                {
                    answer = "";
                    int c_place = Convert.ToInt32(Query("select count(place_id) from sz84..place where place_name like '%" + key1 + "%'"));
                    if (c_place > 0)
                        return "/place/list.aspx?keyword=" + key1;
                    else //线路标签
                    {
                        string tag_id = BusTag.ExistsField("tag_id", key1).ToString();
                        if (tag_id != "0")
                        {
                            if (isMobile)
                                url = "/m/bus/all.aspx?tag_id=" + tag_id;
                            else
                                url = "/bus/all.aspx?tag_id=" + tag_id;
                            return url;
                        }
                        else
                            return "/place/none.aspx?keyword=" + key1;
                    }
                }
            }

            else
            {
                answer = "从 " + key1 + " 到 " + key2 + " 的乘车方案：";
                key1 = Place.GetShortPlaceName(key1);
                key2 = Place.GetShortPlaceName(key2);
                int place_id = Place.GetPlaceID(key1);
                int place_id2 = Place.GetPlaceID(key2);

                if (place_id != 0 & place_id2 != 0)
                {
                    url = GetNostopUrl(place_id, place_id2, isMobile);
                    if (url == "")
                        url = GetChangeUrl(place_id, place_id2, isMobile);
                }

                if (url == "")
                {
                    url = GetNostopUrl2(key1, key2, isMobile);
                    if (url == "")
                    {
                        answer = "";
                        url = "/place/redirect.aspx?key1=" + key1;
                        if (key2 != "")
                            url += "&key2=" + key2;
                    }
                }

            }
            return url;
        }

        public static int AddSearchHis(string clientId, string clientIP, string keyword)
        {
            return Insert([
                new Cov("keyword", keyword),
                new Cov("client_id", clientId),
                new Cov("client_ip", clientIP),
            ]);
        }

        //去掉线路名称中多余的字符
        public static string GetBusNameKeyword(string keyword)
        {
            return keyword.Replace("公交", "").Replace("深圳", "").Replace("路车", "").Replace("号线", "").Replace("路", "").Replace("号", "").Replace("线", "").Replace("专", "专线");
        }

        public static string RemoveNouseKey(string keyword)
        {
            keyword = keyword.Replace("!", " ").Replace("！", " ").Replace("－", " ").Replace("—", " ").Replace("　", " ").Replace("、", " ").Replace(",", " ").Replace("，", " ");
            keyword = keyword.Replace("-", " ").Replace("\"", " ").Replace("“", " ").Replace("”", " ").Replace("?", " ").Replace("？", " ").Replace("\\", "").Replace("：", " ");
            keyword = keyword.Replace("；", " ").Replace("。", " ").Replace("～", " ").Replace("~", " ").Replace("《", " ").Replace("》", " ").Replace(".", " ").Replace("’", " ");
            keyword = keyword.Replace("_", " ").Replace(":", " ").Replace("#", " ").Replace("→", " ");
            if (keyword.Length >= 2)
            {
                keyword = keyword.Replace("那路", "哪路").Replace("座几", "坐几").Replace("座什", "坐什").Replace("座哪", "坐哪").Replace("座地铁", "坐地铁").Replace("公交路线", " ");
                keyword = keyword.Replace("有没有车", " ").Replace("可不可以", " ").Replace("换乘中心", " ").Replace("公交线路", " ").Replace("多长时间", " ").Replace("新年快乐", " ");
                keyword = keyword.Replace("能不能", " ").Replace("会不会", " ").Replace("可直达", " ").Replace("问一下", " ").Replace("请问下", " ").Replace("坐地铁", "");
                keyword = keyword.Replace("几路车", " ").Replace("有没有", " ").Replace("什么车", " ").Replace("直达车", " ").Replace("几号车", " ").Replace("怎么样", " ");
                keyword = keyword.Replace("哪趟车", " ").Replace("那路车", " ").Replace("是不是", " ").Replace("什么站", " ").Replace("哪个站", " ").Replace("问一下", " ").Replace("坐公交", " ");
                keyword = keyword.Replace("哪路车", " ").Replace("哪些车", " ").Replace("公交车", " ").Replace("新年好", " ").Replace("在不在", " ").Replace("多少钱", " ");
                keyword = keyword.Replace("哪些", " ").Replace("哪个", " ").Replace("你好", " ").Replace("谢谢", " ").Replace("多谢", " ").Replace("在吗", " ").Replace("有人", " ").Replace("是否", " ");
                keyword = keyword.Replace("直达", "").Replace("直接", " ").Replace("可以", " ").Replace("公车", " ").Replace("现在", " ").Replace("怎么", " ").Replace("问下", " ").Replace("经不", " ");
                keyword = keyword.Replace("怎样", " ").Replace("如何", " ").Replace("乘车", " ").Replace("坐车", " ").Replace("转车", " ").Replace("换乘", " ").Replace("搭乘", " ").Replace("座车", " ");
                keyword = keyword.Replace("站台", " ").Replace("查询", " ").Replace("知道", " ").Replace("这里", " ").Replace("我想", " ").Replace("想到", " ").Replace("想去", " ").Replace("过去", " ");
                keyword = keyword.Replace("做车", " ").Replace("搭车", "").Replace("经过", " ").Replace("到达", " ").Replace("有车", " ").Replace("应该", " ").Replace("附近", " ").Replace("那里", " ");
                keyword = keyword.Replace("请问", " ").Replace("早晨", " ").Replace("晚上", " ").Replace("几点", " ").Replace("最早", " ").Replace("发车", " ").Replace("您好", " ").Replace("这边", " ");
                keyword = keyword.Replace("路线", " ").Replace("线路", " ").Replace("最晚", " ").Replace("收车", " ").Replace("哪里", " ").Replace("最短", " ").Replace("属于", " ").Replace("都有", " ");
                keyword = keyword.Replace("最快", " ").Replace("最近", " ").Replace("你 好", " ").Replace("那些", " ").Replace("问路", " ").Replace("方便", " ").Replace("出发", "").Replace("多久", " ");
                keyword = keyword.Replace("便宜", " ").Replace("上车", " ").Replace("下车", " ").Replace("什么", " ").Replace("一班", " ").Replace("距离", "").Replace("多远", " ").Replace("远吗", " ");
                keyword = keyword.Replace("不是", " ").Replace("几路", " ").Replace("几号", " ").Replace("的车", " ").Replace("那个", "").Replace("这个", "").Replace("有没", " ").Replace("没有", " ");
                keyword = keyword.Replace("查询", " ").Replace("hi", "");
            }
            if (keyword.Length > 1)
            {
                keyword = keyword.Replace("请", " ").Replace("走", " ").Replace("坐", " ").Replace("做", " ").Replace("去", " ").Replace("到", " ").Replace("至", " ").Replace("从", " ").Replace("呀", " ");
                keyword = keyword.Replace("啊", " ").Replace("呢", " ").Replace("的", "").Replace("我", " ").Replace("在", " ").Replace("吗", " ").Replace("再", " ").Replace("哪", " ").Replace("往", " ");
                keyword = keyword.Replace("该", " ").Replace("咋", " ").Replace("做", " ").Replace("要", " ").Replace("乘", " ").Replace("搭", " ").Replace("这", " ").Replace("那", " ").Replace("离", " ");
                keyword = keyword.Replace("①", " ").Replace("②", " ").Replace("③", " ").Replace("④", " ").Replace("⑤", " ");

                if (keyword.IndexOf("有限", 0, keyword.Length - 1) < 0 && keyword.IndexOf("有线", 0, keyword.Length - 1) < 0 && keyword.IndexOf("有色", 0, keyword.Length - 1) < 0)
                    keyword = keyword.Replace("有", " ");
                if (keyword.IndexOf("自由", 0, keyword.Length - 1) < 0)
                    keyword = keyword.Replace("由", " ");
                if (keyword.IndexOf("求是", 0, keyword.Length - 1) < 0)
                    keyword = keyword.Replace("是", " ");
                if (keyword.IndexOf("迷你", 0, keyword.Length - 1) < 0)
                    keyword = keyword.Replace("你", " ");
            }

            keyword = keyword.Replace("大夏", "大厦");
            while (keyword.IndexOf("  ") >= 0) keyword = keyword.Replace("  ", " ");
            keyword = keyword.Trim();
            return keyword;
        }


        public static void getKey1Key2(string keyword, ref string key1, ref string key2)
        {
            //去掉通讯工具附加的广告信息
            //keyword = Message.remove_qq_ads(keyword);

            //分析命令类型
            Regex rx = new(Regexs.Key, RegexOptions.Compiled | RegexOptions.IgnoreCase);

            MatchCollection matches = rx.Matches(keyword);
            foreach (Match match in matches)
            {
                key1 = match.Groups["para_1"].Value.Trim();
                key2 = match.Groups["para_2"].Value.Trim();
            }
        }


        public static string GetDirectUrl(int place_id, int place_id2, bool isMobile)
        {
            if (IsHaveDirectBus(place_id, place_id2))
            {
                if (isMobile)
                    return "/m/bus/list.aspx?place_id=" + place_id.ToString() + "&place_id2=" + place_id2.ToString();
                else
                    return "/bus/p" + place_id.ToString() + "p" + place_id2.ToString() + ".shtml";
            }
            else
                return "";
        }

        public static string GetNostopUrl(int place_id, int place_id2, bool isMobile)
        {
            string sql = "select count(bus_id) as res from sz84..bus where exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id + ") and bus_id = sz84..bus.bus_id) and exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id2 + ") and bus_id = sz84..bus.bus_id)";
            //取得符合条件的线路的数量，大于0时表明有直达车
            string c_res = Query(sql);
            if (c_res != "0") //有直到的车的情况
            {
                if (isMobile)
                    return "/m/bus/list.aspx?place_id=" + place_id.ToString() + "&place_id2=" + place_id2.ToString();
                else
                    return "/bus/p" + place_id.ToString() + "p" + place_id2.ToString() + ".shtml";
            }
            else
                return "";
        }

        public static string GetNostopUrl2(string key1, string key2, bool isMobile)
        {
            string w = " where ";
            if (key1 != "")
            {
                w += " bus_stops_a like '%" + key1 + "%'";
                if (key2 != "")
                    w += " and bus_stops_a like '%" + key2 + "%'";
            }
            else
                w += " 1=2";
            string sql = "select count(bus_id) as res from sz84..bus " + w;
            //取得符合条件的线路的数量，大于0时表明有直达
            string c_res = Query(sql);
            if (c_res != "0") //有直到的车的情况
            {
                string url = "/bus/" + key1;
                if (key2 != "")
                    url += "/" + key2 + ".shtml";
                return url;
            }
            else
                return "";
        }

        public static bool IsHaveDirectBus(int place_id, int place_id2)
        {
            string sql = "select count(bus_id) as res from sz84..bus where exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id + ") and bus_id = sz84..bus.bus_id) and exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id2 + ") and bus_id = sz84..bus.bus_id)";
            sql += " and dbo.getStopCount2(bus_id, " + place_id + "," + place_id2 + ") >= 0 ";
            //取得符合条件的线路的数量，大于0时表明有直达车
            string c_res = Query(sql);
            return c_res != "0";
        }

        public static string GetChangeUrl(int place_id, int place_id2, bool isMobile)
        {
            if (isMobile)
                return "/m/bus/change.aspx?place_id=" + place_id.ToString() + "&place_id2=" + place_id2.ToString();
            else
                return "/bus/c" + place_id.ToString() + "c" + place_id2.ToString() + ".shtml";
        }


        public static string GetNostopUrl(string place_id, string place_id2)
        {
            string sql = "select count(bus_id) as res from sz84..bus where exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id + ") and bus_id = sz84..bus.bus_id) and exists (select 1 from sz84..bus_stops where stop_id in (select stop_id from sz84..place_stops where place_id = " + place_id2 + ") and bus_id = sz84..bus.bus_id)";
            //取得符合条件的线路的数量，大于0时表明有直达车
            string c_res = Query(sql);
            if (c_res != "0") //有直到的车的情况
            {
                return " 的直达车有：\n" + Place.GetPlaceBuses(place_id, place_id2);
            }
            else
                return "";
        }

    }
}
