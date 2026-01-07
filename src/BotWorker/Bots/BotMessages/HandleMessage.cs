using sz84.Bots.Entries;
using sz84.Bots.Groups;
using sz84.Bots.Public;
using sz84.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using sz84.Core.Database;
using sz84.Core.MetaDatas;

namespace sz84.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 处理消息
        public async Task HandleMessageAsync()
        {
            if (UserInfo.StartWith285or300(UserId))
            {
                Reason += "[同行]";
                return;
            }

            //去掉@me
            if (IsAtMe) CurrentMessage = CurrentMessage.RemoveUserId(SelfId);
            IsAtOthers = IsGroup && CurrentMessage.RemoveQqImage().IsHaveUserId();

            //暗语天王盖地虎
            if (CurrentMessage.Contains("天王盖地虎"))
            {
                if (!IsProxyInGroup && IsRealProxy)                
                    await SendOfficalShareAsync();
                else if (IsBlackSystem)
                    Answer = "你已被列入官方黑名单";
                else if (IsGreySystem)
                    Answer = "你已被列入官方灰名单";
                else if (IsBlack)
                    Answer = "你已被列入黑名单";
                else if (IsGrey)
                    Answer = "你已被列入灰名单";
                else if (!Group.IsPowerOn && !IsGuild)
                    Answer = "机器人已关机，请先开机";
                else if (ClientPublic.SubscribeCompayPublic(UserId))
                    Answer = "✅ 你已确认身份";
                else
                    Answer = "微信搜【早喵AI】公众号，关注后留言【领积分】可领5000积分并完成身份确认";
                return;
            }

            //官方黑名单 
            if (IsBlackSystem)
            {
                IsSend = IsGuild || IsPublic;
                Answer = User.Credit > -5000 && (IsAtMe || IsGuild) ? $"你已被列入官方黑名单{(IsPublic ? $"({UserId})" : "")}{MinusCreditRes(200, "官方黑名单")}" : $""; 
                Reason += "[黑名单]";
                return;
            }
            else if (IsGreySystem)
            {
                IsSend = IsGuild || IsPublic;
                Answer = User.Credit > -5000 && (IsAtMe || IsGuild) ? $"你已被列入官方灰名单{(IsPublic ? $"({UserId})" : "")}{MinusCreditRes(200, "官方灰名单")}" : $"";
                Reason += "[灰名单]";
                return;
            }

            var isHot = IsHot();
            var isCmdMsg = CurrentMessage.IsMatch(BotCmd.GetRegexCmd());
            IsCmd = isHot || isCmdMsg;

            if (IsProxy && ProxyBotUin == 0)
            {
                if (IsWorker || IsNapCat)
                {
                    IsCancelProxy = true;
                    Reason += "[官机不在]";
                    if (IsCmd || IsAtMe)                    
                        await SendOfficalShareAsync();                    
                }
                return;
            }

            //凌晨4点零1分 数据维护
            DateTime dt = SQLConn.GetDate();
            if (!IsGuild && dt.Hour == 4 && dt.Minute < 1)
            {
                if (IsAtMe || IsCmd)
                    Answer = $"每天04:00-04:01系统维护\n{dt:MM-dd HH:mm:ss}";
                else
                {
                    Reason += "[系统维护]";
                    IsSend = false;
                }
                return;
            }

            //群黑名单，包括私聊时默认群
            if (await HandleBlackWarnAsync())
            {
                IsCancelProxy = true;
                return;
            }

            //发言次数统计
            if (GroupId != BotInfo.MonitorGroupUin)
            {
                if (GroupMsgCount.Update(SelfId, GroupId, GroupName, UserId, Name) == -1)
                    ErrorMessage("更新发言统计数据时出错。");
            }

            if (!Group.IsPowerOn && CurrentMessage == "开机")
            {
                Answer = GroupInfo.SetPowerOnOff(SelfId, GroupId, UserId, CurrentMessage);
                return;
            }

            if (IsGroup && !IsGuild)
            {  
                //通知续费
                if (IsVip && (IsCmd || IsAtMe) && GroupVip.RestDays(GroupId) < 0)
                {
                    IsCancelProxy = true;
                    Answer = $"本群机器人已过期，请及时续费";                    
                    Reason += "[通知续费]";
                    return;
                }

                //体验群未设置管理员的，仅回复关注官方公众号的人 或可改为加机器人官方账号好友的人？
                if (Group.IsSz84)
                {
                    if (SelfPerm < 2)
                        GroupInfo.SetValue("IsSz84", false, GroupId);
                    else
                    {
                        if (!UserInfo.SubscribedPublic(UserId))
                        {
                            if ((IsCmd || IsAtMe) && !CurrentMessage.IsMatch(Regexs.BindToken))
                            {
                                Answer = "请先设置我为管理员开启功能";
                                IsCancelProxy = true;
                            }
                            Reason += "[关注官号]";
                            return;
                        }
                    }
                }

                if ((IsCmd || IsAtMe) && !IsProxyInGroup && IsRealProxy)
                {
                    await SendOfficalShareAsync();
                    Reason += "[官机不在]";
                    return;
                }
            }

            var isRobotOwner = UserPerm == 0 || UserId == Group.RobotOwner;
            var isHaveUseRight = IsGuild || HaveUseRight();

            //将对消息进行处理（去掉广告，表情，繁体转换为简体），需要保留这些的功能需要调用 Message
            CurrentMessage = CurrentMessage.RemoveQqAds();
            CurrentMessage = CurrentMessage.AsJianti();

            // =========================================== 机器人关闭后依然起作用的命令 ==================================================

            bool isKick = CurrentMessage.IsMatch(Regexs.Kick);
            bool isMute = CurrentMessage.IsMatch(Regexs.Mute);
            bool isUnMute = CurrentMessage.IsMatch(Regexs.UnMute);
            bool isSetTitle = CurrentMessage.IsMatch(Regexs.SetTitle);
            
            //T人、禁言及取消禁言 
            if (isKick)
            {
                Answer = await GetKickOutAsync();
                return;
            }
            else if (isMute || isUnMute)
            {
                CmdName = $"{(isMute ? "禁言" : "取消禁言")}";
                Answer = await GetMuteResAsync();
                return;
            }
            else if (isSetTitle)
            {
                Answer = await GetSetTitleAsync(CurrentMessage.RegexGetValue(Regexs.SetTitle, "UserId").AsLong(), CurrentMessage.RegexGetValue(Regexs.SetTitle, "title"));
                return;
            }
            
            bool isCmdOpen = CurrentMessage.ToLower().In("开启", "#开启", "kq", "#kq");
            bool isCmdBlack = CurrentMessage.IsMatch(BlackList.regexBlack);
            bool isCmdKeyword = CurrentMessage.IsMatch(GroupWarn.RegexCmdWarn);

            if (isCmdOpen || isCmdBlack || isCmdKeyword)
            {
                Answer = SetupPrivate(true, false);
                if (Answer != "")
                    return;

                if (isCmdOpen && !Group.IsOpen)
                {
                    Answer = GroupInfo.GetSetRobotOpen(GroupId, "开启", "");
                    Answer += GroupId == 0 ? "\n设置群 {默认群}" : "";
                    return;
                }

                //拉黑
                if (isCmdBlack)
                {
                    (CmdName, CmdPara) = GetCmdPara(CurrentMessage, BlackList.regexBlack);
                    CmdName = CmdName.Replace("黑名单", "拉黑").Replace("加黑", "拉黑").Replace("删黑","取消拉黑");
                    Answer = await GetBlackRes();
                    Answer += GroupId == 0 ? "\n设置群 {默认群}" : "";
                    return;
                }

                //敏感词管理
                if (isCmdKeyword)
                {
                    Answer = GroupWarn.GetEditKeyword(GroupId, CurrentMessage);
                    Answer += !IsGroup ? "\n设置群 {默认群}" : "";
                    return;
                }
            }

            // =========================================== 此行前面的命令 机器人关闭后依然起作用 ======================================
            if (IsGroup)
            {
                if (!Group.IsOpen && !IsGuild)
                {
                    if (IsAtMe || IsCmd)                    
                        Answer = "机器人已关闭，请先 开启"; 

                    Reason += "[关闭]";
                    return;
                }

                if (!isHaveUseRight)
                {
                    Reason += "[使用权限]";
                    return;
                }

                //进群确认
                Answer = await GetConfirmNew();
                if (Answer != "")
                    return;

            }
            else if (!IsGuild)
            {
                if (!Group.IsValid)
                {
                    Answer = GroupVip.IsVipOnce(GroupId) 
                        ? $"群({GroupId}) 机器人已过期" 
                        : $"群({GroupId}) 机器人已过体验期";

                    Answer += UserInfo.GetResetDefaultGroup(UserId);
                    return;
                }

                if (!Group.IsOpen)
                {
                    Answer = $"群({GroupId}) 机器人已关闭";
                    if (!isRobotOwner)
                        Answer += UserInfo.GetResetDefaultGroup(UserId);
                    return;
                }

                if (!isHaveUseRight)
                {
                    Answer = $"群({GroupId})你没有使用权限，请联系群主授权";
                    if (!isRobotOwner)
                        Answer += UserInfo.GetResetDefaultGroup(UserId);
                    return;
                }
            }

            //图片 文件 视频 等其它消息类型的处理
            if (IsFile || IsVideo || IsXml || IsJson || IsKeyboard || IsLightApp || IsLongMsg || IsMarkdown || IsStream || IsVoice || IsMusic || IsPoke)
            {
                Answer = HandleOtherMessage();
                Reason += "[非文本]";
                return;
            }

            //前缀
            if (Group.IsRequirePrefix)
            {
                if (!CurrentMessage.IsMatch(Regexs.Prefix))
                {
                    Reason += "[前缀]";
                    return;
                }

                if (!IsCmd)
                    CurrentMessage = CurrentMessage[1..];
            }

            //避免与其它命令和聊天冲突
            if (isCmdMsg) 
                (CmdName, CmdPara) = GetCmdPara(CurrentMessage, BotCmd.GetRegexCmd());
            else
                CmdPara = CurrentMessage;

            if ((CmdName.In("续费", "暗恋", "换群", "换主人", "警告") && !CmdPara.IsNull() && !CmdPara.IsNum())
                || (CmdName.In("剪刀", "石头", "布", "抽奖", "三公", "红", "和", "蓝") && !CmdPara.IsNull() && (CmdPara.Trim() != "梭哈") && !CmdPara.IsNum())
                || (CmdName.In("菜单", "领积分", "签到", "爱群主", "笑话", "鬼故事", "早安", "午安", "晚安", "揍群主", "升级", "降级", "结算", "一键改名") && !CmdPara.IsNull())
                || (CmdName.In("计算") && !CmdPara.IsMatch(Regexs.Formula)))
            {
                IsCmd = false;
                isCmdMsg = false;
                CmdName = "闲聊";
                CmdPara = CurrentMessage;
            }

            if (!Message.IsNull())
            {
                var isAuto = CmdName != "签到" && !Message.Contains("签到") && !Message.Contains("打卡");
                if (isAuto)
                {
                    Answer = TrySignIn(isAuto) ?? string.Empty;
                    if (!string.IsNullOrWhiteSpace(Answer))
                    {
                        CostTime = CurrentStopwatch == null ? 0 : CurrentStopwatch.Elapsed.TotalSeconds;
                        var isCmd = IsCmd;
                        var isCancelProxy = IsCancelProxy;
                        IsCmd = true;
                        IsCancelProxy = true;
                        //IsRecall = CurrentGroup.IsRecall;
                        //RecallAfterMs = CurrentGroup.RecallTime;
                        await SendMessageAsync();
                        IsCmd = isCmd;
                        //IsRecall = false;
                        IsCancelProxy = isCancelProxy;
                        Answer = "";
                    }
                }
            }

            if (IsCmd)
            {
                if (IsRefresh) 
                    HandleRefresh();
                else if (isHot) 
                    await GetHotCmdAsync();
                else if (isCmdMsg) 
                    await GetCmdResAsync();
            }
            else
            {
                await TryParseAgentCall();
                if (IsCallAgent)
                {
                    if (CmdPara.Trim().IsNull())
                    {
                        //如果参数为空，直接切换到该智能体
                        Answer = UserInfo.SetValue("AgentId", CurrentAgent.Id, UserId) == -1
                            ? $"变身{RetryMsg}"
                            : $"【{CurrentAgent.Name}】{CurrentAgent.Info}";
                        return;
                    }
                    if (!IsWeb)  
                        await GetAgentResAsync();                    

                    return;
                }

                //确认执行命令
                await ConfirmCmdAsync();
                if (Answer != "") return;

                //默认功能：聊天/问路/翻译/逗你玩/成语接龙 群聊时不能逗你玩与自己成语接龙
                CmdName = UserInfo.GetStateRes(User.State);
                if (IsGroup && CmdName.In("逗你玩", "接龙"))
                    CmdName = "闲聊"; 
                else if (CmdName == "AI")
                    await GetAgentResAsync();                
                else if (InGame())
                {
                    CmdName = "成语接龙";
                    CmdPara = CurrentMessage;
                    Answer = await GetJielongRes();
                    if (Answer != "")
                        return;                    
                }

                CmdPara = CurrentMessage; 

                await GetCmdResAsync();

                //官方刷屏扣分/拉黑
                if (IsRefresh && !Answer.IsNull())
                    HandleRefresh();
            }
        }
    }
}
