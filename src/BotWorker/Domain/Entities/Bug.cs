using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class Bug
    {
        private static IBugRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBugRepository>() 
            ?? throw new InvalidOperationException("IBugRepository not registered");

        public int Id { get; set; }
        public string BugGroup { get; set; } = string.Empty;
        public string BugInfo { get; set; } = string.Empty;
        public DateTime InsertDate { get; set; }

        // 增加一条Bug信息
        public static int Insert(object bugInfo, string? bugGroup = null)
        {
            var bug = new Bug
            {
                BugGroup = bugGroup ?? string.Empty,
                BugInfo = bugInfo?.ToString() ?? string.Empty
            };
            return Repository.AddAsync(bug).GetAwaiter().GetResult();
        }

        public static async Task<int> InsertAsync(object bugInfo, string? bugGroup = null)
        {
            var bug = new Bug
            {
                BugGroup = bugGroup ?? string.Empty,
                BugInfo = bugInfo?.ToString() ?? string.Empty
            };
            return await Repository.AddAsync(bug);
        }
    }
}
