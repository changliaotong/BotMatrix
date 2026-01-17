namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<string> GetCoinsListAllAsync(long qq, long top = 10)
        {
            return await UserService.GetCoinsListAllAsync(qq, top);
        }

        public async Task<string> GetCoinsListAsync(long top = 10)
        {
            string res = await UserService.GetCoinsListAsync(GroupId, UserId, top);
            res = ReplaceRankWithIcon(res);
            return $"🏆 金币排行榜\n{res}";
        }

        public async Task<long> GetCoinsRankingAsync(long groupId, long qq)
        {
           return await UserService.GetCoinsRankingAsync(groupId, qq);
        }

        public async Task<long> GetCoinsRankingAllAsync(long qq)
        {
            return await UserService.GetCoinsRankingAllAsync(qq);
        }
}
