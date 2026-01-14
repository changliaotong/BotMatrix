using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations.Schema;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.DependencyInjection;

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

        // 兼容旧代码的属性和方法
        public bool IsVoice { get; set; }
        public string VoiceId { get; set; } = string.Empty;
        public string Info { get; set; } = string.Empty;

        public Guid GetGuid() => Guid;
        public long GetId() => Id;
        public int GetSqlPlusCount() => 0; // 占位符
        public int tokensLimit { get; set; } = 4096;
        public int tokensOutputLimit { get; set; } = 1024;
        public int tokensTimes { get; set; } = 1;
        public int tokensTimesOutput { get; set; } = 1;
        public string Prompt => SystemPrompt ?? string.Empty;

        public static async Task<Agent?> LoadAsync(long id)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            return await repo.GetByIdAsync(id);
        }

        public static async Task<Agent?> LoadAsync(Guid guid)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            return await repo.GetByGuidAsync(guid);
        }

        public static T? GetWhere<T>(string field, string where)
        {
            // 这是一个简化实现，用于兼容旧代码
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            
            // 处理常见的模式: Name = 'xxx' and private <> 2
            if (where.Contains("Name =") && typeof(T) == typeof(Guid))
            {
                var parts = where.Split('\'');
                if (parts.Length >= 2)
                {
                    var name = parts[1];
                    var agent = repo.GetByNameAsync(name).GetAwaiter().GetResult();
                    if (agent != null) return (T)(object)agent.Guid;
                }
            }
            
            return default;
        }

        public static string QueryWhere(string field, string where, string order, string format)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();

            if (where.Contains("Name ="))
            {
                var parts = where.Split('\'');
                if (parts.Length >= 2)
                {
                    var name = parts[1];
                    var agent = repo.GetByNameAsync(name).GetAwaiter().GetResult();
                    if (agent != null)
                    {
                        return field.ToLower() switch
                        {
                            "id" => agent.Id.ToString(),
                            "guid" => agent.Guid.ToString(),
                            "name" => agent.Name,
                            "info" => agent.Info,
                            _ => string.Empty
                        };
                    }
                }
            }
            return string.Empty;
        }

        public static long GetIdByName(string name)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            var agent = repo.GetByNameAsync(name).GetAwaiter().GetResult();
            return agent?.Id ?? 0;
        }

        public static Guid GetGuid(long id)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            var agent = repo.GetByIdAsync(id).GetAwaiter().GetResult();
            return agent?.Guid ?? Guid.Empty;
        }

        public static long GetId(Guid guid)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            var agent = repo.GetByGuidAsync(guid).GetAwaiter().GetResult();
            return agent?.Id ?? 0;
        }

        public static string GetSqlPlusCount(long id, int increment)
        {
            return $"UPDATE ai_agents SET used_times = used_times + {increment} WHERE id = {id}";
        }

        public static string GetValue(string field, long id)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            var agent = repo.GetByIdAsync(id).GetAwaiter().GetResult();
            if (agent == null) return string.Empty;

            return field.ToLower() switch
            {
                "info" => agent.Info,
                "name" => agent.Name,
                "description" => agent.Description ?? string.Empty,
                "systemprompt" => agent.SystemPrompt ?? string.Empty,
                _ => string.Empty
            };
        }

        public static async Task UsedTimesIncrementAsync(long id)
        {
            using var scope = LLMApp.ServiceProvider.CreateScope();
            var repo = scope.ServiceProvider.GetRequiredService<IAgentRepository>();
            await repo.IncrementUsedTimesAsync(id);
        }

        public static void UsedTimesIncrement(long id)
        {
            UsedTimesIncrementAsync(id).GetAwaiter().GetResult();
        }
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
    }
}
