using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.SemanticKernel;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.Plugins;
using System.ComponentModel;

namespace BotWorker.Modules.AI.Plugins
{
    public class BotSkillPlugin
    {
        private readonly IRobot _robot;
        private readonly IPluginContext _context;

        public BotSkillPlugin(IRobot robot, IPluginContext context)
        {
            _robot = robot;
            _context = context;
        }

        [KernelFunction]
        [Description("调用机器人的本地技能或命令。例如：2048游戏、积分查询、天气查询等。")]
        public async Task<string> CallSkill(
            [Description("技能名称或命令名称")] string skillName,
            [Description("传递给技能的参数字符串（空格分隔）")] string args = ""
        )
        {
            var argArray = args.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
            return await _robot.CallSkillAsync(skillName, _context, argArray);
        }

        [KernelFunction]
        [Description("列出机器人当前可用的所有技能和命令。")]
        public string ListSkills()
        {
            var skills = _robot.Skills;
            var result = "当前可用技能：\n";
            foreach (var skill in skills)
            {
                result += $"- {skill.Capability.Name}: {skill.Capability.Description}\n";
                if (skill.Capability.Commands.Any())
                {
                    result += $"  命令: {string.Join(", ", skill.Capability.Commands)}\n";
                }
            }
            return result;
        }
    }
}
