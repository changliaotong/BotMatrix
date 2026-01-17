using System;
using System.Collections.Generic;
using System.Linq;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupWarnRepository : BaseRepository<GroupWarn>, IGroupWarnRepository
    {
        private readonly IGroupRepository _groupRepository;

        public GroupWarnRepository(IGroupRepository groupRepository, string? connectionString = null) : base("Warn", connectionString)
        {
            _groupRepository = groupRepository;
        }

        public async Task<long> CountByGroupAndUserAsync(long groupId, long userId)
        {
            using var conn = CreateConnection();
            string sql = $"SELECT COUNT(*) FROM {_tableName} WHERE \"GroupId\" = @groupId AND \"UserId\" = @userId";
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<int> DeleteByGroupAndUserAsync(long groupId, long userId)
        {
            using var conn = CreateConnection();
            string sql = $"DELETE FROM {_tableName} WHERE \"GroupId\" = @groupId AND \"UserId\" = @userId";
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }

        public async Task<string> GetEditKeywordAsync(long groupID, string message)
        {
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

            string keyword = await _groupRepository.GetValueAsync(fieldName, groupID) ?? "";
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

            return await _groupRepository.SetValueAsync(fieldName, keyword, groupID) == -1
                ? $"{operName}{cmdName}{Common.Common.RetryMsg}"
                : $"{operName}{cmdName}结果：{res}";
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
                "踢出" => "Kickkeyword",
                "拉黑" => "BlackKeyword",
                _ => ""
            };
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

        public string GetCmdPara(string cmdPara)
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

        public async Task<string> GetKeysSetAsync(long group_id, string cmdName = "")
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
                        if (await ExistsKeyAsync(group_id, cmdPara, cmdPara2))
                            res = cmdName == "" ? $" {cmdPara2}" : $"{cmdPara2}";
                    }
                }
            }
            return cmdName == "" ? $"群管功能设置：{res}" : res;
        }

        public async Task<bool> ExistsKeyAsync(long group_id, string cmdPara, string cmdPara2)
        {
            cmdPara = GetCmdPara(cmdPara);
            cmdPara2 = GetCmdPara(cmdPara2);
            string key_field = GetFieldName(cmdPara2);
            string keyword = await _groupRepository.GetValueAsync(key_field, group_id) ?? "";
            List<string> keys = [.. keyword.Split('|')];
            return keys.Contains(cmdPara);
        }

        public async Task<string> GetClearResAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 清警告 + QQ";

            if (await DeleteByGroupAndUserAsync(groupId, cmdPara.GetAtUserId()) == -1)
                return Common.Common.RetryMsg;

            return "该用户警告已清除！";
        }

        public async Task<string> GetWarnInfoAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 清警告 + QQ";
            long warn_qq = cmdPara.GetAtUserId();
            return $"群成员[@:{warn_qq}]警告次数:{await CountByGroupAndUserAsync(groupId, warn_qq)}";
        }

        public async Task<int> AppendWarnAsync(long botUin, long userId, long groupId, string warnInfo, long insertBy)
        {
            var warn = new GroupWarn
            {
                BotUin = botUin,
                GroupId = groupId,
                UserId = userId,
                WarnInfo = warnInfo,
                InsertBy = insertBy,
                InsertDate = DateTime.Now
            };
            return await AddAsync(warn);
        }
    }
}
