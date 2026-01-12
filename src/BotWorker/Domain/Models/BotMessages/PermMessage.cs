namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage> 
    { 
        public void GetPermChanged()
        {
            IsCancelProxy = true;

            if (UserId == SelfId)
            {
                if (GroupVip.IsVip(Group.Id))                
                    Answer = IsSet ? "🎉 我升为管理了，大家快来恭喜我" : "我的管理被取消了";                
                else
                {
                    _ = GroupInfo.SetValue("IsSz84", !IsSet, Group.Id);
                    _ = GroupInfo.SetValue("IsWarn", IsSet, Group.Id);
                    Answer = IsSet ? "🎉 我升为管理了，系统已开启" : "我的管理被取消，系统已关闭";
                }
            }
            else if (UserId != 0)           
                Answer = IsSet ? $"🎉 恭喜：[@:{UserId}] 升为管理了" : $"😱 [@:{UserId}] 管理被取消了";                
            
            IsSend = Group.IsRightHint;            
        }
    }
}
