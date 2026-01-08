using OneBotSharp.Objs.Event;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Constants;
using BotWorker.Common.Extensions;

namespace BotWorker.Infrastructure.Extensions
{
    public static class OnebotExt
    {
        public static long GetSelfId(this EventBase message)
        {            
            message.Platform = message.Platform ?? Platforms.NapCat;
            if (message.Platform == Platforms.QQGuild)
                return UserGuild.GetUserId(0, message.SelfId, message.GroupId);
            else if (message.Platform == Platforms.Weixin)
            {
                //todo 把python logic 迁移过来
                return message.SelfId.AsLong();
            }
            else if (message.Platform == Platforms.Public)
            {
                //todo 公众号逻辑
                return message.SelfId.AsLong();
            }
            else
                return message.SelfId.AsLong();
        }

        public static long GetGroupId(this EventBase message)
        {
            if (message.Platform == Platforms.QQGuild)
                return UserGuild.GetUserId(message.GetSelfId(), message.SelfId, message.GroupId);
            else if (message.Platform == Platforms.Weixin)
            {
                //todo 把python logic 迁移过来
                return message.GroupId.AsLong();
            }
            else if (message.Platform == Platforms.Public)
            {
                //todo 公众号逻辑
                return message.GroupId.AsLong();
            }
            else
                return message.GroupId.AsLong();
        }

        public static long GetUserId(this EventBase message)
        {
            if (message.Platform == Platforms.QQGuild)
                return UserGuild.GetUserId(message.GetSelfId(), message.SelfId, message.GroupId);
            else if (message.Platform == Platforms.Weixin)
            {
                //todo 把python logic 迁移过来
                return message.UserId.AsLong();
            }
            else if (message.Platform == Platforms.Public)
            {
                //todo 公众号逻辑
                return message.UserId.AsLong();
            }
            else
                return message.UserId.AsLong();
        }
    }
}


