using BotWorker.Core;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Models.Office
{
    public class Price : MetaData<Price>
    {
        public override string TableName => "Price";
        public override string KeyField => "Id";

        // 取得机器人续费价格
        public static decimal GetRobotPrice(long month)
        {
            if (month > 60)
                month = 60;
            return GetWhere($"price", $"month = {month}").AsDecimal();
        }
    }
}
