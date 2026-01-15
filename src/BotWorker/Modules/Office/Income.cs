using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Infrastructure.Utils.Schema.Attributes;
using Dapper.Contrib.Extensions;

namespace BotWorker.Modules.Office
{
    [Table("income")]
    public class Income
    {
        private static IIncomeRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IIncomeRepository>() 
            ?? throw new InvalidOperationException("IIncomeRepository not registered");

        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long GoodsCount { get; set; }
        public string GoodsName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public decimal IncomeMoney { get; set; }
        public string PayMethod { get; set; } = string.Empty;
        public string IncomeTrade { get; set; } = string.Empty;
        public string IncomeInfo { get; set; } = string.Empty;
        public int InsertBy { get; set; }
        public DateTime IncomeDate { get; set; } = DateTime.Now;

        public static async Task<float> TotalAsync(long userId) => await Repository.GetTotalAsync(userId);
        public static float Total(long userId) => TotalAsync(userId).GetAwaiter().GetResult();
        public static async Task<float> TotalLastYearAsync(long userId) => await Repository.GetTotalLastYearAsync(userId);

        public static async Task<bool> IsVipOnceAsync(long groupId) => await Repository.IsVipOnceAsync(groupId);

        public static async Task<int> GetClientLevelAsync(long userId) => await Repository.GetClientLevelAsync(userId);

        public static async Task<string> GetLevelListAsync(long groupId) => await Repository.GetLevelListAsync(groupId);

        public static async Task<string> GetLeverOrderAsync(long groupId, long userId) => await Repository.GetLeverOrderAsync(groupId, userId);

        public static async Task<string> TodayAsync() => await Repository.GetStatAsync("today");

        public static async Task<string> YesterdayAsync() => await Repository.GetStatAsync("yesterday");

        public static async Task<string> ThisMonthAsync() => await Repository.GetStatAsync("month");

        public static async Task<string> ThisYearAsync() => await Repository.GetStatAsync("year");

        public static async Task<string> AllAsync() => await Repository.GetStatAsync("all");
    }
}
