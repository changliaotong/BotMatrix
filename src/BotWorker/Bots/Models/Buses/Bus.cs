using Microsoft.Data.SqlClient;
using System.Data;
using System.Text.RegularExpressions;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Models.Buses
{

    public class Bus : MetaData<Bus>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus";
        public override string KeyField => "bus_id";

        #region ��·����

        public static string GetEmTitle(string stops, string busName)
        {
            return Regex.Replace(stops, ".shtml\"><em>", ".shtml\" title=\"" + GetEmInfo2(busName) + "\"><em>");
        }

        public static string GetEmInfo2(string busName)
        {
            if (busName.Contains("����"))
                return "��ܰ��ʾ��б����ʾվ����Ҫ���˲��ܵ���";
            else
                return "��ܰ��ʾ����վ��Ϊ����ͣ��վ�㣬������ͣ����վ��";
        }

        public static string GetEmInfo(string busName)
        {
            if (busName.Contains("����"))
                return "��ܰ��ʾ��<em>б����ʾվ��Ϊ���˺�;��վ��</em>";
            else
                return "��ܰ��ʾ��<em>б����ʾվ��Ϊ������ͣ��վ�㣬������ͣ��</em>";
        }

        public static string GetEmInfo(int busId)
        {
            string busName = GetBusName(busId);
            return GetEmInfo(busName);
        }

        public static string GetBusTags(string busId)
        {
            string sql = $"select a.tag_id, b.tag_name, b.tag_info, b.c_bus from bus_tags a inner join bus_tag b on a.tag_id = b.tag_id where a.bus_id = {busId} order by c_bus desc";
            return $"<div class=\"bus_tag\">{QueryRes(sql, "<a href=\"/bus/all.aspx?tag_id={0}\" title=\"{2}\">{1}</a>")} </div>";
        }

        public static string GetBusCount(int stopId)
        {
            var res = Query<string>($"select count(bus_id) as res from bus where bus_type = 1 and bus_name not like '%����%' and bus_id in (select bus_id from bus_stops where stop_id = {stopId})");
            return res == "" ? "0" : res;
        }

        public static string GetBusCountAll()
        {
            var res = Query<string>($"select count(bus_id) as res from bus where bus_type = 1 and bus_name not like '%����%'");
            return res == "" ? "0" : res;
        }


        //����·��ŵõ���·����
        public static string GetBusName(int busId)
        {
            return Query<string>($"select bus_name from bus where bus_id = {busId}");
        }

        //��·���� �����·�����а������ߡ�����·���ȣ����� �������ߣ��߷�ר�ߣ������ߡ����ߡ����ߵ�ʱ���������ӡ�·��
        public static string GetBusName(string busName)
        {
            string[] str = { "��", "·", "��", "��", "����" };
            for (int i = 0; i < str.Length; i++)
            {
                if (busName.Contains(str[i], StringComparison.CurrentCulture))
                    return busName;
            }
            return busName + "·";
        }


        //�� bus_name �õ� bus_id
        public static string GetBusID(string busName)
        {
            return GetWhere($"bus_id", $"bus_name = {busName.Quotes()} or bus_name2 = {busName.Quotes()}");
        }

        //����һ����·
        public static int AddBus(string busName, string busName2, string busType, string startStop, string endStop, string busInfo, string price, string busOrder, string busPic)
        {
            return Insert([
                new Cov("bus_name", busName),
                new Cov("bus_name2", busName2),
                new Cov("bus_type", busType),
                new Cov("start_stop", startStop),
                new Cov("end_stop", endStop),
                new Cov("bus_info", busInfo),
                new Cov("price", price),
                new Cov("bus_order", busOrder),
                new Cov("bus_pic", busPic),
            ]);
        }

        //�༭��·ǰ������·��Ϣ����ʷ��
        public static int AddBusHis(string busId)
        {
            string sql = $"insert into bus_his([bus_id], [bus_name], [bus_name2], [bus_type], [start_stop], [end_stop], [start_time], [start_time2], [end_time], [end_time2], [price], [pay_method],[bus_price_type], [bus_price], [bus_price2], [sztong], [bus_info], [bus_order], [bus_stops_a], [bus_stops_b], [bus_stops_c], [bus_pic],[insert_date], [update_date], [client_id],[client_ip], [search_times])" +
                         $"select [bus_id], [bus_name], [bus_name2], [bus_type], [start_stop], [end_stop], [start_time], [start_time2], [end_time], [end_time2], [price], [pay_method],[bus_price_type], [bus_price], [bus_price2], [sztong], [bus_info], [bus_order], [bus_stops_a], [bus_stops_b], [bus_stops_c], [bus_pic],[insert_date], [update_date], [client_id],[client_ip], [search_times] from bus where bus_id = {busId}";
            return Exec(sql);
        }

        //������·������Ϣ
        public static int UpdateBus(string busId, string busName, string busName2, string busType, string startStop, string endStop, string busInfo, string price, string busOrder, string busPic)
        {
            //����ǰ���ݵ���ʷ���Ͽ���
            AddBusHis(busId);
            return Update($"bus_name = {busName.Quotes()}, bus_name2 = {busName2.Quotes()}, bus_type={busType}, start_stop={startStop.Quotes()}, end_stop={endStop.Quotes()}, " +
                          $"bus_info={busInfo.Quotes()}, price={price.Quotes()}, bus_order = {busOrder},bus_pic={busPic.Quotes()}", busId);
        }

        //������·�۸���Ϣ
        public static int UpdateBusPrice(string busId, string sztong, string busPrice, string busPrice2, string payMethod, string busPriceType)
        {
            return Update($"pay_method = {payMethod.Quotes()},bus_price={busPrice.Quotes()},bus_price2={busPrice2.Quotes()},sztong={sztong.Quotes()},bus_price_type={busPriceType.Quotes()} ", busId);
        }


        //������·;��վ����Ϣ
        public static int UpdateBusStops(string busId)
        {
            return Exec("update bus set bus_stops_a = dbo.getBusStopsA(bus_id),bus_stops_b=dbo.getBusStopsB(bus_id),bus_stops_c=dbo.getBusStopsC(bus_id),update_date=getdate() where bus_id = " + busId);
        }

        //�� keyword �õ����������ĵ�һ����·�� bus_id
        public static string GetBusIDByKey(string keyword)
        {
            keyword = Search.GetBusNameKeyword(keyword);
            return Query("select top 1 bus_id as res from bus where bus_type=1 and bus_name like '%" + keyword + "%' or bus_name2 like '%" + keyword + "%'");
        }

        public static int GetBusCount(string keyword)
        {
            return Convert.ToInt32(Query("select count(bus_id) as res from bus where bus_type=1 and (bus_name like '%" + keyword + "%' or bus_name2 like '%" + keyword + "%')"));
        }

        //�õ�����������Ϣ����·����
        public static string GetBusNameWithCityName(string busName)
        {
            return "����" + GetBusName(busName);
        }

        public static string GetBusNameWithCityName(int busId)
        {
            return GetBusNameWithCityName(GetBusName(busId));
        }


        //�����· bus_id ��·�� palce_id �� palce_id2 ����վ����
        public static string GetStopCount(int bus_id, int place_id, int place_id2)
        {
            if (place_id == 0 | place_id2 == 0)
                return "";
            else
                return "�˳�<strong><font color=\"red\">" + Query("select dbo.getStopCount2(" + bus_id.ToString() + "," + place_id.ToString() + "," + place_id2.ToString() + ") as res") + "</font></strong>վ ��" + GetStopCount(bus_id) + "վ";
        }

        //ȡ����·վ����
        public static int GetStopCount(int bus_id)
        {
            return Convert.ToInt32(Query("select count(stop_id) as res from bus_stops where bus_id = " + bus_id.ToString()));
        }

        //ȡ������վ����
        public static int getStopCountA(int bus_id)
        {
            return Convert.ToInt32(Query("select isnull(max(stop_order),0) as res from bus_stops where bus_id = " + bus_id.ToString()));
        }

        //ȡ�÷���վ����
        public static int getStopCountB(int bus_id)
        {
            return Convert.ToInt32(Query<string>("select isnull(max(stop_order2),0) as res from bus_stops where bus_id = " + bus_id.ToString()));
        }

        //is tag in bus
        public static bool isTagInBus(int tag_id, int bus_id)
        {
            return Query<string>($"select 1 from bus_tags where bus_id = {bus_id} and tag_id = {tag_id}") != "";
        }

        //�ӱ�ǩ
        public static int AddTag2Bus(int bus_id, int tag_id)
        {
            return Exec("insert into bus_tags(bus_id,tag_id) values(" + bus_id.ToString() + "," + tag_id.ToString() + ")");
        }

        //remove bus tags
        public static int RemoveBusTag(int bus_id, int tag_id)
        {
            return Exec("delete from bus_tags  where bus_id = " + bus_id.ToString() + " and tag_id = " + tag_id.ToString());
        }

        //update bus tag name
        public static int UpdateBusTag(int tag_id, string tag_name)
        {
            return Exec("update bus_tag set tag_name = '" + tag_name + "' where tag_id = " + tag_id.ToString());
        }

        //�����·������Ϣ ���������� 
        public static string GetBusStopsA2(string bus_id)
        {
            return Query("select dbo.getBusStopsA2(" + bus_id + ") as res");
        }

        //�����·������Ϣ ����������
        public static string getBusStopsB2(string bus_id)
        {
            return Query("select dbo.getBusStopsB2(" + bus_id + ") as res");
        }

        //�����·��Ϣ���������кϲ�������������
        public static string getBusStopsC2(string bus_id)
        {
            return Query($"select dbo.getBusStopsC2({bus_id}) as res");
        }

        //�����·������Ϣ վ�㷴ת ����������
        public static string GetBusStopsB22(string bus_id) => Query($"select dbo.getBusStopsB22({bus_id}) as res");


        public static int updateStopBusMap(int stop_id)
        {
            return Exec($"update bus set update_map = 1 where bus_id in (select bus_id from bus_stops where stop_id = {stop_id})");
        }

        public static int UpdateBusMap(int bus_id)
        {
            return Exec($"update bus set update_map = 1 where bus_id  = {bus_id}");
        }

        public static string getBusMap(int bus_id)
        {
            string res = Query($"select bus_map from bus where bus_id = {bus_id} and update_map = 0");
            if (res == "")
            {
                res = GetBusMapAll(bus_id);
                updateBusMap(bus_id, res);
            }
            return res;
        }

        public static int updateBusMap(int bus_id, string bus_map)
        {
            return Exec($"update bus set bus_map ='{bus_map}', update_map = 0, update_map_date=getdate() where bus_id = {bus_id}");
        }

        #endregion ��·����
                
        
        #region �û��ύ��·

        //�����¾�����Ϣ
        public static int AddNewBusInfo(string bus_name, string start_stop, string end_stop, string bus_info, string price, string bus_stops_a, string bus_stops_b, string comment, string client_name, string link_info)
        {
            bus_name = bus_name.Replace("'", "");
            start_stop = start_stop.Replace("'", "");
            end_stop = end_stop.Replace("'", "");
            bus_info = bus_info.Replace("'", "");
            price = price.Replace("'", "");
            bus_stops_a = bus_stops_a.Replace("'", "");
            bus_stops_b = bus_stops_b.Replace("'", "");
            comment = comment.Replace("'", "");
            client_name = client_name.Replace("'", "");
            link_info = link_info.Replace("'", "");
            string client_id = "";
            string client_ip = "";

            string sql = $"insert into bus_new(bus_name,start_stop,end_stop,bus_info,price,bus_stopsa,bus_stopsb,comment,client_name,link_info,client_id,client_ip) values('{bus_name}','{start_stop}','{end_stop}','{bus_info}','{price}','{bus_stops_a}','{bus_stops_b}','{comment}','{client_name}','{link_info}',{client_id},'{client_ip}')";
            return Exec(sql);
        }

        public static int DelNewBus(string new_bus_id)
        {
            return Exec($"delete from bus_new where bus_id ={new_bus_id}");
        }

        #endregion �û��ύ��·

        #region ��·�۸��

        //��ʼ����·�۸��
        public static int UpdateBusPriceTable(string bus_id) => Exec($"exec sp_updatepricetable {bus_id}");

        //ȡ����·�۸���Ϣ
        public static string getBusPrice(string bus_id, int place_id, int place_id2)
        {
            return Query(string.Format("select dbo.getBusPrice2({0},{1},{2}) as res", bus_id, place_id, place_id2));
        }

        //ȡ����·�۸���Ϣ
        public static string getBusPrice(string bus_id)
        {
            return getBusPrice(bus_id, 0, 0);
        }

        //�õ���·�۸�վ��վ��
        public static string getBusStopPrice(string bus_id, string stop_id, string stop_id2)
            => Query($"select price from bus_price where bus_id = {bus_id} and ((stop_id = {stop_id} and stop_id2={stop_id2}) or (stop_id = {stop_id2} and stop_id2={stop_id}))");

        //������·�۸�վ��վ��
        public static int updateBusStopPrice(string bus_id, string stop_id, string stop_id2, string price)
        {
            return Exec($"update bus_price set price = {price} where bus_id = {bus_id} and ((stop_id = {stop_id} and stop_id2={stop_id2}) or (stop_id = {stop_id2} and stop_id2={stop_id}))");
        }

        #endregion ��·�۸��

        public static void GetBusInfo(string bus_id, ref string bus_order, ref string bus_name, ref string bus_name2, ref string bus_type, ref string start_stop, ref string end_stop,
            ref string price, ref string bus_price, ref string bus_price2, ref string bus_stops_a, ref string bus_stops_b, ref string bus_stops_c, ref string bus_info,
            ref string bus_pic, ref string pay_method, ref string sztong, ref string bus_price_type, ref string update_date)
        {
            SqlConnection myConnection = new SqlConnection(ConnString);
            if (!bus_id.IsNum()) bus_id = "0";
            string sql = $"select bus_id,bus_name,bus_name2,bus_order,bus_type,bus_info,bus_stops_a,bus_stops_b,bus_stops_c,start_stop,end_stop,price,pay_method,bus_price_type,sztong,bus_price,bus_price2,update_date,bus_pic from bus where bus_id = {bus_id}";
            SqlCommand myCommand = new(sql, myConnection);
            myConnection.Open();
            SqlDataReader reader = myCommand.ExecuteReader(CommandBehavior.CloseConnection);
            if (reader.Read())
            {
                int idCol = reader.GetOrdinal("bus_name");
                bus_name = reader.GetString(idCol);
                idCol = reader.GetOrdinal("bus_name2");
                if (!reader.IsDBNull(idCol))
                    bus_name2 = reader.GetString(idCol);
                else
                    bus_name2 = "";
                idCol = reader.GetOrdinal("bus_order");
                if (!reader.IsDBNull(idCol))
                    bus_order = reader.GetInt32(idCol).ToString();
                else
                    bus_order = "";
                idCol = reader.GetOrdinal("bus_type");
                if (!reader.IsDBNull(idCol))
                    bus_type = reader.GetInt32(idCol).ToString();
                else
                    bus_type = "";
                idCol = reader.GetOrdinal("bus_info");
                if (!reader.IsDBNull(idCol))
                    bus_info = reader.GetString(idCol);
                else
                    bus_info = "";
                idCol = reader.GetOrdinal("bus_stops_a");
                if (!reader.IsDBNull(idCol))
                    bus_stops_a = reader.GetString(idCol);
                else
                    bus_stops_a = "";
                idCol = reader.GetOrdinal("bus_stops_b");
                if (!reader.IsDBNull(idCol))
                    bus_stops_b = reader.GetString(idCol);
                else
                    bus_stops_b = "";
                idCol = reader.GetOrdinal("bus_stops_c");
                if (!reader.IsDBNull(idCol))
                    bus_stops_c = reader.GetString(idCol);
                else
                    bus_stops_c = "";
                idCol = reader.GetOrdinal("start_stop");
                if (!reader.IsDBNull(idCol))
                    start_stop = reader.GetString(idCol).Trim();
                else
                    start_stop = "";
                idCol = reader.GetOrdinal("end_stop");
                if (!reader.IsDBNull(idCol))
                    end_stop = reader.GetString(idCol).Trim();
                else
                    end_stop = "";
                idCol = reader.GetOrdinal("price");
                if (!reader.IsDBNull(idCol))
                    price = reader.GetString(idCol);
                else
                    price = "";
                idCol = reader.GetOrdinal("pay_method");
                if (!reader.IsDBNull(idCol))
                    pay_method = reader.GetInt32(idCol).ToString();
                else
                    pay_method = "0";
                idCol = reader.GetOrdinal("bus_price_type");
                if (!reader.IsDBNull(idCol))
                    bus_price_type = reader.GetInt32(idCol).ToString();
                else
                    bus_price_type = "0";
                idCol = reader.GetOrdinal("sztong");
                if (!reader.IsDBNull(idCol))
                    sztong = reader.GetBoolean(idCol).ToString();
                else
                    sztong = true.ToString();
                idCol = reader.GetOrdinal("bus_price");
                if (!reader.IsDBNull(idCol))
                    bus_price = reader.GetSqlMoney(idCol).ToString();
                else
                    bus_price = "";
                idCol = reader.GetOrdinal("bus_price2");
                if (!reader.IsDBNull(idCol))
                    bus_price2 = reader.GetSqlMoney(idCol).ToString();
                else
                    bus_price2 = "";
                idCol = reader.GetOrdinal("update_date");
                if (!reader.IsDBNull(idCol))
                    update_date = reader.GetDateTime(idCol).ToLongDateString();
                else
                    update_date = "";
                idCol = reader.GetOrdinal("bus_pic");
                if (!reader.IsDBNull(idCol))
                    bus_pic = reader.GetString(idCol);
                else
                    bus_pic = "";
                //bus_stops_a = BLM.bus.getEmTitle(bus_stops_a, bus_name);
                bus_stops_b = GetEmTitle(bus_stops_b, bus_name);
                bus_stops_c = GetEmTitle(bus_stops_c, bus_name);
            }
        }

        public static void GetBusInfoHis(string his_id, ref string bus_id, ref string bus_order, ref string bus_name, ref string bus_name2, ref string bus_type, ref string start_stop, ref string end_stop,
            ref string price, ref string bus_price, ref string bus_price2, ref string bus_stops_a, ref string bus_stops_b, ref string bus_stops_c, ref string bus_info,
            ref string bus_pic, ref string pay_method, ref string sztong, ref string bus_price_type, ref string update_date)
        {
            SqlConnection myConnection = new SqlConnection(ConnString);
            if (!his_id.IsNum()) his_id = "0";
            string sql = $"select bus_id,bus_name,bus_name2,bus_order,bus_type,bus_info,bus_stops_a,bus_stops_b,bus_stops_c,start_stop,end_stop,price,pay_method,bus_price_type,sztong,bus_price,bus_price2,update_date,bus_pic from bus_his where his_id = {his_id}";
            SqlCommand myCommand = new(sql, myConnection);
            myConnection.Open();
            SqlDataReader reader = myCommand.ExecuteReader(CommandBehavior.CloseConnection);
            if (reader.Read())
            {
                int idCol = reader.GetOrdinal("bus_name");
                bus_name = reader.GetString(idCol);
                idCol = reader.GetOrdinal("bus_id");
                if (!reader.IsDBNull(idCol))
                    bus_id = reader.GetInt32(idCol).ToString();
                else
                    bus_id = "";
                idCol = reader.GetOrdinal("bus_name2");
                if (!reader.IsDBNull(idCol))
                    bus_name2 = reader.GetString(idCol);
                else
                    bus_name2 = "";
                idCol = reader.GetOrdinal("bus_order");
                if (!reader.IsDBNull(idCol))
                    bus_order = reader.GetInt32(idCol).ToString();
                else
                    bus_order = "";
                idCol = reader.GetOrdinal("bus_type");
                if (!reader.IsDBNull(idCol))
                    bus_type = reader.GetInt32(idCol).ToString();
                else
                    bus_type = "";
                idCol = reader.GetOrdinal("bus_info");
                if (!reader.IsDBNull(idCol))
                    bus_info = reader.GetString(idCol);
                else
                    bus_info = "";
                idCol = reader.GetOrdinal("bus_stops_a");
                if (!reader.IsDBNull(idCol))
                    bus_stops_a = reader.GetString(idCol);
                else
                    bus_stops_a = "";
                idCol = reader.GetOrdinal("bus_stops_b");
                if (!reader.IsDBNull(idCol))
                    bus_stops_b = reader.GetString(idCol);
                else
                    bus_stops_b = "";
                idCol = reader.GetOrdinal("bus_stops_c");
                if (!reader.IsDBNull(idCol))
                    bus_stops_c = reader.GetString(idCol);
                else
                    bus_stops_c = "";
                idCol = reader.GetOrdinal("start_stop");
                if (!reader.IsDBNull(idCol))
                    start_stop = reader.GetString(idCol).Trim();
                else
                    start_stop = "";
                idCol = reader.GetOrdinal("end_stop");
                if (!reader.IsDBNull(idCol))
                    end_stop = reader.GetString(idCol).Trim();
                else
                    end_stop = "";
                idCol = reader.GetOrdinal("price");
                if (!reader.IsDBNull(idCol))
                    price = reader.GetString(idCol);
                else
                    price = "";
                idCol = reader.GetOrdinal("pay_method");
                if (!reader.IsDBNull(idCol))
                    pay_method = reader.GetInt32(idCol).ToString();
                else
                    pay_method = "";
                idCol = reader.GetOrdinal("bus_price_type");
                if (!reader.IsDBNull(idCol))
                    bus_price_type = reader.GetInt32(idCol).ToString();
                else
                    bus_price_type = "";
                idCol = reader.GetOrdinal("sztong");
                if (!reader.IsDBNull(idCol))
                    sztong = reader.GetBoolean(idCol).ToString();
                else
                    sztong = "";
                idCol = reader.GetOrdinal("bus_price");
                if (!reader.IsDBNull(idCol))
                    bus_price = reader.GetSqlMoney(idCol).ToString();
                else
                    bus_price = "";
                idCol = reader.GetOrdinal("bus_price2");
                if (!reader.IsDBNull(idCol))
                    bus_price2 = reader.GetSqlMoney(idCol).ToString();
                else
                    bus_price2 = "";
                idCol = reader.GetOrdinal("update_date");
                if (!reader.IsDBNull(idCol))
                    update_date = reader.GetDateTime(idCol).ToLongDateString();
                else
                    update_date = "";
                idCol = reader.GetOrdinal("bus_pic");
                if (!reader.IsDBNull(idCol))
                    bus_pic = reader.GetString(idCol);
                else
                    bus_pic = "";
            }
        }


        public static string GetBusMapAll(int bus_id)
        {
            return $"kresult+=\"mapa|mapb|\";var mapa= new Array(\"ȥ��\",{GetBusMapA(bus_id)});var mapb= new Array(\"�س�\",{GetBusMapB(bus_id)});";
        }

        //�����ݳ��ȳ���8000������ʹ��SQL������洢���̴���
        public static string GetBusMapA(int bus_id)
        {
            SqlConnection MyConn = new(ConnString);
            string strSQL = $"select stop_id,stop_order from bus_stops where bus_id = {bus_id} and stop_order is not null order by stop_order";
            SqlCommand MyComm = new(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string res = "";
                string stop_buses = "";
                int stop_id;
                int stop_order;
                while (reader.Read())
                {
                    stop_id = Convert.ToInt32(reader[0]);
                    stop_buses = Query($"select dbo.getStopBus({stop_id})");
                    if (stop_buses != "")
                    {
                        stop_buses = "\"" + stop_buses + "\"";
                        if (!reader.IsDBNull(1))
                        {
                            stop_order = Convert.ToInt32(reader[1]);
                            if (stop_order > 1 & res != "")
                                res = res + ",";
                            res += stop_buses;
                        }
                    }
                }
                reader.Close();
                return res;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }
        }

        public static string GetBusMapB(int bus_id)
        {
            SqlConnection MyConn = new(ConnString);
            string strSQL = $"select stop_id,stop_order2 from bus_stops where bus_id = {bus_id} and stop_order2 is not null order by stop_order2";
            SqlCommand MyComm = new(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string res = "";
                string stop_buses = "";
                int stop_id;
                int stop_order;
                while (reader.Read())
                {
                    stop_id = Convert.ToInt32(reader[0]);
                    stop_buses = Query($"select dbo.getStopBus({stop_id})");
                    if (stop_buses != "")
                    {
                        stop_buses = "\"" + stop_buses + "\"";
                        if (!reader.IsDBNull(1))
                        {
                            stop_order = Convert.ToInt32(reader[1]);
                            if (stop_order > 1 & res != "")
                                res = res + ",";
                            res += stop_buses;
                        }
                    }
                }
                reader.Close();
                return res;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }
        }

        public static string GetPlaceBuses(string place_id)
        {
            SqlConnection MyConn = new(ConnString);
            string strSQL = $"select bus_id,bus_name from bus where bus_id in (select bus_id from bus_stops where stop_id in (select stop_id from place_stops where place_id = {place_id})) order by bus_order";
            SqlCommand MyComm = new SqlCommand(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string buses = "";
                while (reader.Read())
                {
                    buses += $"{GetBusName($"{reader[1]}")}  ";
                }

                // Call Close when done reading.
                reader.Close();

                return buses;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }
        }

        public static string GetPlaceBuses(string place_id, string place_id2)
        {
            SqlConnection MyConn = new SqlConnection(ConnString);
            string strSQL = "select bus_id,bus_name from bus where exists (select 1 from bus_stops where stop_id in (select stop_id from place_stops where place_id = " + place_id + ") and bus_id = bus.bus_id) and exists (select 1 from bus_stops where stop_id in (select stop_id from place_stops where place_id = " + place_id2 + ") and bus_id = bus.bus_id) order by bus_order,bus_name";
            SqlCommand MyComm = new SqlCommand(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string buses = "";// String.Format("����{0}��", reader.RecordsAffected);
                while (reader.Read())
                {
                    buses += GetBusName($"{reader[1]}") + "  ";
                }

                // Call Close when done reading.
                reader.Close();

                return buses;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }

        }

        public static string getStopBuses(string stop_id)
        {
            SqlConnection MyConn = new SqlConnection(ConnString);
            string strSQL = $"select bus_id,bus_name from bus where bus_type = 1 and bus_id in (select bus_id from bus_stops where stop_id = {stop_id}) order by bus_order";
            SqlCommand MyComm = new SqlCommand(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string buses = "";
                while (reader.Read())
                {
                    buses += GetBusName($"{reader[1]}") + "  ";
                }

                // Call Close when done reading.
                reader.Close();

                return buses;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }
        }

        public static string getStopBuses(string stop_id, string stop_id2, string s_split)
        {
            SqlConnection MyConn = new SqlConnection(ConnString);
            string strSQL = "select bus_id,bus_name from bus where bus_id in (select bus_id from bus_stops where stop_id = " + stop_id + ") and bus_id in (select bus_id from bus_stops where stop_id = " + stop_id2 + ") order by bus_order";
            SqlCommand MyComm = new SqlCommand(strSQL, MyConn);
            try
            {
                MyConn.Open();
                SqlDataReader reader = MyComm.ExecuteReader();
                string buses = "";
                int i = 1;
                while (reader.Read())
                {
                    if (i > 1)
                        buses += s_split;
                    i++;
                    buses += string.Format("{1}", reader[0], reader[1]);
                }

                // Call Close when done reading.
                reader.Close();

                return buses;
            }
            catch (Exception e)
            {
                throw new Exception(e.Message);
            }
            finally
            {
                MyComm.Dispose();
                MyConn.Close();
                MyConn.Dispose();
            }

        }

    }

    #region ��·/վ������
    public class BusComment : MetaData<BusComment>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus_comment";
        public override string KeyField => "comment_id";

        //ɾ����·����
        public static int DelComment(string commentId)
        {
            return Delete(commentId);
        }

        //������·����
        public static int AddComment(string busId, string info, string clientId, string clientIP)
        {
            return Insert([
                new Cov("bus_id", busId),
                new Cov("comment_info", info),
                new Cov("client_id", clientId),
                new Cov("ip", clientIP),
            ]);
        }
    }
    #endregion ��·����
}
