using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    [Table("greeting_records")]
    public class GreetingRecords
    {
        private static IGreetingRecordsRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGreetingRecordsRepository>() 
            ?? throw new InvalidOperationException("IGreetingRecordsRepository not registered");

        [Key]
        public long Id { get; set; }
        public long BotQQ { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long QQ { get; set; }
        public string Name { get; set; } = string.Empty;
        public int GreetingType { get; set; }
        public DateTime LogicalDate { get; set; }

        public static int Append(long botQQ, long groupId, string groupName, long qq, string name, int greetingType = 0) => 
            AppendAsync(botQQ, groupId, groupName, qq, name, greetingType).GetAwaiter().GetResult();

        public static async Task<int> AppendAsync(long botQQ, long groupId, string groupName, long qq, string name, int greetingType = 0)
        {
            return await Repository.AppendAsync(botQQ, groupId, groupName, qq, name, greetingType);
        }

        public static bool Exists(long groupId, long qq, int greetingType = 0) => 
            ExistsAsync(groupId, qq, greetingType).GetAwaiter().GetResult();

        public static async Task<bool> ExistsAsync(long groupId, long qq, int greetingType = 0)
        {
            return await Repository.ExistsAsync(groupId, qq, greetingType);
        }

        //全服第x位起床用户
        public static int GetCount(int greetingType = 0) => 
            GetCountAsync(greetingType).GetAwaiter().GetResult();

        public static async Task<int> GetCountAsync(int greetingType = 0)
        {
            return await Repository.GetCountAsync(greetingType);
        }

        //本群第x位起床用户
        public static int GetCount(long groupId, int greetingType = 0) => 
            GetCountAsync(groupId, greetingType).GetAwaiter().GetResult();

        public static async Task<int> GetCountAsync(long groupId, int greetingType = 0)
        {
            return await Repository.GetCountAsync(groupId, greetingType);
        }
    }
}
