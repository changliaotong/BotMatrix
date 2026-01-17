using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;

namespace BotWorker.Modules.Games
{
    public class GroupWarnService : IGroupWarnService
    {
        private readonly IGroupWarnRepository _repository;
        private readonly IGroupRepository _groupRepository;

        public GroupWarnService(IGroupWarnRepository repository, IGroupRepository groupRepository)
        {
            _repository = repository;
            _groupRepository = groupRepository;
        }

        public async Task<string> GetEditKeywordAsync(long groupId, string message)
        {
            // Logic moved from GroupWarnRepository (or implementation delegated if Repository has it)
            // Ideally we move the logic here.
            // For now, if Repository has it, we can call it, BUT Repository shouldn't return strings.
            // Since I saw the logic in GroupWarnRepository, I will COPY/MOVE it here.
            // But GroupWarnRepository implementation I read earlier had the logic.
            // I should really fix GroupWarnRepository to NOT have this logic, but that requires editing it.
            // To be safe and fast, I will call the repository method if it exists, but wait,
            // I want to REMOVE the anti-pattern. The anti-pattern is static access.
            // Using Repository from Service is fine.
            // BUT Repository returning UI strings is bad separation of concerns.
            // However, to minimize changes and potential errors, I will use the code I saw in Repository
            // and put it here, and then I will try to use _groupRepository for data access.
            
            // Re-implementing logic here:
             string res = "";

            var match = message.Matches(GroupWarn.RegexCmdWarn)[0];

            string cmdName = "";
            string cmdOper = "";
            string cmdPara = "";
            string operName = "";

            if (match.Success)
            {
                cmdName = match.Groups[1].Value.Trim();
                cmdOper = match.Groups[2].Value.Trim();
                cmdPara = match.Groups[3].Value.Trim();
            }
            cmdName = cmdName.Replace("加黑", "拉黑");
            cmdName += "词";

            if (cmdOper == "") cmdOper = "+";
            cmdOper = cmdOper.Replace("＋", "+").Replace("－", "-");

            string fieldName = GetFieldName(cmdName);

            if (cmdPara == "")
                return $"命令格式：\n{cmdName} + 煞笔\n{cmdName} - 煞笔";

            if (cmdPara.Length > 10)
                return "敏感词长度不能大于10";

            string keyword = await _groupRepository.GetValueAsync(fieldName, groupId) ?? "";
            if (cmdOper == "+")
            {
                operName = "添加";
                var matches = cmdPara.Matches(GroupWarn.regexParaKeyword);
                foreach (Match ma in matches)
                {
                    string paraKey = ma.Groups["keyword"].Value.Trim();
                    keyword = keyword.Replace("\\+", "+").Replace("\\*", "*");
                    List<string> keys = [.. keyword.Split('|')];
                    if (keys.Contains(paraKey))
                        res += $"\n【{paraKey}】已存在";
                    else
                    {
                        keys.Add(paraKey);
                        keyword = string.Join(" ", [.. keys]).Trim().Replace(" ", "|");
                        res += $"\n【{paraKey}】已添加";
                    }
                }
            }
            else if (cmdOper == "-")
            {
                operName = "删除";
                var matches = cmdPara.Matches(GroupWarn.regexParaKeyword);
                foreach (Match ma in matches.Cast<Match>())
                {
                    string para_key = ma.Groups["keyword"].Value.Trim();
                    List<string> keys = [.. keyword.Split('|')];
                    if (keys.Remove(para_key))
                    {
                        keyword = string.Join(" ", [.. keys]).Trim().Replace(" ", "|");
                        res += $"\n【{para_key}】已删除";
                    }
                    else
                        res += $"\n【{para_key}】不存在";
                }
            }
            else
                return "操作符不正确";

            return await _groupRepository.SetValueAsync(fieldName, keyword, groupId) == -1
                ? $"{operName}{cmdName}{Common.Common.RetryMsg}"
                : $"{operName}{cmdName}结果：{res}";
        }

        public async Task<string> GetClearResAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 清警告 + QQ";

            if (await _repository.DeleteByGroupAndUserAsync(groupId, cmdPara.GetAtUserId()) == -1)
                return Common.Common.RetryMsg;

            return "该用户警告已清除！";
        }

