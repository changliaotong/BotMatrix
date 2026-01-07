using Newtonsoft.Json;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Groups;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        [JsonIgnore]
        public override string TableName => "SendMessage";
        [JsonIgnore]
        public override string KeyField => "MsgId";

        public void OnMuteAll()
        {
            //还要考虑官机的权限问题
            IsSend = false;
            if (SelfPerm == 2)
            {
                if (IsSet)
                {
                    Answer = $"全员禁言开启";
                    IsSend = false;
                }
                else
                {
                    Answer = $"全员禁言关闭（已开机）";
                    if (!Group.IsPowerOn)                    
                        GroupInfo.SetPowerOn(GroupId);    
                }
            }
        }

        public void OnPublicSubscribe()
        {
            Answer = $"感谢关注{GroupName}，发送【菜单】了解功能；" +
                     $"点击<a href=\"https://mp.weixin.qq.com/s/SASwgYj7RimKCVbw77hGTg\">查看使用说明</a>";
            
        }

        public void OnPublicUnSubscribe()
        {
            Answer = $"取消关注公众号{GroupName}";
            IsSend = false;
        }

        public void OnBotOnline()
        {
            BotInfo.IsActive[SelfId] = true;
        }

        public async Task OnGroupMemberMute()
        {
            IsCancelProxy = true;

            if (GroupId == 0)
                OnMuteAll();
            else if (UserId == SelfId)
            {
                Answer = "禁言解除，系统开启";
                if (Period > 0)
                {
                    IsSend = false;                    
                    Reason = "[禁言]";
                    Answer = IsVip ? "我被禁言" : "禁言退群";

                    if (!IsVip) 
                        await LeaveAsync(SelfId, GroupId);                    
                }
            }            
            ShowMessage($"{(Period > 0 ? "" : "解除")}禁言：{GroupName}({GroupId}) {Name}({UserId})  {(Period > 0 ? $"时长{Period} " : "")}操作：{OperaterName}({Operater})");
        }

        public void OnGroupNameChange()        
        {
            GroupInfo.SetValue("GroupName", GroupName, GroupId);
            if (IsVip)
                GroupVip.SetValue("GroupName", GroupName, GroupId);            
        }

        public void OnMemberCardChanged()
        {
            IsSend = Group.IsChangeHint && GroupInfo.GetLastHintTime(GroupId) >= 10;
            if (TargetName != "")
            {
                Answer = $"{Card}({UserId}) 改名为：{TargetName}";                
                GroupInfo.SetHintDate(GroupId);
                IsCancelProxy = true;
            }            
        }

        public void OnMemberTitleChanged()
        {
            IsSend = false;
            if (TargetName != "")
            {
                Answer = $"{Name}({UserId}) 获得新头衔：{TargetName}";
                IsCancelProxy = true;
            }
        }

        public async Task OnFriendRecallAsync()
        {
            Answer = await OnRecallAsync();
        }

        public async Task OnGroupPokeEventAsync()
        {
            ShowMessage($"{EventMessage}");
            if (UserId != SelfId && (OfficalBots.Contains(TargetUin) || TargetUin == SelfId))
            {
                //用户戳了机器人
                Message = "GroupPokeEvent";
                await HandleMessageAsync();
            }
        }

        public async Task OnGroupNotifyEventAsnyc()
        {
            IsCancelProxy = true;
            ShowMessage($"{EventMessage}\n {Message}");            
        }

        public async Task OnGroupIncreaseAsync()
        {
            IsCancelProxy = true;

            if (SelfId == UserId)
                GetJoinedRes();
            else
                await GetMemberJoinedAsync();            
        }

        public void OnGroupMemberDecrease()
        {
            IsCancelProxy = true;
            GetBeKicked();            
        }

        public void OnMemberLeft()
        {
            IsCancelProxy = true;

            GetLeaveRes();
            if (!BotInfo.IsRobot(UserId) && GroupInfo.GetLastHintTime(GroupId) >= 10)
                GroupInfo.SetHintDate(GroupId);            
        }

        public void OnEventRequestGroup()
        {
            (var accept, Reason) = GetRequestJoinGroup();
            Accept = accept switch
            {
                1 => 3,
                0 => -3,
                _ => 0,
            };
        }

        public void OnNewInvitationRequested()
        {
            GroupInfo.Append(GroupId, GroupName, SelfId, SelfName, UserId);
            Accept = 2;
        }
    }
}
