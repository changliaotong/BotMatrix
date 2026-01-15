using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class Friend
    {
        private static IFriendRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IFriendRepository>() 
            ?? throw new InvalidOperationException("IFriendRepository not registered");

        public long BotUin { get; set; }
        public long FriendId { get; set; }
        public string FriendName { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        public static async Task<int> InsertFriendAsync(long botUin, long friendId, string friendName)
        {
            var friend = new Friend
            {
                BotUin = botUin,
                FriendId = friendId,
                FriendName = friendName
            };
            return await Repository.AddAsync(friend);
        }
    }
}
