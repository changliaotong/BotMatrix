using Microsoft.SemanticKernel;
using System.ComponentModel;
using static BotWorker.Infrastructure.Tools.Retirement;

namespace BotWorker.Modules.AI.Plugins
{
    internal class RetirementPlugin
    {
        [KernelFunction("RetirementAge")]
        [Description("计算用户的退休年龄和退休时间，以及延迟的月份.")]
        public static string RetirementAge(DateTime birthday, Gender gender, Cadre cadre = Cadre.未知)
        {
            return CalculateRetirement(birthday, gender, cadre) + "延迟退休月份请务必加上。";
        }
    }
}
