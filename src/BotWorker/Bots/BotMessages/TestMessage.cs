using BotWorker.Bots.Extensions;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string GetTestIt()        
        {
            //this.AddKick(UserId);
            //this.AddGroupMessage(CurrentGroupId, UserId, "洗洗睡吧", true);
            //this.AddMute(UserId, 600);
            //this.AddRecall(MsgId);
            //this.AddSetTitle(UserId, "资深客服");
            //this.AddLeave(GroupId);  
            return "";
        }
    }
}
