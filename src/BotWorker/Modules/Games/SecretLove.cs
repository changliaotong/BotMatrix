
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    class SecretLove : MetaData<SecretLove>
    {

        public override string TableName => "Love";
        public override string KeyField => "UserId";
        public override string KeyField2 => "LoveId";

        public static string GetLoveStatus()
            => GetLoveStatusAsync().GetAwaiter().GetResult();

        public static async Task<string> GetLoveStatusAsync()
        {
            string sql = $"SELECT COUNT(DISTINCT UserId), COUNT(LoveId) FROM {FullName}";
            return await QueryResAsync(sql, "已有{0}人登记暗恋对象{1}个。");
        }

        public static int Append(long botUin, long groupId, long qq, long loveQQ)
            => AppendAsync(botUin, groupId, qq, loveQQ).GetAwaiter().GetResult();

        public static async Task<int> AppendAsync(long botUin, long groupId, long qq, long loveQQ)
        {
            return await InsertAsync([
                            new Cov("UserId", qq),
                            new Cov("LoveId", loveQQ),
                            new Cov("GroupId", groupId),
                            new Cov("BotUin", botUin)
                        ]);
        }

        public static long GetCountLoveMe(long userId)
            => GetCountLoveMeAsync(userId).GetAwaiter().GetResult();

        public static async Task<long> GetCountLoveMeAsync(long userId)
        {
            return await CountWhereAsync($"LoveId={userId}");
        }

        public static long GetCountLove(long userId)
            => GetCountLoveAsync(userId).GetAwaiter().GetResult();

        public static async Task<long> GetCountLoveAsync(long userId)
        {
            return await CountWhereAsync($"UserId={userId}");
        }

        public static bool IsLoveEachother(long userId, long loveId)
            => IsLoveEachotherAsync(userId, loveId).GetAwaiter().GetResult();

        public static async Task<bool> IsLoveEachotherAsync(long userId, long loveId)
        {
            return await ExistsAsync(userId, loveId) && await ExistsAsync(loveId, userId);
        }
    }

}
