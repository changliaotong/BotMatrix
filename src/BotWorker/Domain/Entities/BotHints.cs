using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Dapper.Contrib.Extensions.Table("BotHints")]
    public class BotHints
    {
        private static IBotHintsRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBotHintsRepository>() 
            ?? throw new InvalidOperationException("IBotHintsRepository not registered");

        [Dapper.Contrib.Extensions.Key]
        public int Id { get; set; }
        public string Cmd { get; set; } = string.Empty;
        public string Hint { get; set; } = string.Empty;

        public static async Task<string> GetHintAsync(string cmd)
        {
            return await Repository.GetHintAsync(cmd);
        }

        public static string GetHint(string cmd)
        {
            return GetHintAsync(cmd).GetAwaiter().GetResult();
        }
    }
}
