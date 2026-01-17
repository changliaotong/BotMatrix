using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Domain.Constants;
using BotWorker.Common.Extensions;

namespace BotWorker.Infrastructure.Extensions
{
    public static class OnebotExt
    {
        public static long GetSelfId(this EventBase message)
        {            
            message.Platform ??= Platforms.QQ;
            return message.SelfId;
        }

        public static long GetGroupId(this EventBase message)
        {
            return message.GroupId.AsLong();
        }

        public static long GetUserId(this EventBase message)
        {
            return message.UserId.AsLong();
        }
    }
}


