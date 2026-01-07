using Microsoft.Data.SqlClient;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Agents.Entries
{
    //智能体
    public partial class Agent : MetaDataGuid<Agent>
    {
        public override string TableName => "Agents";
        public override string KeyField => "Id";
        public string Name { get; set; } = string.Empty;
        public string Prompt { get; set; } = string.Empty;        
        public string Info { get; set; } = string.Empty;
        public string? Memo { get; set; } = string.Empty;
        [DbIgnore]
        public List<string> Tags { get; set; } = [];
        public int ModelId { get; set; } = 1;
        public bool IsVoice { get; set; } = true;
        public string? VoiceLang { get; set; }
        public string? VoiceName { get; set; }  
        public double VoiceRate { get; set; } = 2.0;
        public string VoiceId { get; set; } = string.Empty;
        public int Private { get; set; } = 0; //0 公开 1 不可见但可用 2 私密-仅自己可用
        public bool IsPublic => Private == 0;
        public long GroupId => Guid == AgentInfos.DefaultAgent.Guid ? AgentInfos.DefaultAgent.GroupId : Id + groupAgent;
        public long SubscriptionCount { get; set; }
        public long ReviewCount { get; set; }
        public long UsedTimes { get; set; }
        [DbIgnore]
        public long IsSub { get; set; } = 0; 
        public long UserId { get; set; }
        public string? UserPrompt { get; set; } = string.Empty;

        public static readonly long tokensLimit = 128000;
        public static readonly long tokensOutputLimit = 16384;
        public static readonly long tokensTimes = 1;
        public static readonly long tokensTimesOutput = 2;
        
        public static string NoTokensMsg => $"您的算力已用完，可每天签到获得或购买，客服QQ：{{客服QQ}}";
        public static string NoTokensMsgGuild => $"您的算力已用完，请明日再来";
             
        public const long groupAgent = 9900000000;

        public static long GetIdByName(string name)
        {
            return GetWhere<long>("Id", $"name = {name.Quotes()} and private = 0", "id desc");
        }

        public static (string, SqlParameter[]) GetSqlPlusCount(long id, int increment = 1)
        {
            return SqlPlus("SubscriptionCount", increment, id);
        }

        public static int UsedTimesIncrement(long id, int increment = 1)
        {
            var (sql, parameters) = SqlPlus("usedTimes", increment, id);
            return Exec(sql, parameters);
        }

        public static bool AgentExists(string name, long userId)
        {
            return ExistsAandB("Name", name, "userId", userId);
        }

        public static async Task<List<Agent>> GetAgents()
        {
            var sql = $"SELECT * from {FullName} WHERE Private = 0";
            return await QueryListAsync<Agent>(sql);
        }
    }

    public static class AgentInfos
    {
        public static readonly AgentInfo PromptAgent = new(
            guid: Guid.Parse("CEC8423E-F36B-1410-8AEF-0025F3E1B0BD"),
            id: 7,
            name: "智能体生成器"
        );

        public static readonly AgentInfo InfoAgent = new(
            guid: Guid.Parse("54CA423E-F36B-1410-8AEF-0025F3E1B0BD"),
            id: 85,
            name: "GPT-4o"
        );

        public static readonly AgentInfo DefaultAgent = new(
            guid: Guid.Parse("59ca423e-f36b-1410-8aef-0025f3e1b0bd"),
            id: 86,
            name: "早喵",
            groupId: 10084
        );

        public static readonly AgentInfo DallEAgent = new(
            guid: Guid.Parse("F1C8423E-F36B-1410-8AEF-0025F3E1B0BD"),
            id: 14,
            name: "文生图提示词生成器"
        );
    }

    public class AgentInfo(Guid guid, int id, string name, long? groupId = null)
    {
        public Guid Guid { get; set; } = guid;
        public int Id { get; set; } = id;
        public string Name { get; set; } = name;
        public long GroupId { get; set; } = groupId ?? 0;
    }
}
