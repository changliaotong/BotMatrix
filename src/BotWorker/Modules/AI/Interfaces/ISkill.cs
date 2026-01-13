using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ISkill
    {
        string Name { get; }
        string Description { get; }
        string[] SupportedActions { get; }
        Task<string> ExecuteAsync(string action, string target, string reason, Dictionary<string, string> metadata);
    }

    public interface ISkillService
    {
        Task<string> ExecuteSkillAsync(string skillName, string target, string parameter, Dictionary<string, string> metadata);
        IEnumerable<ISkill> GetAvailableSkills();
    }
}
