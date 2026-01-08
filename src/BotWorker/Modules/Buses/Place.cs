using Microsoft.Data.SqlClient;
using System.Data;

/// <summary>
///Class1 的摘要说明
/// </summary>
/// 
namespace BotWorker.Modules.Buses
{
    public class PlaceStops : MetaData<PlaceStops>
    {
        public override string DataBase => "sz84";
        public override string TableName => "place_stops";
        public override string KeyField => "place_id";
        public override string KeyField2 => "stop_id";

        //添加目的地附近的公交站
        public static int AddStop(int placeId, int stopId, string info, long clientId, string clientIP)
        {
            return Insert([
                new Cov("place_id", placeId),
                new Cov("stop_id", stopId),
                new Cov("info", info),
                new Cov("client_id", clientId),
                new Cov("client_ip", clientIP),
            ]);
        }

        //去掉附近的公交站
        public static int RemoveStop(string placeId, string StopId)
        {
            return Delete(placeId, StopId);
        }

        //复制目的地经过的站点
        public static int CopyStopsFrom(string place_id, string place_id2)
        {
            return Exec($"insert into place_stops(place_id,stop_id) select {place_id},stop_id from place_stops where place_id = {place_id2} and stop_id not in (select stop_id from place_stops where place_id = {place_id})");
        }
    }

    public class PlaceTags : MetaData<PlaceTags>
    {
        public override string DataBase => "sz84";
        public override string TableName => "place_tags";
        public override string KeyField => "place_id";
        public override string KeyField2 => "tag_id";

        //为目的地添加标签
        public static long AddTag(long tagId, long placeId, int clientId, string clientIP)
        {
            return Insert([
                new Cov("tag_id", tagId),
                new Cov("place_id", placeId),
                new Cov("client_id", clientId),
                new Cov("client_ip", clientIP),
            ]);
        }

        //移除目的地标签
        public static long RemoveTag(long placeId, long tagId)
        {
            return Delete(placeId, tagId);
        }
    }

    public class Place : MetaData<Place>
    {
        public override string DataBase => "sz84";
        public override string TableName => "place";
        public override string KeyField => "place_id";

        public static bool IsOtherCity(string place_id)
        {
            return (QueryScalar<string>($"select top 1 1 from place_stops where place_id = {place_id} and stop_id in (select stop_id from bus_stops where bus_id in (select bus_id from bus_tags where tag_id = 16)) and stop_id not in (select stop_id from bus_stops where bus_id not in (select bus_id from bus_tags where tag_id = 16))") ?? "") != "";
        }

        public static string GetPlaceStops2(string place_id)
        {
            return QueryScalar<string>($"select dbo.getPlaceStops2({place_id}) as res") ?? "";
        }

        //得到更短的地名 去掉地名中的区和路名信息 递归调用
        public static string GetShortPlaceName(string key, string keyword)
        {
            //地名
            int place_id;
            string place_name = key;

            place_id = GetPlaceID(place_name);
            if (place_id != 0)
                return place_name;

            int keylength = key.Length;
            if (key.Length > 4)
                key = key.Replace("广东省", "").Replace("深圳市", "").Replace("罗湖区", "").Replace("福田区", "").Replace("南山区", "").Replace("宝安区", "").Replace("龙岗区", "").Replace("盐田区", "").Replace("龙华新区", "").Replace("龙华区", "").Replace("光明新区", "").Replace("光明区", "").Replace("大鹏新区", "").Replace("大鹏区", "");
            if (key.Length > 3)
                key = key.Replace("广东", "").Replace("深圳", "").Replace("罗湖", "").Replace("福田", "").Replace("南山", "").Replace("宝安", "").Replace("龙岗", "").Replace("盐田", "").Replace("龙华", "").Replace("光明", "").Replace("大鹏", "");

            if (key.Length < keylength && key.Length > 0)
            {
                return GetShortPlaceName(key, keyword);
            }
            else
                place_name = keyword;

            if (place_name.Substring(place_name.Length - 1, 1) == "站")
            {
                string tstr = place_name.Substring(0, place_name.Length - 1);
                if (GetPlaceId(tstr) != 0)
                    place_name = tstr;
            }
            return place_name;
        }

        public static string GetShortPlaceName(string key)
        {
            return GetShortPlaceName(key, key);
        }

        //取得与目的地同名的站台的编号
        public static string GetSameNameStop(string place_id)
        {
            return QueryScalar<string>("select stop_id from bus_stop a inner join place b on a.stop_name = b.place_name where place_id = " + place_id) ?? "";
        }

        //是否设定附近站台信息 返回站台数
        public static string IsHaveStop(string place_id)
        {
            return QueryScalar<string>("select count(stop_id) from place_stops where place_id = " + place_id) ?? "0";
        }

        //由 place_name 得到  place_id （递归）
        public static int GetPlaceID(string place_name)
        {
            if (place_name == null) return 0;
            string res = QueryScalar<string>(string.Format("select dbo.GetMasterPlace(place_id) from place where place_name = '{0}'", place_name)) ?? "";
            if (res == "")
                return 0;
            else
                return Convert.ToInt32(res);
        }

        public static string GetPlaceName(int placeId)
        {
            return GetPlaceName(placeId.ToString());
        }