        public async Task<string> GetWarnInfoAsync(long groupId, string cmdPara)
        {
             if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 查警告 + QQ";

            long userId = cmdPara.GetAtUserId();
            long count = await _repository.CountByGroupAndUserAsync(groupId, userId);
            
            return count == 0 ? "该用户没有警告记录" : $"该用户当前警告次数：{count}";
        }

        public string GetCmdName(string cmdName)
        {
            return cmdName switch
            {
                "kq" or "kaiqi" or "sz" or "shezhi" or "设置" => "开启",
                "gb" or "guanbi" => "关闭",
                _ => cmdName,
            };
        }

        public async Task<string> GetKeysSetAsync(long groupId, string cmdName = "")
        {
            string res = "";
            string[] cmdParas = { "刷屏", "图片", "网址", "脏话", "广告", "推荐群", "推荐好友", "合并转发" };
            string[] cmdParas2 = { "撤回", "扣分", "警告", "禁言", "踢出", "拉黑" };
            foreach (string cmdPara in cmdParas)
            {
                if (cmdName == "" || cmdName == cmdPara)
                {
                    res += cmdName == "" ? $"\n{cmdPara}:" : $"开启 {cmdPara}";
                    foreach (string cmdPara2 in cmdParas2)
                    {
                        if (await ExistsKeyAsync(groupId, cmdPara, cmdPara2))
                            res = cmdName == "" ? $" {cmdPara2}" : $"{cmdPara2}";
                    }
                }
            }
            return cmdName == "" ? $"群管功能设置：{res}" : res;
        }

        public async Task<bool> ExistsKeyAsync(long groupId, string cmdPara, string cmdPara2)
        {
            cmdPara = GetCmdPara(cmdPara);
            cmdPara2 = GetCmdPara(cmdPara2);
            string key_field = GetFieldName(cmdPara2);
            string keyword = await _groupRepository.GetValueAsync(key_field, groupId) ?? "";
            List<string> keys = [.. keyword.Split('|')];
            return keys.Contains(cmdPara);
        }

        private string GetFieldName(string cmdPara)
        {
            cmdPara = cmdPara.Replace("词", "");
            return cmdPara switch
            {
                "撤回" => "RecallKeyword",
                "扣分" => "CreditKeyword",
                "警告" => "WarnKeyword",
                "禁言" => "MuteKeyword",
                "踢出" => "KickKeyword", // Fixed typo: Kickkeyword -> KickKeyword (checking GroupInfo POCO, it is KickKeyword)
                "拉黑" => "BlackKeyword",
                _ => ""
            };
        }

        private string GetCmdPara(string cmdPara)
        {
             return cmdPara switch
            {
                "tp" => "图片",
                "wz" => "网址",
                "gg" => "广告",
                "zh" => "脏话",
                "qfx" => "群分享",
                "ch" => "撤回",
                "kf" => "扣分",
                "jg" => "警告",
                "jy" => "禁言",
                "tc" => "踢出",
                "jh" => "拉黑",
                "lh" => "拉黑",
                "加黑" => "拉黑",
                _ => cmdPara
            };
        }
        public string RegexReplaceKeyword(string keyword)
        {
            var replacements = new Dictionary<string, string>
            {
                { "网址", Regexs.Url2 },
                { "脏话", Regexs.DirtyWords },
                { "刷屏", "" }  // 删除“刷屏”
            };

            // 拆分为独立关键词（假设是用“|”连接的正则）
            var parts = keyword.Split('|', StringSplitOptions.RemoveEmptyEntries)
                               .Select(p => p.Trim())
                               .ToList();

            // 进行替换
            for (int i = 0; i < parts.Count; i++)
            {
                if (replacements.TryGetValue(parts[i], out var replacement))
                {
                    if (!string.IsNullOrEmpty(replacement))
                        parts[i] = replacement;
                    else
                        parts[i] = ""; // 标记删除
                }
            }

            // 过滤掉被替换为空的项
            parts = parts.Where(p => !string.IsNullOrEmpty(p)).ToList();

            // 重新拼接
            return string.Join('|', parts);
        }

        public string RegexRemove(string regexKey, string keyToRemove)
        {
            var keys = regexKey.Split('|', StringSplitOptions.RemoveEmptyEntries).ToList();

            if (keys.Remove(keyToRemove))
                return string.Join("|", keys);

            return regexKey;
        }
    }
}
