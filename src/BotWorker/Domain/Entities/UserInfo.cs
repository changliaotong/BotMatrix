using System.Data;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities;

public partial class UserInfo : MetaDataGuid<UserInfo>
{
    public override string TableName => "User";
    public override string KeyField => "Id";

    public string Name { get; set; } = string.Empty;
    public string UserOpenId { get; set; } = string.Empty;
    public DateTime InsertDate { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long Credit { get; set; }
    public long CreditTotal => Credit + SaveCredit;
    [JsonIgnore]
    [HighFrequency]
    public long CreditFreeze { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long Coins { get; set; }
    [JsonIgnore]
    [HighFrequency]
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
    [JsonIgnore]
    [HighFrequency]
    public long SaveCredit { get; set; }
    public DateTime UpgradeDate { get; set; }
    public long PartnerUserId { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long FreezeCredit { get; set; }
    public DateTime VipStart { get; set; }
    public DateTime VipEnd { get; set; }
    public string LastChengyu { get; set; } = string.Empty;
    [JsonIgnore]
    [HighFrequency]
    public decimal Balance { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public decimal BalanceFreeze { get; set; }
    public long CreditGiving { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public int LvValue { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public decimal SumIncome { get; set; }
    public DateTime SuperDate { get; set; }
    public bool IsSuper { get; set; }
    public bool IsFreeze { get; set; }
    public long GroupId { get; set; }
    public bool IsSz84 { get; set; }
    public DateTime Sz84Date { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public long AnswerId { get; set; }
    [JsonIgnore]
    [HighFrequency]
    public DateTime AnswerDate { get; set; }
    public bool IsTeach { get; set; }
    public bool IsShutup { get; set; }
    public string SystemPrompt { get; set; } = string.Empty;
    [JsonIgnore]
    [HighFrequency]
    public long Tokens { get; set; }
    public bool IsAgent { get; set; }
    public bool IsAI { get; set; }
    public bool Xxian { get; set; }
    public long AgentId { get; set; }
    public bool IsSendHelpInfo { get; set; }
    public bool IsLog { get; set; }
    public bool IsMusicLogo { get; set; }

    public static async Task<UserInfo?> GetByOpenIdAsync(string openId, long botUin)
    {
        return (await QueryWhere("UserOpenId = {0} AND BotUin = {1}", openId, botUin)).FirstOrDefault();
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

    public static long GetSourceQQ(long botQQ, long qq) => GetSourceQQAsync(botQQ, qq).GetAwaiter().GetResult();

    public static async Task<long> GetSourceQQAsync(long botQQ, long qq)
    {
        return (await GetWhereAsync("Id", $"BotUin = {botQQ} and TargetUserId = {qq}")).AsLong();
    }

    public static async Task<bool> GetIsBlackAsync(long qq)
    {
        return await GetBoolAsync("IsBlack", qq);
    }

    public static async Task<bool> GetIsFreezeAsync(long qq)
    {
        return await GetBoolAsync("IsFreeze", qq);
    }

    public static async Task<bool> GetIsShutupAsync(long qq)
    {
        return await GetBoolAsync("IsShutup", qq);
    }

    public static bool SubscribedPublic(long qq)
    {
        return ClientPublic.SubscribeCompayPublic(qq);
    }

    public static async Task<bool> StartWith285or300Async(long qq)
    {
        return qq > 2850000000 && qq.AsString()[..3].In("285", "300") && await Income.TotalAsync(qq) < 10;
    }

    public static async Task<string> GetCreditTypeAsync(long botUin, long groupId, long qq)
    {
        if (groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId))
            return "本群积分";
        if (await BotInfo.GetIsCreditAsync(botUin))
            return "本机积分";
        return await GetIsSuperAsync(qq) ? "超级积分" : "通用积分";
    }

    public static async Task<bool> GetIsSuperAsync(long qq)
    {
        if (qq == 0) return false;
        return await GetBoolAsync("IsSuper", qq);
    }

    public static async Task<int> NewGuessNumGameAsync(int csz_res, long csz_credit, long qq, IDbTransaction? trans = null)
    {
        return await UpdateCszGameAsync(csz_res, csz_credit, 0, qq, trans);
    }

    public static async Task<int> UpdateCszGameAsync(int csz_res, long csz_credit, int csz_times, long qq, IDbTransaction? trans = null)
    {
        return await UpdateAsync(new
        {
            CszRes = csz_res,
            CszCredit = csz_credit,
            CszTimes = csz_times,
        }, qq, trans: trans);
    }

    public static async Task<bool> IsOwnerAsync(long groupId, long qq)
    {
        return await GroupInfo.IsOwnerAsync(groupId, qq);
    }

    public static bool IsOwner(long groupId, long qq)
    {
        return GroupInfo.IsOwner(groupId, qq);
    }

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

    public static string GetStateRes(int state)
    {
        return ((States)state) switch
        {
            States.Chat => "闲聊",
            States.AI => "AI",
            States.Translate => "翻译",
            States.Douniwan => "逗你玩",
            States.GameCy => "成语接龙",
            States.G2048 => "2048",
            _ => "闲聊",
        };
    }

    public static async Task<int> SetStateAsync(States funcDefault, long qq)
    {
        return await SetValueAsync("state", (int)funcDefault, qq);
    }

    public static async Task<int> AppendAsync(long botQQ, long groupId, long userId, string name, long userRef, string userOpenid = "", string groupOpenid = "", IDbTransaction? trans = null)
    {
        // 优化：ExistsAsync 内部也会使用 trans，确保检查和插入在同一隔离级别下
        if (await ExistsAsync(userId, null, trans)) return 0;
        
        return await InsertAsync(new
        {
            BotUin = botQQ,
            UserOpenid = userOpenid,
            GroupOpenid = groupOpenid,
            GroupId = groupId,
            Id = userId,
            Credit = userOpenid.IsNull() ? 0 : 5000,
            Name = name,
            RefUserId = userRef,
        }, trans);
    }

    public static int Append(long botQQ, long groupId, long userId, string name, long userRef, string userOpenid = "", string groupOpenid = "")
        => AppendAsync(botQQ, groupId, userId, name, userRef, userOpenid, groupOpenid).GetAwaiter().GetResult();

    public static async Task<long> GetCreditAsync(long userId)
    {
        return await GetLongAsync("Credit", userId);
    }

    public static async Task<long> GetCreditAsync(long groupId, long userId)
    {
        if (groupId != 0 && await GroupInfo.GetIsCreditAsync(groupId))
        {
            return await GroupMember.GetLongAsync("GroupCredit", groupId, userId);
        }
        return await GetCreditAsync(userId);
    }

    public static long GetCredit(long groupId, long userId)
        => GetCreditAsync(groupId, userId).GetAwaiter().GetResult();

    public static async Task<int> AppendUserAsync(long botUin, long groupId, long userId, string name, string userOpenid = "", string groupOpenid = "")
    {
        if (userId.In(2107992324, 3677524472, 3662527857, 2174158062, 2188157235, 3375620034, 1611512438, 3227607419, 3586811032,
                    3835195413, 3527470977, 3394199803, 2437953621, 3082166471, 2375832958, 1807139582, 2704647312, 1420694846, 3788007880)) return 0;

        int i = await AppendAsync(botUin, groupId, userId, name, await GroupInfo.GetGroupOwnerAsync(groupId), userOpenid, groupOpenid);
        if (i == -1) return i;

        i = await GroupMember.AppendAsync(groupId, userId, name, "");
        if (i == -1) return i;

        if (await BotInfo.GetIsCreditAsync(botUin))
        {
            i = await Friend.AppendAsync(botUin, userId, name);
            if (i == -1) return i;
        }

        return i;
    }

    public static async Task<string> GetResetDefaultGroupAsync(long qq)
    {
        return await SetValueAsync("DefaultGroup", BotInfo.GroupCrm, qq) == -1 ? "" : $"\n默认群已重置为 {BotInfo.GroupCrm}";
    }
}
