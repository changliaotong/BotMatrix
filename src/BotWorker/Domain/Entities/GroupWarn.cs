using System.Text.RegularExpressions;
using BotWorker.Bots.Entries;
using BotWorker.common;
using BotWorker.BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Domain.Entities
{
    public class GroupWarn : MetaData<GroupWarn>
    {
        public override string TableName => "Warn";
        public override string KeyField => "Id";
        public static string RegexCmdWarn => @"^[#＃﹟]?(撤回|扣分|警告|禁言|踢出|拉黑|加黑)词 *([＋－+-]*) *([\s\S]*)$";
        public const string regexParaKeyword = @"(?<keyword>[^ ]+[\s\S]*?[ $]*)";
        public const string regexQqImage = @"\[Image[\d\w {}-]*(.(jpg|png))*]";


        // 敏感词管理
        public static string GetEditKeyword(long groupID, string message)
        {
            string res = "";

            var match = message.Matches(RegexCmdWarn)[0];

            string cmdName = "";
            string cmdOper = "";
            string cmdPara = "";
            string operName = "";

            if (match.Success)
            {
                cmdName = match.Groups[1].Value.Trim();  // 获取命令名称，例如 "禁言"
                cmdOper = match.Groups[2].Value.Trim();  // 获取操作符，例如 "+"
                cmdPara = match.Groups[3].Value.Trim();  // 获取参数内容，例如 "参数内容"
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

            //增加敏感词
            string keyword = GroupInfo.GetValue(fieldName, groupID);
            if (cmdOper == "+")
            {
                operName = "添加";
                var matches = cmdPara.Matches(regexParaKeyword);
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
                var matches = cmdPara.Matches(regexParaKeyword);
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

            return GroupInfo.SetValue(fieldName, keyword, groupID) == -1
                ? $"{operName}{cmdName}{RetryMsg}"
                : $"{operName}{cmdName}结果：{res}";
        }

        public static string ImageUrl(long groupId)
        {
            var keywords = new[]
{
                ("RecallKeyword", "撤回词"),
                ("CreditKeyword", "扣分词"),
                ("WarnKeyword", "警告词"),
                ("MuteKeyword", "禁言词"),
                ("KickKeyword", "踢出词"),
                ("BlackKeyword", "拉黑词")
            };

            var res = string.Join("\n\n", keywords.Select(k => $"{k.Item2}：{RegexRemove(GroupInfo.GetValue(k.Item1, groupId)).Replace("|", " ").WrapWord(45)}"));
            return ImageGen.ImageUrl(res);
        }

        public static string RegexReplaceKeyword(string keyword)
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


        public static bool RegexExists(string regex_key, string key)
        {
            List<string> keys = new(regex_key.Split('|'));
            return keys.Contains(key);
        }

        public static string RegexAdd(string regex_key, string key)
        {
            List<string> keys = [.. regex_key.Split('|')];
            if (keys.Contains(key))
                return regex_key;
            else
            {
                keys.Add(key);
                return string.Join(" ", [.. keys]).Trim().Replace(" ", "|");
            }
        }

        public static string RegexRemove(string regexKey)
        {
            if (string.IsNullOrWhiteSpace(regexKey))
                return string.Empty;

            string[] cmdParas = { "刷屏", "图片", "网址", "脏话", "广告", "推荐群", "推荐好友", "合并转发" };
            List<string> keys = [.. regexKey.Split('|', StringSplitOptions.RemoveEmptyEntries)];

            foreach (string cmdPara in cmdParas)
            {
                keys.RemoveAll(k => k == cmdPara);
            }

            return string.Join("|", keys).Trim('|');
        }


        public static string RegexRemove(string regexKey, string keyToRemove)
        {
            var keys = regexKey.Split('|', StringSplitOptions.RemoveEmptyEntries).ToList();

            if (keys.Remove(keyToRemove))
                return string.Join("|", keys);

            return regexKey;
        }


        // 警告系统设置命令
        public static string GetCmdName(string cmdName)
        {
            return cmdName switch
            {
                "kq" or "kaiqi" or "sz" or "shezhi" or "设置" => "开启",
                "gb" or "guanbi" => "关闭",
                _ => cmdName,
            };
        }

        // 参数
        public static string GetCmdPara(string cmdPara)
        {
            //图片 网址 广告 脏话 
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

        public static bool ExistsKey(string keyword, string cmd_para, string cmd_para2)
        {
            cmd_para = GetCmdPara(cmd_para);
            cmd_para2 = GetCmdPara(cmd_para2);

            if (string.IsNullOrWhiteSpace(keyword))
                return false;

            List<string> keys = [.. keyword.Split('|', StringSplitOptions.RemoveEmptyEntries)];

            return keys.Contains(cmd_para);
        }


        // 开启 图片禁言
        public static bool ExistsKey(long group_id, string cmdPara, string cmdPara2)
        {
            cmdPara = GetCmdPara(cmdPara);
            cmdPara2 = GetCmdPara(cmdPara2);
            string key_field = GetFieldName(cmdPara2);
            string keyword = GroupInfo.GetValue(key_field, group_id);
            List<string> keys = [.. keyword.Split('|')];
            return keys.Contains(cmdPara);
        }

        public static string GetFieldName(string cmdPara)
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

        public static string GetKeysSet(long group_id, string cmdName = "")
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
                        if (ExistsKey(group_id, cmdPara, cmdPara2))
                            res = cmdName == "" ? $" {cmdPara2}" : $"{cmdPara2}";
                    }
                }
            }
            return cmdName == "" ? $"群管功能设置：{res}" : res;
        }

        // 清除警告
        public static string GetClearRes(long groupId, string cmdPara)
        {
            if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 清警告 + QQ";

            if (ClearWarn(groupId, cmdPara.GetAtUserId()) == -1)
                return  RetryMsg;

           return "该用户警告已清除！";
        }

        // 查警告
        public static string  GetWarnInfoAsync(long groupId, string cmdPara)
        {
            if (!cmdPara.IsMatchQQ())
                return "格式不正确，请发送 清警告 + QQ";
            long warn_qq = cmdPara.GetAtUserId();
            return $"群成员[@:{warn_qq}]警告次数:{WarnCount(warn_qq, groupId)}";
        }



        public static long WarnCount(long userId, long groupId)
        {
            return CountWhere($"GroupId = {groupId} and UserId = {userId}");
        }

        public static int AppendWarn(long botUin, long userId, long groupId, string warnInfo, long insertBy)
        {
            return Insert([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("UserId", userId),
                            new Cov("WarnInfo", warnInfo),
                            new Cov("InsertBy", insertBy),
                        ]);
        }

        public static int ClearWarn(long groupId, long qq)
        {
            return Delete(groupId, qq);
        }

    }
}
