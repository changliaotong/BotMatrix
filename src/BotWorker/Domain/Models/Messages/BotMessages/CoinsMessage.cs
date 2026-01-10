namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task<string> GetCoinsListAllAsync(long qq, long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
            string res = await QueryResAsync($"select top {top} Id, coins from {UserInfo.FullName} order by coins desc", format);
            if (!res.Contains(qq.ToString()))
                res += $"{{金币总排名}} {qq}：{{金币}}\n";
            return res;
        }

        public async Task<string> GetCoinsListAsync(long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsQQ) ? "第{i}名[@:{0}] 💰{1:N0}\n" : "第{i}名{0} 💰{1:N0}\n";
            string res = await UserInfo.QueryResAsync($"select top {top} Id, coins from {UserInfo.FullName} where Id in (select UserId from {CoinsLog.FullName} where GroupId = {GroupId}) order by coins desc", format);
            if (!res.Contains(UserId.ToString()))
                res += $"{{金币排名}} [@:{UserId}] 💰{{金币}}\n";
            res = ReplaceRankWithIcon(res);
            return $"🏆 金币排行榜\n{res}";
        }

        public static async Task<long> GetCoinsRankingAsync(long groupId, long qq)
        {
           var coins = await UserInfo.GetCoinsAsync(qq);
           return await UserInfo.CountWhereAsync($"coins > {coins} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1;
        }

        public static async Task<long> GetCoinsRankingAllAsync(long qq)
        {
            return await UserInfo.CountWhereAsync($"Coins > {await UserInfo.GetCoinsAsync(qq)}") + 1;
        }
}
