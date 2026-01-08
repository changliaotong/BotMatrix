using sz84.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Groups;

namespace BotWorker.Modules.Games
{
    internal class Fishing : MetaData<Fishing>
    {
        public override string TableName => "GroupMember";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "UserId";

        public static List<string> fish_fields = ["yugan", "yugou", "yuer", "yuxian", "shuigui", "jingyu", "zhangyu", "huangyu", "qingwa", "beike", "neiyi", "poxie"];
        public static List<string> fish_names = ["鱼竿", "鱼钩", "鱼饵", "鱼线", "水鬼", "鲸鱼", "章鱼", "黄鱼", "青蛙", "贝壳", "内衣", "破鞋"];
        
        //fish state  0 鱼竿在手上 1 鱼竿在水里

        //todo 钓鱼命令 没有鱼竿鱼线鱼钩的自动购买，积分不足的提示不足。简化游戏流程。

        /// 购买渔具
        public static string GetBuyTools(long botUin,long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2)
        {
            int count = cmdPara2.AsInt();
            int coins_type = (int)CoinsLog.CoinsType.purpleCoins;
            long minus_coins = 100 * count;

            long coins_value = GroupMember.GetCoins(coins_type, groupId, qq);
            if (coins_value < minus_coins)
                return string.Format($"您的紫币{coins_value}不足{minus_coins}");
            //扣除紫币 记录扣币记录 更新数量
            var sql = GroupMember.SqlPlus(CoinsLog.conisFields[coins_type], minus_coins, groupId, qq);
            var sql2 = CoinsLog.SqlCoins(botUin, groupId, groupName, qq, name, coins_type, -minus_coins, ref coins_value, $"购买渔具：{cmdPara}*{cmdPara2}");
            var sql3 = GroupMember.SqlPlus(fish_fields[fish_names.IndexOf(cmdPara)], count, groupId, qq);
            int i = ExecTrans(sql, sql2, sql3);
            string res;
            if (i == -1)
                res = RetryMsg;
            else
                res = $"✅ 购买渔具：{cmdPara}*{cmdPara2}\n紫币：-{minus_coins}，累计：{coins_value}";
            return res;
        }

        // 钓鱼
        public static string GetFishing(long groupId, string groupName, long userId, string name, string cmdName, string cmdPara)
        {
            string res = "";
            if (res != "")
                return res;
            if (cmdName == "钓鱼")
            {
                return "✅ 开始钓鱼，请 抛竿";
            }
            else if (cmdName == "抛竿")
            {
                Update("FishDate=GETDATE(), FishState = 1", groupId, userId);
                return "✅ 抛竿成功，请5分钟后收竿";
            }
            else if (cmdName == "收竿")
            {
                int s = GroupMember.GetInt("FishState", groupId, userId);
                if (s == 0)
                    return "请先 抛竿";
                GroupMember.SetValue("FishState", 0, groupId, userId);
                int fishTime = GroupMember.GetInt("ABS(DATEDIFF(MINUTE, GETDATE(), FishDate))", groupId, userId);
                if (fishTime < 5)
                    return $"很遗憾，什么都没钓到，重新【抛竿】吧";
                else
                {
                    int i = RandomInt(4, 18);
                    if (i > 11)
                        return "很遗憾，什么都没钓到";
                    else
                    {
                        GroupMember.Plus(fish_fields[i], 1, groupId, userId);
                        return string.Format("✅ 恭喜你，钓到{0}！", fish_names[i]);
                    }

                }
            }
            return res;
        }
    }
}
