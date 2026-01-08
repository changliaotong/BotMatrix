using System.Data;
using BotWorker.Infrastructure.Persistence.Database;

namespace BotWorker.Modules.Buses
{
    public class BusStops : MetaData<BusStops>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus_stops";
        public override string KeyField => "bus_id";
        public override string KeyField2 => "stop_id";

        public static long GetStopID(long stopsId)
        {
            return GetLong("stop_id", stopsId);            
        }
    }
    
    public class Stop : MetaData<Stop>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus_stop";
        public override string KeyField => "stop_id";

        //① ② ③ ④ ⑤ ⑥ ⑦ ⑧ ⑨ ⑩
        public static string[] stop_nos = { "", "①", "②", "③", "④", "⑤", "⑥", "⑦", "⑧", "⑨", "⑩" };

        //获得站点坐标
        public static string GetStopPos(long stop_id)
        {
            return Get<string>("stop_pos", stop_id);
        }

        //站点坐标更新
        public static int UpdateStopPos(long stop_id, string stop_pos)
        {
            return Update($"stop_pos = {stop_pos.Quotes()}", stop_id);
        }

        //判断是否存在站点名 stop_name,存在返回 stop_id 否则返回 0
        public static long GetStopId(string stopName)
        {
            return GetWhere("stop_id", $"stop_name = {stopName.Quotes()} or stop_name2 = {stopName.Quotes()}").AsLong();            
        }

        public static string StopInBusA(string bus_id, string stop_id)
        {
            return QueryScalar<string>($"select top 1 stops_id from bus_stops where stop_order is null and bus_id = {bus_id} and stop_id = {stop_id}") ?? "";
        }

        public static string StopInBusB(string bus_id, string stop_id)
        {
            return QueryScalar<string>($"select top 1 stops_id from bus_stops where stop_order2 is null and bus_id = {bus_id} and stop_id = {stop_id}") ?? "";
        }

        //途经站点信息预处理
        public static string HandleStopsStr(string StopStr)
        {
            string tStr = StopStr;
            tStr = tStr.Replace("'", "").Replace(" ", "-").Replace("\\", "-").Replace(",", "-").Replace("，", "-").Replace("、", "-").Replace("－", "-").Replace("\r\n", "-");
            while (tStr.IndexOf("--") > 0)
            {
                tStr = tStr.Replace("--", "-");
            }
            return tStr;
        }

        //新增一个站点，如果存在返回 stop_id，否则新增并返回 stop_id
        public static long AddStop(string stopName)
        {
            string stop_no = "";
            HandelStopName(ref stopName, ref stop_no);
            long stopId = GetStopId(stopName);
            if (stopId != 0)
                return stopId;
            Exec($"insert into bus_stop(stop_name,stop_name_py) values('{stopName}',dbo.GetPY('{stopName}'))");
            return GetAutoId("bus_stop");
        }

        //按顺序添加上行站点到线路 bus_id 
        public static int AddStopToBusA(string stop_id, string bus_id, string stop_order, string stop_no)
        {
            if (stop_no == "") stop_no = "null";
            string stops_id = StopInBusA(bus_id, stop_id).ToString();
            string sql;
            if (stops_id.IsNum())
                sql = $"update bus_stops set stop_order={stop_order},stop_no={stop_no} where stops_id = {stops_id}";
            else
                sql = $"insert into bus_stops(bus_id,stop_id,stop_order,stop_no) values({bus_id},{stop_id},{stop_order},{stop_no})";
            return Exec(sql);
        }

        //按顺序添加下行站点到线路 bus_id
        public static int AddStopToBusB(string stop_id, string bus_id, string stop_order, string stop_no)
        {
            if (stop_no == "") stop_no = "null";
            string stops_id = StopInBusB(bus_id, stop_id);
            string sql;
            if (stops_id.IsNum())
                sql = $"update bus_stops set stop_order2 = {stop_order},stop_no2={stop_no} where stops_id = {stops_id}";
            else
                sql = $"insert into bus_stops(bus_id,stop_id,stop_order2,stop_no2) values({bus_id},{stop_id},{stop_order},{stop_no})";
            return Exec(sql);
        }

        //站点名称预处理
        public static void HandelStopName(ref string stop_name, ref string stop_no)
        {
            //④③②①的处理① ② ③ ④ ⑤ ⑥ ⑦ ⑧ ⑨ ⑩
            if (stop_name.IndexOf("①") > 0 || stop_name.IndexOf("(1)") > 0)
                stop_no = "1";
            else if (stop_name.IndexOf("②") > 0 || stop_name.IndexOf("(2)") > 0)
                stop_no = "2";
            else if (stop_name.IndexOf("③") > 0 || stop_name.IndexOf("(3)") > 0)
                stop_no = "3";
            else if (stop_name.IndexOf("④") > 0 || stop_name.IndexOf("(4)") > 0)
                stop_no = "4";
            else if (stop_name.IndexOf("⑤") > 0 || stop_name.IndexOf("(5)") > 0)
                stop_no = "5";
            else if (stop_name.IndexOf("⑥") > 0 || stop_name.IndexOf("(6)") > 0)
                stop_no = "6";
            else if (stop_name.IndexOf("⑦") > 0 || stop_name.IndexOf("(7)") > 0)
                stop_no = "7";
            else if (stop_name.IndexOf("⑧") > 0 || stop_name.IndexOf("(8)") > 0)
                stop_no = "8";
            else if (stop_name.IndexOf("⑨") > 0 || stop_name.IndexOf("(9)") > 0)
                stop_no = "9";
            else if (stop_name.IndexOf("⑩") > 0 || stop_name.IndexOf("(10)") > 0)
                stop_no = "10";
            else
                stop_no = "";

            stop_name = stop_name.Replace("⑤", "").Replace("④", "").Replace("③", "").Replace("②", "").Replace("①", "").Replace("(1)", "").Replace("(2)", "").Replace("(3)", "").Replace("(4)", "").Replace("(5)", "");
            stop_name = stop_name.Replace("⑩", "").Replace("⑨", "").Replace("⑧", "").Replace("⑦", "").Replace("⑥", "").Replace("(6)", "").Replace("(7)", "").Replace("(8)", "").Replace("(9)", "").Replace("(10)", "");
            // to do
            // 公交站点名称的替换换到数据库中，需可以添加、修改；
        }

        public static void AddStopStrToBusA(string stops_str, string bus_id)
        {
            //去掉无效的字符 分割字符替换为 -
            stops_str = HandleStopsStr(stops_str.Trim()); //"1 2 333 456 2234 323"
            int i = stops_str.IndexOf("-");
            int j = 1;
            string stop_name = "";
            string stop_no = "";
            while (i >= 0)
            {
                stop_name = stops_str.Substring(0, i);
                HandelStopName(ref stop_name, ref stop_no);
                AddStopToBusA(AddStop(stop_name).ToString(), bus_id, j.ToString(), stop_no);
                j++;
                stops_str = stops_str.Substring(i + 1, stops_str.Length - i - 1);
                i = stops_str.IndexOf("-");
            }
            stop_name = stops_str;
            HandelStopName(ref stop_name, ref stop_no);
            AddStopToBusA(AddStop(stop_name).ToString(), bus_id, j.ToString(), stop_no);
        }

        public static void AddStopStrToBusB(string stops_str, string bus_id)
        {
            //去掉无效的字符 分割字符替换为 -
            stops_str = HandleStopsStr(stops_str.Trim()); //"1 2 333 456 2234 323"
            int i = stops_str.IndexOf("-");
            int j = 1;
            string stop_name = "";
            string stop_no = "";
            while (i >= 0)
            {
                stop_name = stops_str.Substring(0, i);
                HandelStopName(ref stop_name, ref stop_no);
                AddStopToBusB(AddStop(stop_name).ToString(), bus_id, j.ToString(), stop_no);
                j++;
                stops_str = stops_str.Substring(i + 1, stops_str.Length - i - 1);
                i = stops_str.IndexOf("-");
            }
            stop_name = stops_str;
            HandelStopName(ref stop_name, ref stop_no);
            AddStopToBusB(AddStop(stop_name).ToString(), bus_id, j.ToString(), stop_no);
        }

        public static void UpdateBusStopsInfo(string bus_id, string bus_stops_a, string bus_stops_b)
        {
            //所有 stop_order 置为 null 更新后仍未null的删除
            SetStopOrderNull(bus_id);

            //添加上行途经站点信息
            AddStopStrToBusA(bus_stops_a, bus_id);
            if (bus_stops_b != "")
                //添加下行途经站点信息
                AddStopStrToBusB(bus_stops_b, bus_id);

            RemoveAllStopsNull(bus_id);

            UpdateBusStopOrder(bus_id);
            //更新缓存线路信息 bus_stops
            Bus.UpdateBusStops(bus_id);

            //更新线路图
            int i = Convert.ToInt32(bus_id);
            Bus.UpdateBusMap(i, Bus.GetBusMap(i));

            //如果站点信息有变化，则需要更新线路价格表
            Bus.UpdateBusPriceTable(bus_id);
        }

        //得到经过某站点的所有线路 默认空格隔开
        public static string GetStopBuses(string stop_id)
        {
            return Bus.GetStopBuses(stop_id);
        }

        //从站点1到站点2的所有线路名称 用 / 隔开
        public static string GetStopBuses(string stop_id, string stop_id2)
        {
            return Bus.GetStopBuses(stop_id, stop_id2, "/");
        }

        //获得站点ID


        //获得站点名称
        public static string GetStopName(string stop_id)
        {
            if (stop_id.IsNum())
                return QueryScalar<string>("select stop_name as res from bus_stop where stop_id = " + stop_id) ?? "";
            else
                return "";
        }

        //获得站点名称
        public static string GetStopName2(string stop_id)
        {
            if (stop_id.IsNum())
                return QueryScalar<string>("select stop_name2 as res from bus_stop where stop_id = " + stop_id) ?? "";
            else
                return "";
        }

        //更新站点名称（如果存在则合并两个站点）
        public static int UpdateStopName(string stop_id, string stop_name)
        {
            stop_name = stop_name.Replace("'", "");
            string sql = "exec sp_UpdateStopName " + stop_id + ",'" + stop_name + "'";
            return Exec(sql);
        }

        //更新线路途经站点顺序
        public static int UpdateBusStopOrder(string bus_id)
        {
            return Exec($"update bus_stops set stop_order3 = dbo.getstoporder3(stops_id) where bus_id = {bus_id}");
        }

        //去除某线路所有途经站点
        public static int RemoveAllStopsNull(string bus_id)
        {
            return Exec($"delete from bus_stops where stop_order is null and stop_order2 is null and bus_id = {bus_id}");
        }

        //标记某线路所有途经站点顺序为空
        public static int SetStopOrderNull(string bus_id)
        {
            return UpdateWhere($"stop_order = null,stop_order2 = null", $"bus_id = {bus_id}");
        }

        //mark stops around the place
        public static string MarkStops(string stops, string place_id)
        {
            return MarkStops(stops, "", place_id);
        }

        //mark stops around the place
        public static string MarkStops(string stops, string bus_name, string place_id)
        {
            string connString = ConnString;
            using var myConnection = DbProviderFactory.CreateConnection();
            string sql = "select * from bus_stop where stop_id in (select stop_id from place_stops where place_id = " + place_id + ") order by len(stop_name) desc";
            using var myCommand = myConnection.CreateCommand();
            myCommand.CommandText = sql;
            myConnection.Open();
            using var reader = myCommand.ExecuteReader(CommandBehavior.CloseConnection);
            while (reader.Read())
            {
                int idCol = reader.GetOrdinal("stop_name");
                string stop_name = reader.GetString(idCol);
                if (stop_name != null && stop_name != "")
                {
                    string place_name = Place.GetPlaceName(place_id);
                    stops = MarkStop(stops, bus_name, place_name, stop_name);
                }
            }
            return stops;
        }

        public static string MarkStop(string stops, string bus_name, string place_name, string stop_name)
        {
            string hint = string.Format("({0}附近)", place_name);
            //① ② ③ ④ ⑤ ⑥ ⑦ ⑧ ⑨ ⑩
            string[] stop_nos = { "", "①", "②", "③", "④", "⑤", "⑥", "⑦", "⑧", "⑨", "⑩" };
            for (int i = 0; i < 10; i++)
            {
                //如果 place_name 是 stop_name 的一部分则不用显示
                if (stop_name.Contains(place_name))
                    hint = "";
                if (bus_name.Contains("地铁"))
                    stops = stops.Replace("><em>" + stop_name + stop_nos[i] + "</em></a>", "><font color=\"red\"><strong><em>" + stop_name + stop_nos[i] + "</em></strong></font></a><em>" + hint + "(温馨提示：需要换乘才能到达此站)</em>");
                else
                {
                    stops = stops.Replace("><em>" + stop_name + stop_nos[i] + "</em></a>", "><font color=\"red\"><strong><em>" + stop_name + stop_nos[i] + "</em></strong></font></a><em>(单边停靠站)</em>" + hint);
                    stops = stops.Replace("><em>" + stop_name + "</em>" + stop_nos + "</a>", "><font color=\"red\"><strong><em>" + stop_name + stop_nos[i] + "</em></strong></font></a><em>(单边停靠站)</em>" + hint);
                }
                stops = stops.Replace(">" + stop_name + stop_nos[i] + "</a>", "><font color=\"red\"><strong>" + stop_name + stop_nos[i] + "</strong></font></a>" + hint);
            }

            if (stop_name.Contains("地铁站"))
                stops = MarkStop(stops, bus_name, place_name, stop_name.Replace("地铁站", ""));

            return stops;
        }
    }
}
