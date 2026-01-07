using BotWorker.Bots.Entries;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {

        private readonly string[] funnyLeaveMessages =
            [
                "退群了，挥一挥手，不带走一片云彩",
                "退群了，可能是去拯救地球了",
                "退群了，留下我们在风中凌乱",
                "退群了，此生不见，也别想念",
            ];

        // 有人退群
        public void GetLeaveRes()
        {
            if (UserId == 0)
                return;

            SubInviteCount();
            
            IsSend = Group.IsExitHint;
            Answer = $"🚀 {Name.ReplaceInvalid()}({UserId}) {funnyLeaveMessages.RandomOne()}\n";

            if (Group.IsBlackExit && !BotInfo.IsRobot(UserId))
            {
                Answer += AddBlack(UserId, "退群拉黑") == -1
                    ? $"退群拉黑{RetryMsg}"
                    : $" {(Answer.IsNull() ? $"({UserId}) 退群" : "")} 已拉黑！";
            }
            
            //本机发送退群消息及欢迎语
            if (!IsGuild && IsProxy)
                IsCancelProxy = true;
        }
    }
}
