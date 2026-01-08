using sz84.Bots.BotMessages;
using sz84.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;
using sz84.Core.Data;

namespace BotWorker.Modules.Games
{
    public class Jielong : MetaData<Jielong>
    {
        public override string TableName => "Chengyu";
        public override string KeyField => "Id";


        // 为机器人获取一个随机接龙结果
        public static string GetJielong(long groupId, long UserId, string currCy)
        {
            string pinyin = Chengyu.PinYinLast(currCy);
            string sql = $"SELECT TOP 1 chengyu FROM {Chengyu.FullName} " +
                           $"WHERE pinyin LIKE '{pinyin} %' AND chengyu NOT IN " +
                           $"(SELECT chengyu FROM {FullName} WHERE GroupId = {groupId} AND UserId = {UserId} " +
                           $"AND Id > (SELECT TOP 1 Id FROM {FullName} WHERE GroupId = {groupId} " +
                           $"AND UserId = {UserId} AND GameNo = 1 ORDER BY InsertDate DESC)) " +
                           $"ORDER BY NEWID()";

            return Query(sql);
        }



        // 接龙游戏最大ID
        public static int GetMaxId(long groupId)
        {
            return Query($"SELECT MAX(Id) FROM {FullName} WHERE GroupId = {groupId} AND GameNo = 1").AsInt();
        }

        // 接龙成功数量
        public static string GetGameCount(long groupId, long qq)
        {
            return Query($"SELECT {DbName}.DBO.[getChengyuGameCount]({groupId},{qq})");
        }

        // 接龙加分总数
        public static long GetCreditAdd(long userId)
        {
            string query = $"SELECT ISNULL(SUM(CreditAdd), 0) FROM {CreditLog.FullName} " +
                           $"WHERE UserId = {userId} AND CreditInfo = '成语接龙' " +
                           $"AND ABS(DATEDIFF(DAY, InsertDate, GETDATE())) < 1";

            return Query(query).AsLong();
        }

        // 成语接龙加分
        public static string AddCredit(BotMessage bm)
        {
            //接龙加分，接龙自己的不加分，答错扣分
            var creditAdd = 10;
            string res = "";
            if ((bm.IsGuild || GetCreditAdd(bm.UserId) < 2000) && bm.Group.IsCreditSystem)
            {
                (int i, long creditValue) = bm.AddCredit(creditAdd, "成语接龙");
                if (i != -1)
                    res = $"\n💎 积分：+{creditAdd}，累计：{creditValue:N0}";
            }
            return res;
        }

        // 成语接龙扣分
        public static string MinusCredit(BotMessage bm)
        {
            if (bm.IsGuild || bm.IsRealProxy) return "";

            string res = "";

            var creditAdd = 10;
            int c_chengyu = GetCount(bm.RealGroupId, bm.UserId);
            if (c_chengyu > 0 && bm.Group.IsCreditSystem)
            {
                (int i, long creditValue) = bm.MinusCredit(creditAdd, "成语接龙");
                if (i != -1)
                    res = $"\n💎 积分：-{creditAdd} 累计：{creditValue:N0}";
            }
            return res;
        }

        // 接龙成功数量
        public static int GetCount(long groupId, long userId)
        {
            string query = $"SELECT ISNULL(COUNT(Id), 0) FROM {FullName} " +
                           $"WHERE UserId = {userId} AND Id >= {GetMaxId(groupId)}";

            return Query(query).AsInt();
        }

        // 添加接龙成功的数据到数据库
        public static int Append(long groupId, long qq, string name, string chengYu, int gameNo)
        {
            return Insert([
                new Cov("GroupId", groupId),
                new Cov("UserId", qq),
                new Cov("UserName", name),
                new Cov("chengyu", chengYu),
                new Cov("GameNo", gameNo)
                        ]);
        }

        // 是否重复成语
        public static bool IsDup(long groupId, long qq, string chengYu)
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

            return Query(query).AsInt() == 1;
        }
    }
}
