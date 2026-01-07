using BotWorker.Bots.Entries;
using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Users;

namespace BotWorker.Bots.Groups
{
    class GroupProps : MetaData<GroupProps>
    {
        /// <summary>
        /// 道具系统 禁言卡 飞机票 免踢卡 座骑系统
        /// </summary>
        public override string TableName => "Props";
        public override string KeyField => "Id";

        public const string PropClosed = "道具系统已关闭";

        public static long GetId(long groupId, long qq, long propId)
        {
            return GetWhere($"Id", $"GroupId = {groupId} AND UserId = {qq} AND PropId = {propId} AND IsUsed = 0").AsLong();
        }

        public static bool HaveProp(long groupId, long userId, long propId)
        {
            return GetId(groupId, userId, propId) != 0;
        }

        public static int UseProp(long groupId, long userId, long propId, long qqProp)
        {
            return Update($"UsedDate = getdate(), UsedUserId = {qqProp}, IsUsed = 1", GetId(groupId, userId, propId));
        }

        public static string GetMyPropList(long groupId, long userId)
        {
            if (IsClosed(groupId)) return PropClosed;

            string sql = $"select top 10 PropName, PropPrice from {FullName} a inner join sz84_robot..Prop b " +
                         $"on a.PropId = b.Id where a.GroupId = {groupId} and a.UserId = {userId}";
            return QueryRes(sql, "{0} 价格：{1}分");
        }

        public static bool IsClosed(long groupId)
        {
            return GroupInfo.GetBool("IsProp", groupId);
        }

        public static string GetBuyRes(long botUin, long groupId, string groupName, long qq, string name, string cmdPara)
        {
            if (IsClosed(groupId)) return PropClosed;

            if (cmdPara == "" | cmdPara == "道具")
                return Prop.GetPropList();
            else
            {
                long prop_id = Prop.GetId(cmdPara);
                if (prop_id != 0)
                {
                    long credit_value = UserInfo.GetCredit(groupId, qq);
                    int prop_price = Prop.GetInt("PropPrice", prop_id);
                    if (credit_value < prop_price)
                        return $"您的积分{credit_value}不足{prop_price}";
                    credit_value -= prop_price;
                    var sqlAddCredit = UserInfo.SqlAddCredit(botUin, groupId, qq, -prop_price);
                    var sqlCreditHis = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -prop_price, $"购买道具:{prop_id}");
                    var sqlBuyPorp = SqlInsert([
                                                new Cov("GroupId", groupId),
                                                new Cov("UserId", qq),
                                                new Cov("PropId", prop_id),
                                            ]);
                    int i = ExecTrans(sqlAddCredit, sqlCreditHis, sqlBuyPorp);
                    return i == -1
                        ? RetryMsg
                        : $"购买道具成功\n积分：-{prop_price}，累计：{credit_value}";
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

        public static long GetId(string propName)
        {
            return GetWhere("Id", $"PropName = {propName.Quotes()}").AsLong();
        }

        public static string GetPropList()
        {
            return QueryRes($"select top 10 PropName, PropPrice from {FullName} where IsValid = 1 order by PropName", "{0} 价格：{1}分");
        }

        public static string GetPropRes(long groupId)
        {
            int is_prop = GroupInfo.GetInt("IsProp", groupId);
            return is_prop == 1 
                ? "道具系统\n可用道具：\n禁言卡\n飞机票\n免踢卡\n购买道具请发送【购买 + 道具名称】"
                : GroupProps.PropClosed;
        }

    }
}
