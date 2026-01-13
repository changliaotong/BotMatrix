using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface ISkillDefinitionRepository : IRepository<SkillDefinition, long>
    {
        Task<SkillDefinition?> GetByKeyAsync(string skillKey);
        Task<IEnumerable<SkillDefinition>> GetByKeysAsync(IEnumerable<string> skillKeys);
        Task<IEnumerable<SkillDefinition>> GetByActionAsync(string actionName);
    }
}