        //通过ID取得目的地名称
        public static string GetPlaceName(string placeId)
        {
            return Get<string>("place_name", placeId);
        }

        //由 place_id 获得目的地名称 （递归）
        public static long GetMasterPlace(long placeId)
        {
            return QueryScalar<long>("select dbo.getMasterPlace(" + placeId.ToString() + ") as res");            
        }

        public static long GetMasterID(long placeId)
        {
            return GetMasterID(placeId.ToString()).AsLong();
        }

        //由 place_id 得到上级 place_id
        public static string GetMasterID(string place_id)
        {
            string res = QueryScalar<string>("select master_id from place where place_id = " + place_id);
            if (res == "")
                return "0";
            else
                return res;
        }

        //增加新的目的地
        public static long AddPlace(string placeName, string placeInfo, string clientId, string clientIP)
        {
            return Insert([
                new Cov("place_name", placeName),
                new Cov("place_py", QueryScalar<string>($"select dbo.getPY({placeName.Quotes()})")),
                new Cov("place_info", placeInfo),
                new Cov("client_id", clientId),
                new Cov("client_ip", clientIP),
            ]);
        }

        //目的地是否存在,返回ID,不存在返回0
        public static long GetPlaceId(string place_name)
        {
            return GetWhere("place_id", $"place_name = {place_name.Quotes()}").AsLong();            
        }

        //设置目的地跳转 url
        public static int SetPlaceURL(int placeId, string URL)
        {
            return Update($"place_url = {URL.Quotes()}", placeId);            
        }


        //设置目的地别名
        public static long SetAlias(long placeId, long masterPlaceId)
        {
            if (placeId != masterPlaceId)
            {
                //需要解决两个问题：1.避免循环设置别名 2.别名的别名
                string sql = "update place set master_id = ";
                if (masterPlaceId == 0)
                    sql += "null";
                else
                {
                    if (GetMasterID(masterPlaceId) != 0)
                        masterPlaceId = GetMasterPlace(masterPlaceId);
                    sql += masterPlaceId.ToString();
                }
                sql += " where place_id = " + placeId;
                if (placeId == masterPlaceId)
                    return -2;
                return Exec(sql);
            }
            return 0;
        }

        public static int GetStopCountByPlace(string place_id)
        {
            return Convert.ToInt32(QueryScalar<string>("select count(stop_id) as res from place_stops where place_id = " + place_id));
        }

        public static string GetPlaceUrl(string place_id)
        {
            return QueryScalar<string>($"select place_url from place where place_id = {place_id}") ?? "";
        }

        //删除目的地
        public static async Task<int> DeleteAsync(string place_id)
        {
            using var trans = await BeginTransactionAsync();
            try
            {
                await ExecAsync($"delete from bus_search where keyword = (select place_name from place where place_id = {place_id})", trans);
                await ExecAsync($"delete from place_stops where place_id = {place_id}", trans);
                await ExecAsync($"delete from place_tags where place_id = {place_id}", trans);
                await ExecAsync($"update place set master_id = null where master_id = {place_id}", trans);
                await ExecAsync($"delete from place where place_id = {place_id}", trans);
                await trans.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Place.DeleteAsync error: {ex.Message}");
                await trans.RollbackAsync();
                return -1;
            }
        }

        //删除目的地
        public static int Delete(string place_id)
        {
            return DeleteAsync(place_id).GetAwaiter().GetResult();
        }

        //更新目的地名称
        public static int UpdatePlace(string placdId, string placeName, string placeInfo)
        {
            return Update($"update_date=getdate(), place_name={placeName.Quotes()}, place_info={placeInfo.Quotes()}", placdId);
        }


        public static void GetPlaceInfo(string place_id, ref string place_name, ref string place_info, ref string place_url, ref int master_id)
        {
            SqlConnection myConnection = new(ConnString);
            if (!place_id.IsNum()) place_id = "0";
            string sql = $"select * from place where place_id = {place_id}";
            SqlCommand myCommand = new(sql, myConnection);
            myConnection.Open();
            SqlDataReader reader = myCommand.ExecuteReader(CommandBehavior.CloseConnection);
            if (reader.Read())
            {
                int idCol = reader.GetOrdinal("place_name");
                place_name = reader.GetString(idCol);
                idCol = reader.GetOrdinal("place_info");
                if (!reader.IsDBNull(idCol))
                    place_info = reader.GetString(idCol);
                else
                    place_info = "";
                idCol = reader.GetOrdinal("place_url");
                if (!reader.IsDBNull(idCol))
                    place_url = reader.GetString(idCol);
                else
                    place_url = "";
                idCol = reader.GetOrdinal("master_id");
                if (!reader.IsDBNull(idCol))
                    master_id = reader.GetInt32(idCol);
                else
                    master_id = 0;
            }
        }


        //取得经过目的地附近的所有线路名称
        public static string GetPlaceBuses(string place_id)
        {
            return Bus.GetPlaceBuses(place_id);
        }

        //取得经过A目的地到B目的地的所有线路名称
        public static string GetPlaceBuses(string place_id, string place_id2)
        {
            return Bus.GetPlaceBuses(place_id, place_id2);
        }

    }
}