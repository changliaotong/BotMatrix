using System.Text.RegularExpressions;
using sz84.Bots.Extensions;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        /*
        {+credit(10)}：加积分
        {-credit(5)}：扣积分
        {+chance(5)}：获得额外抽奖机会
        {mute(30s)}：禁言30秒（用于娱乐）
        {bonus("双倍积分券")}：获得道具 / 称号
        { set("flag_name")}：设置成就标志
        {xp(3)}：经验值 
        
        */
        public void ProcessEvent()
        {
            var pattern = @"{([+\-]?\w+)\(([^)}]*)\)}";
            var matches = Regex.Matches(Answer, pattern);

            foreach (Match match in matches)
            {
                string action = match.Groups[1].Value;
                string arg = match.Groups[2].Value;
                string replacement = "";

                switch (action.ToLower())
                {
                    case "credit":
                        int value = ParseRandomValue(arg);
                        AddCredit(value, $"");
                        replacement = $"(+{arg}💎)";
                        break;
                    case "-credit":
                        value = ParseRandomValue(arg);
                        AddCredit(-value, $"");
                        replacement = $"(-{arg}💎)";
                        break;
                    case "xp":
                        //AddXp(int.Parse(argument));
                        break;
                    case "mute":
                        //this.AddMute(UserId, ParseTime(arg).TotalSeconds.AsInt());
                        replacement = $"(禁言{arg})";
                        break;
                    case "set":
                        //this.AddSetTitle(UserId, arg);
                        replacement = $"(头衔：{arg.Trim('"')})";
                        break;
                    // 添加更多指令支持
                    default:
                        replacement = $"";
                        break;
                }

                // 替换掉指令标签（也可以选择替换为具体说明）
                Answer = Answer.Replace(match.Value, replacement);
            }
        }
        private int ParseRandomValue(string arg)
        {
            if (arg.Contains("~"))
            {
                var parts = arg.Split('~');
                int min = int.Parse(parts[0]);
                int max = int.Parse(parts[1]);
                var _random = new Random();
                return _random.Next(min, max + 1); // 包含 max
            }
            else
            {
                return int.Parse(arg);
            }
        }
        private static TimeSpan ParseTime(string input)
        {
            if (input.EndsWith('s')) return TimeSpan.FromSeconds(int.Parse(input.TrimEnd('s')));
            if (input.EndsWith('m')) return TimeSpan.FromMinutes(int.Parse(input.TrimEnd('m')));
            if (input.EndsWith('h')) return TimeSpan.FromHours(int.Parse(input.TrimEnd('h')));
            return TimeSpan.Zero;
        }
    }

}
