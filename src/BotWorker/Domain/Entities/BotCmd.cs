using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("Cmd")]
    public partial class BotCmd
    {
        private static IBotCmdRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBotCmdRepository>() 
            ?? throw new InvalidOperationException("IBotCmdRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        [Key]
        public long Id { get; set; }
        public string CmdName { get; set; } = string.Empty;
        public string CmdText { get; set; } = string.Empty;
        public int IsClose { get; set; }

        private static readonly Dictionary<string, string> _baseCommandMap = new(StringComparer.OrdinalIgnoreCase)
        {
            { "帮助", "帮助" }, { "help", "帮助" }, { "指令", "帮助" },
            { "签到", "签到" }, { "checkin", "签到" },
            { "计算", "计算" }, { "calc", "计算" },
            { "钓鱼", "钓鱼" }, { "fish", "钓鱼" },
            { "抛竿", "抛竿" }, { "收竿", "收竿" },
            { "购买", "购买" }, { "buy", "购买" },
            { "买分", "买分" },
            { "我的宠物", "我的宠物" }, { "pet", "我的宠物" },
            { "我的待办", "todo" }, { "todo", "todo" }, { "td", "todo" },
            { "添加待办", "todo" },
            { "拍砖", "拍砖" },
            { "早安", "早安" }, { "午安", "午安" }, { "晚安", "晚安" },
            { "闲聊", "闲聊" }, { "chat", "闲聊" }, { "ai", "闲聊" },
            { "成语", "成语" },
            { "点歌", "点歌" },
        };

        private static readonly HashSet<string> _extraCommandKeywords = new(StringComparer.OrdinalIgnoreCase);

        public static void RegisterExtraCommands(IEnumerable<string> commands)
        {
            foreach (var cmd in commands)
            {
                _extraCommandKeywords.Add(cmd);
            }
        }

        public static async Task<string> GetRegexCmdAsync()
        {
            var dbCommands = await Repository.GetAllCommandNamesAsync();
            var allCommands = dbCommands
                .SelectMany(c => c.Split('|'))
                .Where(c => !string.IsNullOrEmpty(c))
                .Concat(_baseCommandMap.Keys)
                .Concat(_extraCommandKeywords)
                .Distinct()
                .OrderByDescending(cmd => cmd.Length)
                .ToArray();

            return @$"^[#＃﹟/／ ]*(?<cmdName>({string.Join('|', allCommands)}))\s*(?<cmdPara>.*)";            
        }

        public static async Task<string> GetCmdNameAsync(string cmdText)
        {
            if (string.IsNullOrEmpty(cmdText)) return "";

            // 优先从基础命令映射中查找
            if (_baseCommandMap.TryGetValue(cmdText, out var baseName))
                return baseName;

            // 再从数据库中查找
            return await Repository.GetCmdNameAsync(cmdText) ?? "";
        }

        public static async Task<string> GetClosedCmdAsync()
        {
            var closed = await Repository.GetClosedCommandsAsync();
            var res = string.Join(" ", closed);
            return string.IsNullOrEmpty(res) ? "没有功能被关闭" : res;
        }

        public static async Task<bool> IsClosedCmdAsync(long groupId, string message)
        {
            var regex = (await GroupRepository.GetClosedRegexAsync(groupId))?.Trim();
            return !string.IsNullOrEmpty(regex) && message.RemoveQqAds().IsMatch(regex);
        }

        public static async Task<bool> IsCmdCloseAllAsync(string cmdName)
        {
            return await Repository.IsCmdCloseAllAsync(cmdName);
        }

        public static async Task<string> GetCmdTextAsync(string cmdName)
        {
            return await Repository.GetCmdTextAsync(cmdName);
        }

        public static async Task EnsureCommandExistsAsync(string name, string text)
        {
            await Repository.EnsureCommandExistsAsync(name, text);
        }
    }
}
