using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games.Gift
{
    [Table("Gift")]
    public class Gift
    {
        private static IGiftRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGiftRepository>() 
            ?? throw new InvalidOperationException("IGiftRepository not registered");

        [Key]
        public long Id { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public long GiftCredit { get; set; }
        public bool IsValid { get; set; } = true;

        public static async Task<long> GetGiftIdAsync(string giftName)
        {
            return await Repository.GetGiftIdAsync(giftName);
        }

        public static async Task<long> GetRandomGiftAsync(long botUin, long groupId, long qq)
        {
            return await Repository.GetRandomGiftAsync(botUin, groupId, qq);
        }

        public static async Task<string> GetGiftListAsync(long botUin, long groupId, long qq)
        {
            return await Repository.GetGiftListAsync(botUin, groupId, qq);
        }

        public static async Task<Gift?> GetAsync(long id)
        {
            return await Repository.GetByIdAsync(id);
        }
    }
}
