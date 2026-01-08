using Microsoft.AspNetCore.HttpOverrides;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {        
        //防撤回（已失效）
        public async Task OnGroupRecallAsync()
        {
            InfoMessage($"{(IsSend ? "" : "[未发送]")}{Answer}撤回消息：{MsgId}");
        }

        public async Task<string> OnRecallAsync()
        {
            if (!IsGroup || Group.IsCloudAnswer >= 4)
            {
                if (!IsGuild && UserId != SelfId && Operater != SelfId && Group.IsReplyRecall)
                {
                    var answers = new List<string>
                        {
                            "撤回也没用，窥屏的我都看到了",
                            "怀孕了直说啊 撤回干嘛 大家一起帮你想办法",
                            "我都已经看见了哦！不用装啦！",
                            "怀孕了就要大家一起想办法嘛，撤回有什么用",
                            "说出口的话怎么能撤回呢？",
                            "为什么要撤回呢",
                            "我可是看见了哟（黑化脸）",
                            "撤回我也看见了",
                        };
                    return answers[new Random().Next(answers.Count)];
                }
            }
            return "";
        }

        //撤回消息
        public async Task<string> GetRecallMsgRes()
        {
            if (IsNewAnswer) return "";

            IsCancelProxy = true;

            if (IsReply)
            {
                if (!HaveSetupRight())
                    return $"你没有撤回消息的权限";

                if (!IsGroup)
                    return $"无法撤回私聊消息";

                if (SelfPerm == 2)
                    return $"我不是管理员不能撤回消息哦~";

                if (IsReply)
                    await RecallForwardAsync(SelfId, RealGroupId, MsgId, ReplyMsgId);
                else
                {
                    if (SelfPerm >= UserPerm)
                        return $"不能撤回群主和管理员的消息~";

                    await RecallAsync(SelfId, RealGroupId, MsgId);
                }

                return "";
            }

            return "撤回系统：可设置撤回词，也可以在拉黑、踢出、禁言、警告的同时撤回该消息";
        }
    }
}
