using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    public class BotCmdService : IBotCmdService
    {
        private readonly IBotCmdRepository _botCmdRepository;
        private readonly IGroupRepository _groupRepository;
        
        private string _regexCmd = "";
        private string _closedCmd = "";
        private readonly List<string> _extraCommandKeywords = new();
        private readonly Dictionary<string, string> _baseCommandMap = new(StringComparer.OrdinalIgnoreCase)
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

        public BotCmdService(IBotCmdRepository botCmdRepository, IGroupRepository groupRepository)
        {
            _botCmdRepository = botCmdRepository;
            _groupRepository = groupRepository;
        }

        public async Task<string> GetRegexCmdAsync()
        {
            if (!string.IsNullOrEmpty(_regexCmd)) return _regexCmd;

            var dbCommands = await _botCmdRepository.GetAllCommandNamesAsync();
            var allCommands = dbCommands
                .SelectMany(c => c.Split('|'))
                .Where(c => !string.IsNullOrEmpty(c))
                .Concat(_baseCommandMap.Keys)
                .Concat(_extraCommandKeywords)
                .Distinct()
                .OrderByDescending(cmd => cmd.Length)
                .ToArray();

            _regexCmd = @$"^[#＃﹟/／ ]*(?<cmdName>({string.Join('|', allCommands)}))\s*(?<cmdPara>.*)";
            return _regexCmd;
        }

        public async Task<string> GetCmdNameAsync(string cmdText)
        {
            if (string.IsNullOrEmpty(cmdText)) return "";

            // 优先从基础命令映射中查找
            if (_baseCommandMap.TryGetValue(cmdText, out var baseName))
                return baseName;

            // 再从数据库中查找
            return await _botCmdRepository.GetCmdNameAsync(cmdText) ?? "";
        }

        public async Task<string> GetClosedCmdAsync()
        {
            if (!string.IsNullOrEmpty(_closedCmd)) return _closedCmd;

            var closed = await _botCmdRepository.GetClosedCommandsAsync();
            _closedCmd = string.Join(" ", closed);
            return string.IsNullOrEmpty(_closedCmd) ? "没有功能被关闭" : _closedCmd;
        }

        public async Task<bool> IsClosedCmdAsync(long groupId, string message)
        {
            var regex = (await _groupRepository.GetClosedRegexAsync(groupId))?.Trim();
            return !string.IsNullOrEmpty(regex) && message.RemoveQqAds().IsMatch(regex);
        }

        public async Task<bool> IsCmdCloseAllAsync(string cmdName)
        {
            return await _botCmdRepository.IsCmdCloseAllAsync(cmdName);
        }

        public async Task<string> GetCmdTextAsync(string cmdName)
        {
            return await _botCmdRepository.GetCmdTextAsync(cmdName);
        }

        public async Task EnsureCommandExistsAsync(string name, string text)
        {
            await _botCmdRepository.EnsureCommandExistsAsync(name, text);
            _regexCmd = ""; // 重置缓存
        }

        public async Task<int> SetCmdCloseAllAsync(string cmdName, int isClose)
        {
            var result = await _botCmdRepository.SetCmdCloseAllAsync(cmdName, isClose);
            _regexCmd = ""; // 重置缓存
            _closedCmd = ""; // 重置缓存
            return result;
        }

        public void RegisterExtraCommands(IEnumerable<string> commands)
        {
            foreach (var cmd in commands)
            {
                if (!_extraCommandKeywords.Contains(cmd))
                {
                    _extraCommandKeywords.Add(cmd);
                }
            }
            _regexCmd = ""; // 重置缓存
        }
    }
}
