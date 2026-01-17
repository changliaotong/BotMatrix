using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations.Schema;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Models
{
    [Table("ai_agents")]
    public class Agent
    {
        public long Id { get; set; }
        public Guid Guid { get; set; } = Guid.NewGuid();
        public string Name { get; set; } = string.Empty;
        public string? Description { get; set; }
        public string? SystemPrompt { get; set; }
        public string? UserPromptTemplate { get; set; }
        public long ModelId { get; set; }
        
        // Tags 存储为 JSONB，在 C# 中作为 List<string> 使用需要转换
        public string Tags { get; set; } = "[]"; 
        public string Config { get; set; } = "{}"; // JSONB: 存储运行时参数、技能列表等
        
        public long OwnerId { get; set; }
        public bool IsPublic { get; set; }
        
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

        // 兼容旧代码的属性
        public bool IsVoice { get; set; }
        public string VoiceId { get; set; } = string.Empty;
        public string Info { get; set; } = string.Empty;

        public int tokensLimit { get; set; } = 4096;
        public int tokensOutputLimit { get; set; } = 1024;
        public int tokensTimes { get; set; } = 1;
        public int tokensTimesOutput { get; set; } = 1;
        public string Prompt => SystemPrompt ?? string.Empty;
    }

    public record AgentInfo
    {
        public Guid Guid { get; init; }
        public long Id { get; init; }
        public string Name { get; init; } = string.Empty;
        public long GroupId { get; init; }

        public AgentInfo(Guid Guid, long Id, string Name, long GroupId = 0)
        {
            this.Guid = Guid;
            this.Id = Id;
            this.Name = Name;
            this.GroupId = GroupId;
        }
    }

    public static class AgentInfos
    {
        public static readonly AgentInfo PromptAgent = new(
            Guid: Guid.Parse("CEC8423E-F36B-1410-8AEF-0025F3E1B0BD"),
            Id: 7,
            Name: "智能体生成器"
        );

        public static readonly AgentInfo InfoAgent = new(
            Guid: Guid.Parse("54CA423E-F36B-1410-8AEF-0025F3E1B0BD"),
            Id: 85,
            Name: "GPT-4o"
        );

        public static readonly AgentInfo DefaultAgent = new(
            Guid: Guid.Parse("59ca423e-f36b-1410-8aef-0025f3e1b0bd"),
            Id: 86,
            Name: "早喵",
            GroupId: 10084
        );

        public static readonly AgentInfo DallEAgent = new(
            Guid: Guid.Parse("F1C8423E-F36B-1410-8AEF-0025F3E1B0BD"),
            Id: 14,
            Name: "文生图提示词生成器"
        );
    }
}
