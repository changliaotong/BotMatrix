using sz84.Bots.Entries;
using sz84.Bots.Models.Office;
using sz84.Bots.Public;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;
using sz84.Groups;
using Microsoft.Data.SqlClient;

namespace BotWorker.Domain.Entities;

public partial class UserInfo : MetaDataGuid<UserInfo>
{
    public override string TableName => "User";
    public override string KeyField => "Id";

    public string Name { get; set; } = string.Empty;
    public string UserOpenId { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; }
    public long Credit { get; set; }
    public long CreditTotal => Credit + SaveCredit;
    public long CreditFreeze { get; set; }
    public long Coins { get; set; }
    public long CoinsFreeze { get; set; }
    public int Sz84Uid { get; set; }
    public string Sz84UserName { get; set; } = string.Empty;
    public int HomeUid { get; set; }
    public string HomeUserName { get; set; } = string.Empty;
    public string HomeRealName { get; set; } = string.Empty;
    public long BotUin { get; set; }
    public int State { get; set; }
    public int IsOpen { get; set; }
    public string SzTong { get; set; } = string.Empty;
    public string CityName { get; set; } = string.Empty;
    public long DefaultGroup { get; set; }
    public bool IsDefaultHint { get; set; }
    public DateTime BindDate { get; set; }
    public DateTime BindDateHome { get; set; }
    public bool IsBlack { get; set; }
    public int TeachLevel { get; set; }
    public bool IsCoins { get; set; }
    public bool XCredit { get; set; }
    public int RCq { get; set; }
    public long RefUserId { get; set; }
    public Guid UserGuid { get; set; }
    public bool IsBlock { get; set; }
    public long SaveCredit { get; set; }
    public DateTime UpgradeDate { get; set; }
    public long PartnerUserId { get; set; }
    public long FreezeCredit { get; set; }
    public DateTime VipStart { get; set; }
    public DateTime VipEnd { get; set; }
    public string LastChengyu { get; set; } = string.Empty;
    public decimal Balance { get; set; }
    public decimal BalanceFreeze { get; set; }
    public long CreditGiving { get; set; }
    public int LvValue { get; set; }
    public decimal SumIncome { get; set; }
    public DateTime SuperDate { get; set; }
    public bool IsSuper { get; set; }
    public bool IsFreeze { get; set; }
    public long GroupId { get; set; }
    public bool IsSz84 { get; set; }
    public DateTime Sz84Date { get; set; }
    public long AnswerId { get; set; }
    public DateTime AnswerDate { get; set; }
    public bool IsTeach { get; set; }
    public bool IsShutup { get; set; }
    public string SystemPrompt { get; set; } = string.Empty;
    public long Tokens { get; set; }
    public bool IsAgent { get; set; }
    public bool IsAI { get; set; }
    public bool Xxian { get; set; }
    public long AgentId { get; set; }
    public bool IsSendHelpInfo { get; set; }
    public bool IsLog { get; set; }
    public bool IsMusicLogo { get; set; }

