using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

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
        public static async Task<string> GetBuyToolsAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2)
        {
            int count = cmdPara2.AsInt();
            int coins_type = (int)CoinsLog.CoinsType.purpleCoins;
            long minus_coins = 100 * count;

            long coins_value = await GroupMember.GetCoinsAsync(coins_type, groupId, qq);
            if (coins_value < minus_coins)
                return string.Format($"您的紫币{coins_value}不足{minus_coins}");

            //扣除紫币 记录扣币记录 更新数量
            var sql1 = GroupMember.SqlPlus(CoinsLog.conisFields[coins_type], -minus_coins, groupId, qq);
            var (sql2_str, sql2_paras, new_coins_value) = await CoinsLog.SqlCoinsAsync(botUin, groupId, groupName, qq, name, coins_type, -minus_coins, $"购买渔具：{cmdPara}*{cmdPara2}");
            var sql3 = GroupMember.SqlPlus(fish_fields[fish_names.IndexOf(cmdPara)], count, groupId, qq);

            using var trans = await BeginTransactionAsync();
            try
            {
                var (s1, p1) = sql1;
                await ExecAsync(s1, trans, p1);

                await ExecAsync(sql2_str, trans, sql2_paras);

                var (s3, p3) = sql3;
                await ExecAsync(s3, trans, p3);

                await trans.CommitAsync();

                return $"✅ 购买渔具：{cmdPara}*{cmdPara2}\n紫币：-{minus_coins}，累计：{new_coins_value}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"Fishing.GetBuyToolsAsync error: {ex.Message}");
                return RetryMsg;
            }
        }

        /// 购买渔具
        public static string GetBuyTools(long botUin, long groupId, string groupName, long qq, string name, string cmdName, string cmdPara, string cmdPara2)
        {
            return GetBuyToolsAsync(botUin, groupId, groupName, qq, name, cmdName, cmdPara, cmdPara2).GetAwaiter().GetResult();
        }

        // 钓鱼
        public static async Task<string> GetFishingAsync(long groupId, string groupName, long userId, string name, string cmdName, string cmdPara)
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
                await UpdateAsync("FishDate=GETDATE(), FishState = 1", groupId, userId);
                return "✅ 抛竿成功，请5分钟后收竿";
            }
            else if (cmdName == "收竿")
            {
                int s = await GroupMember.GetIntAsync("FishState", groupId, userId);
                if (s == 0)
                    return "请先 抛竿";
                await GroupMember.SetValueAsync("FishState", 0, groupId, userId);
                int fishTime = await GroupMember.GetIntAsync("ABS(DATEDIFF(MINUTE, GETDATE(), FishDate))", groupId, userId);
                if (fishTime < 5)
                    return $"很遗憾，什么都没钓到，重新【抛竿】吧";
                else
                {
                    int i = RandomInt(4, 18);
                    if (i > 11)
                        return "很遗憾，什么都没钓到";
                    else
                    {
                        await GroupMember.PlusAsync(fish_fields[i], 1, groupId, userId);
                        return string.Format("✅ 恭喜你，钓到{0}！", fish_names[i]);
                    }

                }
            }
            return res;
        }

        public static string GetFishing(long groupId, string groupName, long userId, string name, string cmdName, string cmdPara)
        {
            return GetFishingAsync(groupId, groupName, userId, name, cmdName, cmdPara).GetAwaiter().GetResult();
        }
    }
}
