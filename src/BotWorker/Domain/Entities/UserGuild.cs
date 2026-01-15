using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class UserGuild
    {
        private static IUserRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserRepository>() 
            ?? throw new InvalidOperationException("IUserRepository not registered");

        public const long MIN_USER_ID = 980000000000;
        public const long MAX_USER_ID = 990000000000;

        public static async Task<long> GetUserIdAsync(long botUin, string userOpenid, string groupOpenid)
        {
            if (string.IsNullOrEmpty(userOpenid)) 
                return 0;

            var userId = await GetTargetUserIdAsync(userOpenid);
            if (userId != 0)
            {
                var bot = await Repository.GetBotUinByOpenidAsync(userOpenid);
                if (bot != botUin)
                    await Repository.SetValueAsync("bot_uin", botUin, userId);
                return userId;
            }

            userId = await GetMaxUserIdAsync();
            int i = await UserInfo.AppendAsync(botUin, 0, userId, "", 0, userOpenid, groupOpenid);
            return i == -1 ? 0 : userId;
        }

        public static async Task<long> GetTargetUserIdAsync(string userOpenid)
        {
            return await Repository.GetTargetUserIdAsync(userOpenid);
        }

        private static async Task<long> GetMaxUserIdAsync()
        {
            var userId = await Repository.GetMaxIdInRangeAsync(MIN_USER_ID, MAX_USER_ID);
            return userId <= MIN_USER_ID ? MIN_USER_ID + 1 : userId + 1;
        }

        public static async Task<string> GetUserOpenidAsync(long selfId, long user)
        {
            return await Repository.GetUserOpenidAsync(selfId, user);
        }
    }
}