    //增加积分
    public static (int, long) AddCredit(long botUin, long groupId, string groupName, long qq, string name, long creditAdd, string creditInfo)
    {
        var creditValue = GetCredit(groupId, qq);
        if (Append(botUin, groupId, qq, name, GroupInfo.GetGroupOwner(groupId)) == -1)
            return (-1, creditValue);

        var sql = SqlAddCredit(botUin, groupId, qq, creditAdd);
        var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditAdd, creditInfo);
        var result = ExecTrans(sql, sql2);
        if (result == 0)
        {
            // 同步更新缓存中的 Credit 字段
            SyncCacheField(qq, groupId, "Credit", creditValue + creditAdd);
        }
        return (result, creditValue + creditAdd);
    }

    public static (int, long) MinusCredit(long botUin, long groupId, string groupName, long qq, string name, long creditMinus, string creditInfo)
        => AddCredit(botUin, groupId, groupName, qq, name, -creditMinus, creditInfo);


    //增加积分sql
    public static (string, SqlParameter[]) SqlAddCredit(long botUin, long groupId, long userId, long creditPlus)
    {
        if (GroupInfo.GetIsCredit(groupId))
        {
            return GroupMember.SqlAddCredit(groupId, userId, creditPlus);
        }
        else if (BotInfo.GetIsCredit(botUin))
        {
            return Friend.SqlAddCredit(botUin, userId, creditPlus);
        }
        else
        {
            if (Exists(userId))
                return SqlPlus("Credit", creditPlus, userId);
            else
                return SqlInsert([
                    new Cov("BotUin", botUin),
                        new Cov("GroupId", groupId),
                        new Cov("Id", userId),
                        new Cov("Credit", creditPlus),
                ]);
        }
    }

    //转账积分
    public static int TransferCredit(long botUin, long groupId, string groupName, long qq, string name, long qqTo, string nameTo, long creditMinus, long creditAdd, ref long creditValue, ref long creditValue2, string transferInfo)
    {
        int i = AppendUser(botUin, groupId, qqTo, nameTo);
        if (i == -1)
            return i;

        creditValue = GetCredit(groupId, qq);
        if (creditValue < creditMinus)
            return -1;

        creditValue -= creditMinus;
        creditValue2 = GetCredit(groupId, qqTo) + creditAdd;

        var sql = SqlAddCredit(botUin, groupId, qq, -creditMinus);
        var sql2 = SqlAddCredit(botUin, groupId, qqTo, creditAdd);
        var sql3 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditMinus, $"{transferInfo}扣分：{qqTo}");
        var sql4 = CreditLog.SqlHistory(botUin, groupId, groupName, qqTo, nameTo, creditAdd, $"{transferInfo}加分：{qq}");

        int result = ExecTrans(sql, sql3, sql2, sql4);
        if (result == 0)
        {
            SyncCacheField(qq, groupId, "Credit", creditValue);
            SyncCacheField(qqTo, groupId, "Credit", creditValue2);
        }
        return result;
    }


    //读取积分
    public static long GetCredit(long botUin, long groupId, long qq)
    {
        return groupId != 0 && GroupInfo.GetIsCredit(groupId)
            ? GroupMember.GetGroupCredit(groupId, qq)
            : GetCredit(botUin, qq);
    }

    public static long GetCredit(long userId)
    {
        return GetLong("Credit", userId);
    }

    //读取积分
    public static long GetCredit(long botUin, long userId)
    {
        return BotInfo.GetIsCredit(botUin) ? Friend.GetCredit(botUin, userId) : GetLong("credit", userId);
    }

    //积分总额
    public static long GetTotalCredit(long userId) => GetCredit(userId) + GetSaveCredit(userId);

    public static long GetTotalCredit(long groupId, long qq) => GetCredit(groupId, qq) + GetSaveCredit(groupId, qq);

    public static long GetSaveCredit(long botUin, long userId)
    {
        return BotInfo.GetIsCredit(botUin)
            ? Friend.GetSaveCredit(botUin, userId)
            : GetSaveCredit(userId);
    }

    public static long GetSaveCredit(long botUin, long groupId, long qq)
    {
        return GroupInfo.GetIsCredit(groupId)
            ? GroupMember.GetLong("SaveCredit", groupId, qq)
            : GetSaveCredit(qq);
    }

    public static long GetSaveCredit(long userId)
    {
        return GetLong("SaveCredit", userId);
    }

    public static (string, SqlParameter[]) SqlSaveCredit(long botUin, long groupId, long userId, long creditSave)
    {
        return GroupInfo.GetIsCredit(groupId)
            ? GroupMember.SqlSaveCredit(groupId, userId, creditSave)
            : BotInfo.GetIsCredit(botUin) ? Friend.SqlSaveCredit(botUin, userId, creditSave)
                             : SqlSetValues($"Credit = Credit - ({creditSave}), SaveCredit = isnull(SaveCredit, 0) + ({creditSave})", userId);
    }

    public static (string, SqlParameter[]) SqlFreezeCredit(long userId, long creditFreeze)
    {
        return SqlSetValues($"Credit = Credit - ({creditFreeze}), FreezeCredit = isnull(FreezeCredit, 0) + ({creditFreeze})", userId);
    }

    public static int DoFreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditFreeze)
    {
        long creditValue = GetCredit(groupId, qq);
        if (creditValue < creditFreeze)
            return -1;

        var sql = SqlFreezeCredit(qq, creditFreeze);
        var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, -creditFreeze, "冻结积分");
        var result = ExecTrans(sql, sql2);
        if (result == 0)
        {
            // 获取当前最新值并同步到缓存
            var currentCredit = GetCredit(groupId, qq);
            var currentFreeze = GetFreezeCredit(qq);
            SyncCacheField(qq, groupId, "Credit", currentCredit);
            SyncCacheField(qq, groupId, "FreezeCredit", currentFreeze);
        }
        return result;
    }

    public static long GetFreezeCredit(long qq) => GetLong("FreezeCredit", qq);

    public static int UnfreezeCredit(long botUin, long groupId, string groupName, long qq, string name, long creditUnfreeze)
    {
        long creditValue = GetFreezeCredit(qq);
        if (creditValue < creditUnfreeze)
            return -1;

        var sql = SqlFreezeCredit(qq, -creditUnfreeze);
        var sql2 = CreditLog.SqlHistory(botUin, groupId, groupName, qq, name, creditUnfreeze, "解冻积分");
        var result = ExecTrans(sql, sql2);
        if (result == 0)
        {
            SyncCacheField(qq, groupId, "Credit", GetCredit(groupId, qq) + creditUnfreeze);
        }
        return result;
    }

    public static long GetCreditRanking(long botUin, long groupId, long qq)
    {
        long credit_value = GetCredit(groupId, qq);
        return GroupInfo.GetIsCredit(groupId)
            ? GroupMember.CountWhere($"GroupId = {groupId} and Credit > {credit_value}") + 1
            : BotInfo.GetBool("IsCredit", botUin)
                ? Friend.CountWhere($"BotUin = {botUin} and Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1
                : CountWhere($"Credit > {credit_value} and Id in (select UserId from {GroupMember.FullName} where GroupId = {groupId})") + 1;
    }

    public static long GetCreditRankingAll(long qq)
    {
        return CountWhere($"Credit + SaveCredit > {GetTotalCredit(qq)}") + 1;
    }

    //用户头像CQ
    public static string GetHeadCQ(long user, int size = 100)
    {
        return $"[CQ:image,file={GetHead(user, size)}]";
    }
    public static string GetHead(long user, int size = 100)
    {
        return $"https://q1.qlogo.cn/g?b=qq&nk={user}&s={size}";
    }

    public static long GetSourceQQ(long botQQ, long qq)
    {
        return GetWhere("Id", $"BotUin = {botQQ} and TargetUserId = {qq}").AsLong();
    }

    public static bool SubscribedPublic(long qq)
    {
        return ClientPublic.SubscribeCompayPublic(qq);
    }

    public static bool StartWith285or300(long qq)
    {
        return qq > 2850000000 && qq.AsString()[..3].In("285", "300") && Income.Total(qq) < 10;
    }

    public static string GetCreditType(long groupId, long qq)
    {
        return groupId != 0 && GroupInfo.GetIsCredit(groupId) ? "本群积分" : GetIsSuper(qq) ? "超级积分" : "通用积分";
    }

    public static bool GetIsSuper(long qq)
    {
        if (qq == 0) return false;
        return GetBool("IsSuper", qq);
    }

    public static int NewGuessNumGame(int csz_res, long csz_credit, long qq)
    {
        return UpdateCszGame(csz_res, csz_credit, 0, qq);
    }

    public static int UpdateCszGame(int csz_res, long csz_credit, int csz_times, long qq)
    {
        return Update([
            new Cov("CszRes", csz_res),
                new Cov("CszCredit", csz_credit),
                new Cov("CszTimes", csz_times),
        ], qq);
    }

    public static bool IsOwner(long groupId, long qq)
    {
        return GroupInfo.IsOwner(groupId, qq);
    }

    public static string GetStateRes(int funcDefault)
    {
        return funcDefault switch
        {
            0 => "闲聊",
            1 => "AI",
            2 => "翻译",
            3 => "逗你玩",
            4 => "接龙",
            5 => "2048",
            _ => "闲聊"
        };
    }

    public enum States
    {
        Chat,
        AI,
        Translate,
        Douniwan,
        GameCy,
        G2048,
    }

    public static int SetState(States funcDefault, long qq)
    {
        return SetValue("state", (int)funcDefault, qq);
    }

    public static int Append(long botQQ, long groupId, long userId, string name, long userRef, string userOpenid = "", string groupOpenid = "")
    {
        return Exists(userId)
            ? 0
            : Insert([
                new Cov("BotUin", botQQ),
                    new Cov("UserOpenid", userOpenid),
                    new Cov("GroupOpenid", groupOpenid),
                    new Cov("GroupId", groupId),
                    new Cov("Id", userId),
                    new Cov("Credit", userOpenid.IsNull() ? 50 : 5000),
                    new Cov("Name", name),
                    new Cov("RefUserId", userRef),
            ]);
    }

    public static int AppendUser(long botUin, long groupId, long userId, string name, string userOpenid = "", string groupOpenid = "")
    {
        if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                    3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

        int i = Append(botUin, groupId, userId, name, GroupInfo.GetGroupOwner(groupId), userOpenid, groupOpenid);
        if (i == -1) return i;

        i = GroupMember.Append(groupId, userId, name, "");
        if (i == -1) return i;

        if (BotInfo.GetBool("IsCredit", botUin))
        {
            i = Friend.Append(botUin, userId, name);
            if (i == -1) return i;
        }

        return i;
    }

    public static string GetResetDefaultGroup(long qq)
    {
        return SetValue("DefaultGroup", BotInfo.GroupCrm, qq) == -1 ? "" : $"\n默认群已重置为 {BotInfo.GroupCrm}";
    }
}
