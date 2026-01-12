using System.Data;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Modules.Office
{
    public class Income : MetaData<Income>
    {
        public override string TableName => "Income";
        public override string KeyField => "Id";

        public static async Task<float> TotalAsync(long userId)
        {
            return (await GetWhereAsync("sum(IncomeMoney) as res", $"UserId={userId}")).AsFloat();
        }

        public static float Total(long userId) => TotalAsync(userId).GetAwaiter().GetResult();

        public static async Task<float> TotalLastYearAsync(long userId)
        {
            return (await GetWhereAsync("sum(IncomeMoney) as res", $"UserId={userId} and abs({SqlDateDiff("year", SqlDateTime, "IncomeDate")}) <= 1")).AsFloat();
        }

        //曾经
        public static async Task<bool> IsVipOnceAsync(long groupId)
        {
            return await ExistsFieldAsync("GroupId", groupId);
        }

        public static (string, IDataParameter[]) SqlInsert(long groupId, long goodsCount, string goodsName, decimal incomeMoney, string payMethod, string incomeTrade, string incomeInfo,
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
        public static async Task<int> GetClientLevelAsync(long userId)
        {
            string func = IsPostgreSql ? "get_client_level" : $"{DbName}.dbo.get_client_level";
            return (await QueryAsync($"select {func}({SqlIsNull("sum(IncomeMoney)", "0")}) as res from {FullName} " +
                         $"where UserId = {userId}")).AsInt();
        }

        // 荣誉榜
        public static async Task<string> GetLevelListAsync(long groupId)
        {
            string func = IsPostgreSql ? "get_client_level" : $"{DbName}.dbo.get_client_level";
            return await QueryResAsync($"select {SqlTop(3)} UserId, {SqlIsNull("sum(IncomeMoney)", "0")} as SIncome, {func}({SqlIsNull("sum(IncomeMoney)", "0")}) as client_level from {FullName} " +
                            $"where UserId in (select UserId from {CreditLog.FullName} where GroupId = {groupId}) group by UserId order by SIncome desc {SqlLimit(3)}",
                            "【第{i}名】：[@:{0}]   荣誉等级：LV{1}\n");
        }

        // 荣誉排名
        public static async Task<string> GetLeverOrderAsync(long groupId, long userId)
        {
            return await QueryScalarAsync($"select count(UserId) + 1 from (select UserId from {FullName} " +
                         $"where UserId in (select UserId from {CreditLog.FullName} where GroupId = {groupId}) " +
                         $"group by UserId having sum(IncomeMoney) > (select sum(IncomeMoney) from {FullName} where UserId = {userId})) a");
        }

        public static async Task<string> TodayAsync()
        {
            return (await GetWhereAsync(SqlIsNull("SUM(IncomeMoney)", "0"), $"{SqlDateDiff("DAY", SqlDateAdd("hour", -5, "IncomeDate"), SqlDateTime)} < 1")).AsCurrency();
        }

        public static async Task<string> YesterdayAsync()
        {
            return (await GetWhereAsync(SqlIsNull("SUM(IncomeMoney)", "0"), $"{SqlDateDiff("DAY", SqlDateAdd("hour", -5, "IncomeDate"), SqlDateTime)} = 1")).AsCurrency();
        }

        public static async Task<string> ThisMonthAsync()
        {
            return (await GetWhereAsync(SqlIsNull("SUM(IncomeMoney)", "0"), $"{SqlDateDiff("DAY", SqlDateAdd("hour", -5, "IncomeDate"), SqlDateTime)} <= 30")).AsCurrency();
        }

        public static async Task<string> ThisYearAsync()
        {
            return (await GetWhereAsync(SqlIsNull("SUM(IncomeMoney)", "0"), $"{SqlDateDiff("DAY", SqlDateAdd("hour", -5, "IncomeDate"), SqlDateTime)} <= 365")).AsCurrency();
        }

        public static async Task<string> AllAsync()
        {
            return (await GetWhereAsync($"SUM(IncomeMoney)", $"")).AsCurrency();
        }
    }
}
