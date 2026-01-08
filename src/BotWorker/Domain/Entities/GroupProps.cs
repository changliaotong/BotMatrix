namespace BotWorker.Domain.Entities
{
    class GroupProps : MetaData<GroupProps>
    {
        /// <summary>
        /// 道具系统 禁言卡 飞机票 免踢卡 座骑系统
        /// </summary>
        public override string TableName => "Props";
        public override string KeyField => "Id";

        public const string PropClosed = "道具系统已关闭";

        public static async Task<long> GetIdAsync(long groupId, long qq, long propId)
        {
            return (await GetWhereAsync($"Id", $"GroupId = {groupId} AND UserId = {qq} AND PropId = {propId} AND IsUsed = 0")).AsLong();
        }

        public static async Task<bool> HavePropAsync(long groupId, long userId, long propId)
        {
            return await GetIdAsync(groupId, userId, propId) != 0;
        }

        public static async Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp)
        {
            return await UpdateAsync($"UsedDate = getdate(), UsedUserId = {qqProp}, IsUsed = 1", await GetIdAsync(groupId, userId, propId));
        }

        public static async Task<string> GetMyPropListAsync(long groupId, long userId)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;

            string sql = $"select top 10 PropName, PropPrice from {FullName} a inner join sz84_robot..Prop b " +
                         $"on a.PropId = b.Id where a.GroupId = {groupId} and a.UserId = {userId}";
            return await QueryResAsync(sql, "{0} 价格：{1}分");
        }

        public static async Task<bool> IsClosedAsync(long groupId)
        {
            return await GroupInfo.GetBoolAsync("IsProp", groupId);
        }

        public static string GetBuyRes(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
            => GetBuyResAsync(botUin, groupId, groupName, qq, name, cmdPara).GetAwaiter().GetResult();

        public static async Task<string> GetBuyResAsync(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (await IsClosedAsync(groupId)) return PropClosed;

            if (cmdPara == "" | cmdPara == "道具")
                return await Prop.GetPropListAsync();
            else
            {
                long prop_id = await Prop.GetIdAsync(cmdPara);
                if (prop_id != 0)
                {
                    long credit_value = await UserInfo.GetCreditAsync(groupId, qq);
                    int prop_price = await Prop.GetIntAsync("PropPrice", prop_id);
                    if (credit_value < prop_price)
                        return $"您的积分{credit_value}不足{prop_price}";
                    
                    using var trans = await BeginTransactionAsync();
                    try
                    {
                        // 1. 通用加积分函数 (含日志记录)
                        var res = await UserInfo.AddCreditAsync(botUin, groupId, groupName, qq, name, -prop_price, $"购买道具:{prop_id}", trans);
                        if (res.Result == -1) throw new Exception("更新积分失败");

                        // 2. 插入道具购买记录
                        var (sql, paras) = SqlInsert([
                                                        new Cov("GroupId", groupId),
                                                        new Cov("UserId", qq),
                                                        new Cov("PropId", prop_id),
                                                    ]);
                        await ExecAsync(sql, trans, paras);

                        await trans.CommitAsync();

                        UserInfo.SyncCacheField(qq, groupId, "Credit", res.CreditValue);

                        return $"购买道具成功\n积分：-{prop_price}，累计：{res.CreditValue}";
                    }
                    catch (Exception ex)
                    {
                        await trans.RollbackAsync();
                        Console.WriteLine($"[GetBuyRes Error] {ex.Message}");
                        return RetryMsg;
                    }
                }
                else
                    return "没有此道具";
            }
        }

    }

    /// <summary>
    /// 道具系统 禁言卡 飞机票 免踢卡 座骑系统
    /// </summary>
    class Prop : MetaData<Prop>
    {
        
        public override string TableName => "Prop";
        public override string KeyField => "Id";

        public static async Task<long> GetIdAsync(string propName)
        {
            return (await GetWhereAsync("Id", $"PropName = {propName.Quotes()}")).AsLong();
        }

        public static long GetId(string propName)
            => GetIdAsync(propName).GetAwaiter().GetResult();

        public static async Task<string> GetPropListAsync()
        {
            return await QueryResAsync($"select top 10 PropName, PropPrice from {FullName} where IsValid = 1 order by PropName", "{0} 价格：{1}分");
        }

        public static string GetPropList()
            => GetPropListAsync().GetAwaiter().GetResult();

        public static async Task<string> GetPropResAsync(long groupId)
        {
            int is_prop = await GroupInfo.GetIntAsync("IsProp", groupId);
            return is_prop == 1 
                ? "道具系统\n可用道具：\n禁言卡\n飞机票\n免踢卡\n购买道具请发送【购买 + 道具名称】"
                : GroupProps.PropClosed;
        }

        public static string GetPropRes(long groupId)
            => GetPropResAsync(groupId).GetAwaiter().GetResult();

    }
}
