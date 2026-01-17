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
    }
}
