using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Games
{
    public class Jielong : MetaData<Jielong>
    {
        public override string TableName => "Chengyu";
        public override string KeyField => "Id";


        // 为机器人获取一个随机接龙结果
        public static async Task<string> GetJielongAsync(long groupId, long UserId, string currCy)
        {
            string pinyin = Chengyu.PinYinLast(currCy);
            string sql = $"SELECT TOP 1 chengyu FROM {Chengyu.FullName} " +
                           $"WHERE pinyin LIKE '{pinyin} %' AND chengyu NOT IN " +
                           $"(SELECT chengyu FROM {FullName} WHERE GroupId = {groupId} AND UserId = {UserId} " +
                           $"AND Id > (SELECT TOP 1 Id FROM {FullName} WHERE GroupId = {groupId} " +
                           $"AND UserId = {UserId} AND GameNo = 1 ORDER BY InsertDate DESC)) " +
                           $"ORDER BY NEWID()";

            return await QueryAsync(sql);
        }

        public static string GetJielong(long groupId, long UserId, string currCy)
        {
            return GetJielongAsync(groupId, UserId, currCy).GetAwaiter().GetResult();
        }

        // 接龙游戏最大ID
        public static async Task<int> GetMaxIdAsync(long groupId)
        {
            var res = await QueryAsync($"SELECT MAX(Id) FROM {FullName} WHERE GroupId = {groupId} AND GameNo = 1");
            return res.AsInt();
        }

        public static int GetMaxId(long groupId)
        {
            return GetMaxIdAsync(groupId).GetAwaiter().GetResult();
        }

        // 接龙成功数量
        public static async Task<string> GetGameCountAsync(long groupId, long qq)
        {
            return await QueryAsync($"SELECT {DbName}.DBO.[getChengyuGameCount]({groupId},{qq})");
        }

        public static string GetGameCount(long groupId, long qq)
        {
            return GetGameCountAsync(groupId, qq).GetAwaiter().GetResult();
        }

        // 接龙加分总数
        public static async Task<long> GetCreditAddAsync(long userId)
        {
            string query = $"SELECT ISNULL(SUM(CreditAdd), 0) FROM {CreditLog.FullName} " +
                           $"WHERE UserId = {userId} AND CreditInfo = '成语接龙' " +
                           $"AND ABS(DATEDIFF(DAY, InsertDate, GETDATE())) < 1";

            var res = await QueryAsync(query);
            return res.AsLong();
        }

        public static long GetCreditAdd(long userId)
        {
            return GetCreditAddAsync(userId).GetAwaiter().GetResult();
        }

        // 成语接龙加分
        public static async Task<string> AddCreditAsync(BotMessage bm)
        {
            //接龙加分，接龙自己的不加分，答错扣分
            var creditAdd = 10;
            string res = "";
            if ((bm.IsGuild || await GetCreditAddAsync(bm.UserId) < 2000) && bm.Group.IsCreditSystem)
            {
                (int i, long creditValue) = await bm.AddCreditAsync(creditAdd, "成语接龙");
                if (i != -1)
                    res = $"\n💎 积分：+{creditAdd}，累计：{creditValue:N0}";
            }
            return res;
        }

        public static string AddCredit(BotMessage bm)
        {
            return AddCreditAsync(bm).GetAwaiter().GetResult();
        }

        // 成语接龙扣分
        public static async Task<string> MinusCreditAsync(BotMessage bm)
        {
            if (bm.IsGuild || bm.IsRealProxy) return "";

            string res = "";

            var creditAdd = 10;
            int c_chengyu = await GetCountAsync(bm.RealGroupId, bm.UserId);
            if (c_chengyu > 0 && bm.Group.IsCreditSystem)
            {
                (int i, long creditValue) = await bm.MinusCreditAsync(creditAdd, "成语接龙");
                if (i != -1)
                    res = $"\n💎 积分：-{creditAdd} 累计：{creditValue:N0}";
            }
            return res;
        }

        public static string MinusCredit(BotMessage bm)
        {
            return MinusCreditAsync(bm).GetAwaiter().GetResult();
        }

        // 接龙成功数量
        public static async Task<int> GetCountAsync(long groupId, long userId)
        {
            int maxId = await GetMaxIdAsync(groupId);
            string query = $"SELECT ISNULL(COUNT(Id), 0) FROM {FullName} " +
                           $"WHERE UserId = {userId} AND Id >= {maxId}";

            var res = await QueryAsync(query);
            return res.AsInt();
        }

        public static int GetCount(long groupId, long userId)
        {
            return GetCountAsync(groupId, userId).GetAwaiter().GetResult();
        }

        // 添加接龙成功的数据到数据库
        public static async Task<int> AppendAsync(long groupId, long qq, string name, string chengYu, int gameNo)
        {
            return await InsertAsync([
                new Cov("GroupId", groupId),
                new Cov("UserId", qq),
                new Cov("UserName", name),
                new Cov("chengyu", chengYu),
                new Cov("GameNo", gameNo)
            ]);
        }

        public static int Append(long groupId, long qq, string name, string chengYu, int gameNo)
        {
            return AppendAsync(groupId, qq, name, chengYu, gameNo).GetAwaiter().GetResult();
        }

        // 是否重复成语
        public static async Task<bool> IsDupAsync(long groupId, long qq, string chengYu)
        {
            string query;
            if (groupId == 0)
            {
                query = $"SELECT TOP 1 1 FROM {FullName} " +
                        $"WHERE GroupId = {groupId} AND UserId = {qq} AND chengyu = '{chengYu}' " +
                        $"AND Id > (SELECT TOP 1 Id FROM {FullName} " +
                        $"WHERE GroupId = {groupId} AND UserId = {qq} AND GameNo = 1 ORDER BY Id DESC)";
            }
            else
            {
                query = $"SELECT TOP 1 1 FROM {FullName} " +
                        $"WHERE GroupId = {groupId} AND chengyu = '{chengYu}' " +
                        $"AND Id > (SELECT TOP 1 Id FROM {FullName} " +
                        $"WHERE GroupId = {groupId} AND GameNo = 1 ORDER BY Id DESC)";
            }

            return (await QueryScalarAsync<int>(query)) == 1;
        }

        public static bool IsDup(long groupId, long qq, string chengYu)
        {
            return IsDupAsync(groupId, qq, chengYu).GetAwaiter().GetResult();
        }
    }
}
