using sz84.Bots.Entries;
using sz84.Bots.Groups;
using sz84.Bots.Users;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {        
        public async Task HandleEventAsync()
        {
            //官机openid处理
            (var isNewGroup, var isBot) = await HandleGuildMessageAsync();
            if (isBot)
            {
                IsSend = false;
                Reason += "[机器人]";
                return;
            }

            if (!IsGuild)
            {
                UserInfo.AppendUser(SelfId, GroupId, UserId, Name);
                if (IsAnotherRobot() && EventType.In("FriendMessageEvent", "GroupMessageEvent")) return;

                if (UserId.In(BotInfo.NimingUin, BotInfo.NimingUin2))
                {
                    Reason += "[匿名]";
                    return;
                }
            }

            User = await UserInfo.LoadAsync(UserId);
            
            if (IsGroup)
                GroupInfo.Append(GroupId, GroupName, SelfId, SelfName, UserId, UserId);

            RealGroupId = Group.Id;
            RealGroupName = Group.GroupName;
            Group = await GroupInfo.LoadAsync(!IsGroup ? User.DefaultGroup : GroupId);
            if (Group.ParentGroup != 0)            
                ParentGroup = await GroupInfo.LoadAsync(Group.ParentGroup);            

            IsProxy = IsGroup && !IsGuild && Group.IsProxy;
            CurrentMessage = Message ?? "";

            if (User.IsLog) BotLog.Log($"{GroupName}({GroupId}) {Name}({UserId}) {EventType}：\n{CurrentMessage}", "处理前", this);

            if (IsGroup)
            {
                IsVip = GroupVip.IsVip(GroupId);

                if ((IsNapCat || IsWeixin || IsWorker) && !Group.IsValid && !GroupInfo.IsCanTrial(GroupId))
                {
                    Answer = $"{(GroupVip.IsVipOnce(GroupId) ? "已过期" : "体验期已过")}退群";
                    IsCancelProxy = true;
                    await SendMessageAsync();
                    await LeaveAsync(SelfId, RealGroupId);
                    return;
                }                
            }
            
            CurrentMessage = CurrentMessage.Trim();   
            IsRobot = BotInfo.IsRobot(UserId);
            //IsSystemWhite = WhiteList.IsSystemWhite(UserId);         
            IsBlackSystem = BlackList.IsSystemBlack(UserId);
            IsGreySystem = GreyList.IsSystemGrey(UserId);            
            IsWhite = WhiteList.Exists(GroupId, UserId);
            IsBlack = !IsRobot && !IsWhite && ((IsBlackSystem && Group.IsCloudBlack) || BlackList.Exists(GroupId, UserId));
            IsGrey = !IsRobot && !IsWhite && GreyList.Exists(GroupId, UserId);

            if (SelfInfo.BotType == 8 && !EventType.In("EventNoticeGroupIncrease")) return;

            //if (UserId.In(2107992324,3677524472,3662527857,2174158062,2188157235,3375620034,1611512438,3227607419,3586811032,
            //    3835195413,3527470977,3394199803,2437953621,3082166471,2375832958,1807139582,2704647312,1420694846,3788007880)) return;

            //关机时：开机命令、官机、私聊、解除禁言可用 机器人进群事件（进群自动开机）
            if (IsGuild || !IsGroup || CurrentMessage.Trim() == "开机" || EventType.In("GroupMemberMuteEvent", "JoinedEvent") || Group.IsPowerOn)
            {
                switch (EventType)
                {
                    //关注公众号
                    case "PUBLIC_SUBSCRIBE":
                        OnPublicSubscribe();
                        break;
                    //取消关注公众号
                    case "PUBLIC_UNSUBSCRIBE":
                        OnPublicUnSubscribe();
                        break;
                    //机器人上线、重连
                    case "BotOnlineEvent" or "OnlineEvent" or "ReconnectedEvent" or "OnReadyEvent":
                        OnBotOnline();
                        break;
                    //加机器人好友（官机）
                    case "FRIEND_ADD":
                        await OnFriendAddAsync();
                        break;
                    //删除机器人好友（官机）
                    case "FRIEND_DEL":
                        await OnFriendDelAsync();
                        break;
                    //新人加入频道
                    case "GUILD_MEMBER_ADD":
                        await OnGuildMemberAddAsync();
                        break;
                    //被邀请进群（官机）
                    case "GROUP_ADD_ROBOT":
                        await OnGroupAddRobotAsync(isNewGroup);
                        break;
                    //被踢出群（官机）
                    case "GROUP_DEL_ROBOT":
                        await OnGroupDelRobot();
                        break;
                    //机器人进群
                    case "JoinedEvent":
                        GetJoinedRes();
                        break;
                    //群成员增加
                    case "EventNoticeGroupIncrease" or "MemberJoinedEvent":
                        await OnGroupIncreaseAsync();
                        break;
                    //群成员减少、被踢
                    case "EventNoticeGroupDecrease" or "KickedEvent" or "MemberKickedEvent":
                        OnGroupMemberDecrease();
                        break;
                    //群成员主动离开
                    case "MemberLeftEvent":
                        OnMemberLeft();
                        break;
                    //好友申请
                    case "EventNoticeFriendAdd":
                        if (!IsBlackSystem &&　!IsGreySystem)
                            Accept = 1;
                        break;
                    //新成员加群申请
                    case "EventRequestGroup":
                        OnEventRequestGroup();
                        break;
                    //禁言
                    case "MemberMutedEvent":
                        await OnGroupMemberMute();
                        break;
                    //全员禁言
                    case "GroupMutedAllEvent":
                        OnMuteAll();
                        break;
                    //群消息撤回
                    case "EventNoticeGroupRecall":
                        await OnGroupRecallAsync();
                        break;
                    //好友撤回消息
                    case "EventNoticeFriendRecall":
                        await OnFriendRecallAsync();
                        break;
                    //权限改变
                    case "EventNoticeGroupAdmin":
                        GetPermChanged();
                        break;
                    //群名改变
                    case "GroupNameChangedEvent":
                        OnGroupNameChange();
                        break;
                    //群名片改变
                    case "EventNoticeGroupCard":
                        OnMemberCardChanged();
                        break;
                    //头衔改变
                    case "MemberTitleChangedEvent":
                        OnMemberTitleChanged();
                        break;
                    //邀请机器人加群
                    case "NewInvitationRequestedEvent":
                        OnNewInvitationRequested();
                        break;
                    //设备登录
                    case "DeviceLoginEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //好友戳一戳
                    case "FriendPokeEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //群戳一戳
                    case "GroupPokeEvent":
                        await OnGroupPokeEventAsync();
                        break;
                    //GroupNotify
                    case "GroupNotifyEvent":
                        await OnGroupNotifyEventAsnyc();
                        break;
                    //群精华
                    case "GroupEssenceEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //GroupReactionEvent
                    case "GroupReactionEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //其它设备消息
                    case "OtherClientMessageReceiver":
                        ShowMessage($"{EventMessage}");
                        break;
                    //其它设备上线
                    case "OtherClientOnlineEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //其它设备下线
                    case "OtherClientOfflineEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //好友输入状态改变
                    case "FriendInputStatusChangedEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //好友昵称改变
                    case "FriendNickChangedEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //bot退群
                    case "LeftEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //群公告改变
                    case "GroupEntranceAnnouncementChangedEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //匿名聊天
                    case "GroupAllowedAnonymousChatEvent":
                        if (IsSet)
                            Message = "开启匿名聊天";
                        else
                            Message = "关闭匿名聊天";
                        ShowMessage($"{EventMessage}");
                        break;
                    //其它设备消息
                    case "OtherClientMessageEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    //临时消息
                    case "TempMessageEvent":
                        ShowMessage($"{EventMessage}");
                        break;
                    default:
                        ShowMessage($"{EventType}");
                        await HandleMessageAsync();
                        break;
                }
            }
            else if (!Group.IsPowerOn)            
                Reason += "[关机]";                
            
            await GetFriendlyResAsync();
        }



        //判断是否是机器人
        public bool IsAnotherRobot()
        {
            if (Message.IsNull()) return false; 

            //其它官方机器人
            var bots = new[] { 2854208500, 2854197266, 3889001246, 3889019833 };
            foreach (var bot in bots)
            {
                if (Message.Contains($":{bot}") && !IsReply)
                {
                    Reason += "[艾特官机]";
                    return true;
                }
            }

            if (GroupId == BotInfo.MusicGroup && UserId == 2976260341 && IsLightApp)
                return false;
            
            var isRobot = BotInfo.IsRobot(UserId);
            if (isRobot) 
                Reason += "[机器人]";

            return isRobot; 
        }
    }
}
