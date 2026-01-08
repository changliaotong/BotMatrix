using BotWorker.common.Exts;
using sz84.Core;
using BotWorker.Domain.MetaDatas;

namespace sz84.Bots.Models.Buses
{
    public class BusTag : MetaData<BusTag>
    {
        public override string DataBase => "sz84";
        public override string TableName => "bus_tag";
        public override string KeyField => "tag_id";

        public static string GetId(string tag_name)
        {
            return GetWhere($"tag_id", $"tag_name = {tag_name.Quotes()}");
        }

        public static int AddTag(string tag_name, int client_id, string client_ip)
        {
            return Insert([
                new Cov("tag_name", tag_name),
                new Cov("insert_by", client_id),
                new Cov("client_ip", client_ip),
            ]);
        }

        public static long AddTag(string tag_name)
        {
            return AddTag(tag_name, 1, "");
        }

        public static string GetTagName(string tag_id)
        {
            return GetValue("tag_name", tag_id);
        }
    }
}
