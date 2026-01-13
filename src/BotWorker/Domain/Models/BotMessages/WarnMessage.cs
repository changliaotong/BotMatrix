using System.Text.RegularExpressions;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task<bool> HandleBlackWarnAsync()
        {
            //黑名单、敏感词系统
            if (IsGroup)
            {
                if (IsBlack && SelfPerm < UserPerm && SelfPerm < 2)
                {
                    Answer = $"黑名单成员 {UserId} 将被T出群";
                    await KickOutAsync(SelfId, RealGroupId, UserId);
                    IsRecall = Group.IsRecall;
                    RecallAfterMs = Group.RecallTime * 1000;
                    return true;
                }

                if (Group.IsWarn)
                {
                    await GetKeywordWarnAsync();
                    if (Answer != "")
                    {
                        IsRecall = Group.IsRecall;
                        RecallAfterMs = Group.RecallTime * 1000;
                        return true;
                    }
                }
            }
            else if (IsBlack)
            {
                Answer = $"您已被群({GroupId})拉黑";
                return true;
            }
            
            if (IsGrey)
                return true;

            return false;
        }

        // 警告命令
        public async Task<string> GetWarnRes()
        {
            IsCancelProxy = true;

            if (!CmdPara.IsMatchQQ())
                return "格式：警告 + QQ\n例如：警告 {客服QQ}";

            long qqWarn = CmdPara.GetAtUserId();

            if (await GroupWarn.AppendWarnAsync(SelfId, qqWarn, GroupId, "", UserId) == -1)
                return RetryMsg;

            long countWarn = await GroupWarn.WarnCountAsync(qqWarn, GroupId);
            IsCancelProxy = true;
            if (countWarn >= Group.BlackCount)
            {
                int i = await BlackList.AddBlackListAsync(SelfId, GroupId, GroupName, UserId, Name, qqWarn, "警告超限拉黑");
                if (i == -1)
                    return RetryMsg;

                i = await GroupWarn.ClearWarnAsync(GroupId, qqWarn);
                if (i == -1)
                    Logger.Error($"清空警告{RetryMsg}");

                await KickOutAsync(SelfId, RealGroupId, UserId);                
                return $"[@:{qqWarn}] 警告{countWarn}次，将拉黑并T飞！\n已拉黑！";
            }
            else if (countWarn >= Group.KickCount)
            {
                await KickOutAsync(SelfId, RealGroupId, UserId);
                return $"[@:{qqWarn}] 警告{countWarn}次，将被T飞！";
            }
            else            
                return $"[@:{qqWarn}] 警告{countWarn}次!\n{Group.KickCount}次T飞，{Group.BlackCount}次拉黑";            
        }

        public async Task<bool> IsAdInfoAsync()
        {
            //白名单 群主/管理员 返回false
            if (SelfPerm == 2 || UserPerm < 2)
                return false;

            if (IsWhiteList(UserId))
                return false;

            //包含非官方网址的==广告
            if (Message.ContainsURL())
                return true;

            //@的号码全部替换为空
            var message = Message.RemoveQqFace().RemoveQqImage().RegexReplace(Regexs.AtUsers, "");

            //包含QQ号码群号 继续判断
            if (!IsCmd && message.HaveUserId())
            {
                var matches = message.Matches(Regexs.HaveUserId);
                foreach (Match match in matches)
                {
                    //其中一个QQ是就判定为广告
                    var num = match.Groups["UserId"].Value.Trim().AsLong();
                    if (await IsAdQqAsync(num))
                        return true;
                }
            }

            //message  行数，字数，
            int c_row = message.Split('\n').Length - 1;
            int c_text = message.Length;
            int c_days = GroupMember.GetInt("ABS(DATEDIFF(DAY, GETDATE(), InsertDate))", GroupId, UserId);

            if (message.IsMatch(Regexs.AdWords))
            {
                if ((c_row > 3) | (c_text > 18) | (c_days < 3))
                    return true;
            }

            return false;
        }

        public async Task<bool> IsAdQqAsync(long num)
        {
            await Task.Yield();

            //机器人/客户/群号，所有机器人号码
            if (num == SelfId || num == GroupId || num == UserId || BotInfo.IsRobot(num))
                return false;

            //号码在群里
            if (await IsInGroupAsync(SelfId, RealGroupId, num))
                return false;

            //白名单
            if (IsWhiteList(num))
                return false;

            return true;
        }

        public async Task<(bool, string)> GetMatchAsync(bool isMatch, string regexKey)
        {
            regexKey = regexKey.ReplaceRegex();
            if ("广告".IsMatch(regexKey))
            {
                isMatch = await IsAdInfoAsync();
                regexKey = GroupWarn.RegexRemove(regexKey, "广告");
            }
            if ("图片".IsMatch(regexKey))
            {
                isMatch = isMatch || IsImage;
                regexKey = GroupWarn.RegexRemove(regexKey, "图片");
            }
            if ("推荐群".IsMatch(regexKey))
            {
                isMatch = isMatch || IsContactGroup;
                regexKey = GroupWarn.RegexRemove(regexKey, "推荐群");
            }
            if ("推荐好友".IsMatch(regexKey))
            {
                isMatch = isMatch || IsContactFriend;
                regexKey = GroupWarn.RegexRemove(regexKey, "推荐好友");
            }
            if ("合并转发".IsMatch(regexKey))
            {
                isMatch = isMatch || IsForward;
                regexKey = GroupWarn.RegexRemove(regexKey, "合并转发");
            }
            regexKey = GroupWarn.RegexReplaceKeyword(regexKey);
            isMatch = isMatch || (regexKey != "" && CurrentMessage.RemoveQqFace().RemoveQqImage().IsMatch(regexKey));
            return (isMatch, regexKey);
        }

        // 敏感词 撤回 警告 禁言 扣分 踢出 拉黑
        public async Task GetKeywordWarnAsync()
        {
            //白名单、QQ管家、没有权限
            if (IsWhiteList() || UserId == 2854196310 || SelfPerm >= UserPerm)
                return;

            var message = CurrentMessage.RemoveQqAds();
            //网址白名单
            message = message.RegexReplace(Regexs.UrlWhite, "");

            bool isMatch = false;

            //敏感词黑名单            
            string regexKey = Group.BlackKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    await BlackList.AddBlackListAsync(SelfId, GroupId, GroupName, SelfId, SelfName, UserId, "敏感词拉黑");
                    await GroupWarn.ClearWarnAsync(GroupId, UserId);
                    await RecallAsync(SelfId, RealGroupId, MsgId);
                    await KickOutAsync(SelfId, RealGroupId, UserId);
                    Answer = $"[@:{UserId}] 发言违规已拉黑";
                    GroupEvent.Append(this, $"拉黑", $"敏感词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                    return;
                }
            }

            //敏感词踢人
            regexKey = Group.KickKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    await GroupWarn.AppendWarnAsync(SelfId, UserId, GroupId, Message, SelfId);
                    await RecallAsync(SelfId, RealGroupId, MsgId);
                    await KickOutAsync(SelfId, RealGroupId, UserId);
                    Answer = $"[@:{UserId}] 发言违规将被T飞！\n警告次数+1";
                    GroupEvent.Append(this, $"踢出", $"敏感词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                    return;
                }
            }

            //敏感词禁言
            regexKey = Group.MuteKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    int time = Group.MuteKeywordCount;
                    if (time != 0)
                    {
                        await RecallAsync(SelfId, RealGroupId, MsgId);
                        await MuteAsync(SelfId, RealGroupId, UserId, time * 60);                        
                        Answer = $"[@:{UserId}] 发言违规将被禁言{time}分钟！";
                        GroupEvent.Append(this, $"禁言", $"时长：{time}分钟 敏感词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                        return;
                    }
                }
            }

            //敏感词扣分
            regexKey = Group.CreditKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    await RecallAsync(SelfId, RealGroupId, MsgId);
                    Answer = $"[@:{UserId}] 发言违规扣分";
                    (int i, long creditValue) = await MinusCreditAsync(100, "敏感词扣分");
                    if (i != -1)
                        Answer += $"\n积分：-100，累计：{creditValue}";
                    GroupEvent.Append(this, $"扣分", $"扣分：-100 敏感词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                    return;
                }
            }

            //敏感词警告
            regexKey = Group.WarnKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    await RecallAsync(SelfId, RealGroupId, MsgId);
                    Answer = await AddWarn(UserId, Name, SelfId);
                    GroupEvent.Append(this, $"警告", $"敏感词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                    return;
                }
            }

            //敏感词撤回（只撤回）
            regexKey = Group.RecallKeyword;
            if (!regexKey.IsNullOrEmpty())
            {
                (isMatch, regexKey) = await GetMatchAsync(isMatch, regexKey);
                if (isMatch)
                {
                    await RecallAsync(SelfId, RealGroupId, MsgId);
                    Answer = $"[@:{{UserId}}] 发言违规已撤回";
                    GroupEvent.Append(this, $"撤回", $"撤回词：{Regex.Match(message, regexKey).Value}\n正则：{regexKey}");
                    return;
                }
            }

            //刷屏
            if (IsRefresh)
                await GetRefreshRes();
        }

        public async Task<string> AddWarn(long TargetId, string targetName, long UserId)
        {
            if (await GroupWarn.AppendWarnAsync(SelfId, TargetId, GroupId, Message, SelfId) == -1)
                return RetryMsg;

            long countWarn = await GroupWarn.WarnCountAsync(TargetId, ParentGroup?.Id ?? 0);
            if (countWarn >= Group.BlackCount)
            {
                await KickOutAsync(SelfId, RealGroupId, TargetId);
                return AddBlack(TargetId, "警告超限拉黑") == -1
                    ? RetryMsg
                    : $"已警告{Group.KickCount}次，{TargetId}({targetName})已拉黑!";
            }
            else if (countWarn >= Group.KickCount)
            {
                await KickOutAsync(SelfId, RealGroupId, UserId);
                return $"已警告{Group.KickCount}次，{TargetId}({targetName})将被T飞！";
            }
            else
                return $"警告{countWarn}次！{Group.KickCount}次T飞，{Group.BlackCount}次拉黑";
        }
}
