using Microsoft.Data.SqlClient;
using sz84.Bots.Entries;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Models.Office
{
    public class Income : MetaData<Income>
    {
        public override string TableName => "Income";
        public override string KeyField => "Id";

        public static float Total(long userId)
        {
            return GetWhere("sum(IncomeMoney) as res", $"UserId={userId}").AsFloat();
        }

        public static float TotalLastYear(long userId)
        {
            return GetWhere("sum(IncomeMoney) as res", $"UserId={userId} and abs(datediff(year, getdate(), IncomeDate)) <= 1").AsFloat();
        }

        //曾经
        public static bool IsVipOnce(long groupId)
        {
            return ExistsField("GroupId", groupId);
        }

        public static (string, SqlParameter[]) SqlInsert(long groupId, long goodsCount, string goodsName, decimal incomeMoney, string payMethod, string incomeTrade, string incomeInfo,
            long qqBuy, int insertBy)
        {
            return SqlInsert([
                                new Cov("GroupId", groupId),
                                new Cov("GoodsCount", goodsCount),
                                new Cov("GoodsName", goodsName),
                                new Cov("UserId", qqBuy),
                                new Cov("IncomeMoney", incomeMoney),
                                new Cov("PayMethod", payMethod),
                                new Cov("IncomeTrade", incomeTrade),
                                new Cov("IncomeInfo", incomeInfo),
                                new Cov("InsertBy", insertBy),
                            ]);
        }

        // 荣誉等级
        public static int GetClientLevel(long userId)
        {
            return Query($"select dbo.get_client_level(isnull(sum(IncomeMoney),0)) as res from {FullName} " +
                         $"where UserId = {userId}").AsInt();
        }

        // 荣誉榜
        public static string GetLevelList(long groupId)
        {
            return QueryRes($"select top 3 UserId, isnull(sum(IncomeMoney) as SIncome, dbo.get_client_level(isnull(sum(IncomeMoney),0) as client_level from {FullName} " +
                            $"where UserId in (select UserId from {CreditLog.FullName} where GroupId = {groupId}) group by UserId order by SIncome desc",
                            "【第{i}名】：[@:{0}]   荣誉等级：LV{1}\n");
        }

        // 荣誉排名
        public static string GetLeverOrder(long groupId, long userId)
        {
            return Query($"select count(UserId) + 1 from (select UserId from {FullName} " +
                         $"where UserId in (select UserId from {CreditLog.FullName} where GroupId = {groupId}) " +
                         $"group by UserId having sum(IncomeMoney) > (select sum(IncomeMoney) from {FullName} where UserId = {userId})) a");
        }

        public static string Today()
        {
            return GetWhere($"ISNULL(SUM(IncomeMoney),0)", $"DATEDIFF(DAY, IncomeDate - 5/24.0, GETDATE() ) < 1").AsCurrency();
        }

        public static string Yesterday()
        {
            return GetWhere($"ISNULL(SUM(IncomeMoney),0)", $"DATEDIFF(DAY, IncomeDate - 5/24.0, GETDATE() ) = 1").AsCurrency();
        }

        public static string ThisMonth()
        {
            return GetWhere($"ISNULL(SUM(IncomeMoney),0)", $"DATEDIFF(DAY, IncomeDate - 5/24.0, GETDATE()) <= 30").AsCurrency();
        }

        public static string ThisYear()
        {
            return GetWhere($"ISNULL(SUM(IncomeMoney),0)", $"DATEDIFF(DAY, IncomeDate - 5/24.0, GETDATE()) <= 365").AsCurrency();
        }

        public static string All()
        {
            return GetWhere($"SUM(IncomeMoney)", $"").AsCurrency();
        }

    }
}
