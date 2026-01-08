using Microsoft.Data.SqlClient;
using BotWorker.Infrastructure.Communication.Platforms.BotPublic;

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
        => GetSourceQQAsync(botQQ, qq).GetAwaiter().GetResult();

    public static async Task<long> GetSourceQQAsync(long botQQ, long qq)
    {
        return (await GetWhereAsync("Id", $"BotUin = {botQQ} and TargetUserId = {qq}")).AsLong();
    }

    public static bool GetIsBlack(long qq)
        => GetIsBlackAsync(qq).GetAwaiter().GetResult();

    public static async Task<bool> GetIsBlackAsync(long qq)
    {
        return await GetBoolAsync("IsBlack", qq);
    }

    public static bool GetIsFreeze(long qq)
        => GetIsFreezeAsync(qq).GetAwaiter().GetResult();

    public static async Task<bool> GetIsFreezeAsync(long qq)
    {
        return await GetBoolAsync("IsFreeze", qq);
    }

    public static bool GetIsShutup(long qq)
        => GetIsShutupAsync(qq).GetAwaiter().GetResult();

    public static async Task<bool> GetIsShutupAsync(long qq)
    {
        return await GetBoolAsync("IsShutup", qq);
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
        => GetCreditTypeAsync(groupId, qq).GetAwaiter().GetResult();

    public static async Task<string> GetCreditTypeAsync(long groupId, long qq)
    {
        return groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId) ? "本群积分" : await GetIsSuperAsync(qq) ? "超级积分" : "通用积分";
    }

    public static bool GetIsSuper(long qq)
        => GetIsSuperAsync(qq).GetAwaiter().GetResult();

    public static async Task<bool> GetIsSuperAsync(long qq)
    {
        if (qq == 0) return false;
        return await GetBoolAsync("IsSuper", qq);
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
        => GetStateResAsync(funcDefault).GetAwaiter().GetResult();

    public static async Task<string> GetStateResAsync(int funcDefault)
    {
        return await Task.FromResult(funcDefault switch
        {
            0 => "闲聊",
            1 => "AI",
            2 => "翻译",
            3 => "逗你玩",
            4 => "接龙",
            5 => "2048",
            _ => "闲聊"
        });
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
        => SetStateAsync(funcDefault, qq).GetAwaiter().GetResult();

    public static async Task<int> SetStateAsync(States funcDefault, long qq)
    {
        return await SetValueAsync("state", (int)funcDefault, qq);
    }

    public static async Task<int> AppendAsync(long botQQ, long groupId, long userId, string name, long userRef, string userOpenid = "", string groupOpenid = "")
    {
        return await ExistsAsync(userId)
            ? 0
            : await InsertAsync(new List<Cov> {
                new Cov("BotUin", botQQ),
                    new Cov("UserOpenid", userOpenid),
                    new Cov("GroupOpenid", groupOpenid),
                    new Cov("GroupId", groupId),
                    new Cov("Id", userId),
                    new Cov("Credit", userOpenid.IsNull() ? 50 : 5000),
                    new Cov("Name", name),
                    new Cov("RefUserId", userRef),
            });
    }

    public static async Task<int> AppendUserAsync(long botUin, long groupId, long userId, string name, string userOpenid = "", string groupOpenid = "")
    {
        if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                    3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

        int i = await AppendAsync(botUin, groupId, userId, name, GroupInfo.GetGroupOwner(groupId), userOpenid, groupOpenid);
        if (i == -1) return i;

        i = await GroupMember.AppendAsync(groupId, userId, name, "");
        if (i == -1) return i;

        if (BotInfo.GetBool("IsCredit", botUin))
        {
            i = await Friend.AppendAsync(botUin, userId, name);
            if (i == -1) return i;
        }

        return i;
    }

    public static int Append(long botQQ, long groupId, long userId, string name, long userRef, string userOpenid = "", string groupOpenid = "")
    {
        return Exists(userId)
            ? 0
            : Insert(new List<Cov> {
                new Cov("BotUin", botQQ),
                    new Cov("UserOpenid", userOpenid),
                    new Cov("GroupOpenid", groupOpenid),
                    new Cov("GroupId", groupId),
                    new Cov("Id", userId),
                    new Cov("Credit", userOpenid.IsNull() ? 50 : 5000),
                    new Cov("Name", name),
                    new Cov("RefUserId", userRef),
            });
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
