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
    [Table("gift")]
    public class Gift
    {
        [Key]
        public long Id { get; set; }
        public string GiftName { get; set; } = string.Empty;
        public long GiftCredit { get; set; }
        public bool IsValid { get; set; } = true;
    }
}
