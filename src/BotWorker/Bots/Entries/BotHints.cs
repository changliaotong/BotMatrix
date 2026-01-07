using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class BotHints : MetaData<BotHints>
    {
        public override string TableName => "Hints";
        public override string KeyField => "Id";

        /*
        每个小号机器人可以自定义提示语？ （后台设置）
        每个群只能在已定义的提示语列表中选择？（后台选择）
        没有数据的从官方号码获得（1098299491）
        官方号也没数据的取代码中的默认值。

        风格 每条数据都可以定义风格，相当于分组，供选择时参考。可以通过风格批量选择提示语
         */

        // 获得提示语 有多条记录的随机一条
        public static string GetHint(long qqRobot, string hintCode)
        {
            long hints_id = GetWhere("Id", $"UserId = {qqRobot} and HintCode = {hintCode}", "newid()").AsLong();
            if (hints_id == 0)
            {
                hints_id = GetWhere("Id", $"UserId = {BotInfo.BotUinDef} and HintCode = {hintCode}", "newid()").AsLong();
            }
            PlusTimes(hints_id);
            return GetValue("HintInfo", hints_id);
        }

        // 增加提示语
        public static int Append(long botUin, long qq, string hintCode, string hintInfo)
        {
            return Insert([
                new Cov("BotUin", botUin),
                new Cov("UserId", qq),
                new Cov("HintCode", hintCode),
                new Cov("HintInfo", hintInfo),
            ]);
        }

        // 更新提示语
        public static int UpdateHint(long hintsId, long qq, string hintInfo)
        {
            return Update([
                new Cov("HintInfo", hintInfo),
                new Cov("UpdateDate", DateTime.MinValue),
                new Cov("Updateby", qq),], hintsId);
        }

        // 更新提示语使用次数
        public static int PlusTimes(long hintsId)
        {
            return Plus("UsedTimes", 1, hintsId);
        }
    }
}
