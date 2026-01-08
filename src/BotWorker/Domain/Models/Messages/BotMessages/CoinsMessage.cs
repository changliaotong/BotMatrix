namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public string GetCoinsListAll(long qq, long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsNapCat) ? "{i} [@:{0}]：{1}\n" : "{i} {0} {1}\n";
            string res = QueryRes($"select top {top} Id, coins from {UserInfo.FullName} order by coins desc", format);
            if (!res.Contains(qq.ToString()))
                res += $"{{金币总排名}} {qq}：{{金币}}\n";
            return res;
        }

        public string GetCoinsList(long top = 10)
        {
            var format = !IsRealProxy && (IsMirai || IsNapCat) ? "第{i}名[@:{0}] 💰{1:N0}\n" : "第{i}名{0} 💰{1:N0}\n";
            string res = UserInfo.QueryWhere($"top {top} Id, coins", $"Id in (select UserId from {CoinsLog.FullName} where GroupId = {GroupId})",
                                 $"coins desc", format);
            if (!res.Contains(UserId.ToString()))
                res += $"{{金币排名}} [@:{UserId}] 💰{{金币}}\n";
            res = ReplaceRankWithIcon(res);
            return $"🏆 金币排行榜\n{res}";
        }

        public static long GetCoinsRanking(long groupId, long qq)
        {
           var coins = UserInfo.GetCoins(qq);
           return UserInfo.CountWhere($"coins > {coins} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1;
        }

        public static long GetCoinsRankingAll(long qq)
        {
            return UserInfo.CountWhere($"Coins > {UserInfo.GetCoins(qq)}") + 1;
        }
}
