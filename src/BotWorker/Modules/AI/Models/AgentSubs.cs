using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Modules.AI.Models
{
    public class AgentSubs
    {
        public long Id { get; set; }
        public long UserId { get; set; }
        public long AgentId { get; set; }
        public bool IsSub { get; set; }
        public DateTime CreatedAt { get; set; }
        public DateTime UpdatedAt { get; set; }
    }
}
