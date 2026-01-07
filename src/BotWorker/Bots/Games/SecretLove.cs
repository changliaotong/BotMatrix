using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Games
{
    class SecretLove : MetaData<SecretLove>
    {

        public override string TableName => "Love";
        public override string KeyField => "UserId";
        public override string KeyField2 => "LoveId";

        public static string GetLoveStatus()
        {
            string sql = $"SELECT COUNT(DISTINCT UserId), COUNT(LoveId) FROM {FullName}";
            return QueryRes(sql, "已有{0}人登记暗恋对象{1}个。");
        }

        public static int Append(long botUin, long groupId, long qq, long loveQQ)
        {
            return Insert([
                            new Cov("UserId", qq),
                            new Cov("LoveId", loveQQ),
                            new Cov("GroupId", groupId),
                            new Cov("BotUin", botUin)
                        ]);
        }

        public static long GetCountLoveMe(long userId)
        {
            return CountWhere($"LoveId={userId}");
        }

        public static long GetCountLove(long userId)
        {
            return CountWhere($"UserId={userId}");
        }

        public static bool IsLoveEachother(long userId, long loveId)
        {
            return Exists(userId, loveId) && Exists(loveId, userId);
        }
    }

}
