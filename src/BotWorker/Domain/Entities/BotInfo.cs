using System.Collections.Concurrent;
using System.Data;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities
{
    public partial class BotInfo : MetaData<BotInfo>
    {
        public override string TableName => "Member";
        public override string KeyField => "BotUin";

        public const long AdminUin = 51437810;    //客服
        public const long AdminUin2 = 1653346663;  //客服
        public const long BotUinDef = 3889418604;
        public const string BotNameDef = "早喵";
        public const long MonitorGroupUin = 290581239;   //500人监控群
        public const long MonitorUin = 51437810;     //监控QQ 004号        
        public const int SystemUid = 10; //自助开通业务员工ID
        public const int WebUid = 11; //网页中自助开通的员工ID
        public const long GroupCrm = 251026193; //官方客服群
        public const long CrmUin = 1653346663;
        public const long NimingUin = 1000000;
        public const long NimingUin2 = 80000000;
        public const long DallePromptGroup = 308;
        public const long GroupIdDef = 86433316;
        public const long DefaultProxyBotUin = BotUinDef;
        public const long ProxyBotUinTest = 3889420782;
        public const long DefaultGroupUinGuild = 990000000003;
        public const long DefaultRobotId = 1098299491;
        public const long MusicGroup = 903734128;
        public static int CTimes { get; set; } = 0;
        public static ConcurrentDictionary<string, string> DictTimes { get; set; } = [];
        public static DateTime HeartbetTime { get; set; } = DateTime.Now;

        public long BotUin { get; set; }
        public string Password { get; set; } = string.Empty;
        public string BotName { get; set; } = "早喵";
        public int BotType { get; set; }
        public long AdminId { get; set; }
        public DateTime InsertDate { get; set; }
        public string BotMemo { get; set; } = string.Empty;
        public string WemcomeMessage { get; set; } = string.Empty;
        public string ApiIP { get; set; } = string.Empty;
        public string ApiPort { get; set; } = string.Empty;
        public string ApiKey { get; set; } = string.Empty;
        public string WebUIToken { get; set; } = string.Empty;
        public string WebUIPort { get; set; } = string.Empty;
        public bool IsSignalR { get; set; } = false;
        public bool IsCredit { get; set; } = false;
        public string Platform => Platforms.ToPlatform(BotType);
        public bool IsGroup { get; set; }
        public bool IsPrivate { get; set; }

        [JsonIgnore]
        public DateTime ValidDate { get; set; }
        [JsonIgnore]
        public DateTime LastDate { get; set; }
        public int Valid { get; set; }
        public bool IsFreeze { get; set; }
        [JsonIgnore]
        public DateTime FreezeDate { get; set; }
        public int FreezeTimes { get; set; }
        public bool IsBlock { get; set; }
        [JsonIgnore]
        public DateTime BlockDate { get; set; }
        [JsonIgnore]
        public DateTime HeartbeatDate { get; set; }
        [JsonIgnore]
        public DateTime ReceiveDate { get; set; }
        public bool IsVip { get; set; }
        public static ConcurrentDictionary<long, bool> IsActive { get; set; } = [];


        public static bool GetIsCredit(long botUin)
            => GetIsCreditAsync(botUin).GetAwaiter().GetResult();

        public static async Task<bool> GetIsCreditAsync(long botUin)
        {
            return await GetBoolAsync("IsCredit", botUin);
        }

        // 超级管理员
        public static bool IsSuperAdmin(long user)
        {
            return user.In(AdminUin, AdminUin2);
        }

        // 是否机器人管理员
        public static bool IsAdmin(long botUin, long userId)
            => IsAdminAsync(botUin, userId).GetAwaiter().GetResult();

        public static async Task<bool> IsAdminAsync(long botUin, long userId)
        {
            return IsSuperAdmin(userId) || await GetRobotAdminAsync(botUin) == userId;
        }

        // 小号机器人 同主人并且使用该小号的群 可以使用终极模式。
        public static long GetRobotAdmin(long botUin)
            => GetRobotAdminAsync(botUin).GetAwaiter().GetResult();

        public static async Task<long> GetRobotAdminAsync(long botUin)
        {
            return await GetLongAsync("AdminId", botUin);
        }

        public static string GetBotGuid(long botUin)
        {
            return GetValue("BotGuid", botUin);
        }

        public static async Task<string> GetBotGuidAsync(long botUin, IDbTransaction? trans = null)
        {
            return await GetValueAsync("BotGuid", botUin, null, trans);
        }

        public static bool IsRobot(long qq)
        {
            return !IsMonitorQQ(qq) && ExistsWhere($"valid not in (0,4,5) and {Key} = {qq}");
        }

        public static bool IsMonitorQQ(long botUin)
        {
            return botUin == MonitorUin;
        }
    }
}
