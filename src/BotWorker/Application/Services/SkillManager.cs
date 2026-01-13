namespace BotWorker.Application.Services
{
    public interface ISkillManager
    {
        Task<IEnumerable<Skill>> GetSkillsAsync();
        Task RegisterSkillAsync(Skill skill);
    }

    public class Skill
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public SkillCapability Capability { get; set; } = new();
    }

    public class SkillManager : ISkillManager
    {
        private readonly List<Skill> _skills = new();

        public Task<IEnumerable<Skill>> GetSkillsAsync()
        {
            return Task.FromResult<IEnumerable<Skill>>(_skills);
        }

        public Task RegisterSkillAsync(Skill skill)
        {
            if (!_skills.Any(s => s.Name == skill.Name))
            {
                _skills.Add(skill);
            }
            return Task.CompletedTask;
        }
    }
}


