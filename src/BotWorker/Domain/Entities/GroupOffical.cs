using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class GroupOffical
    {
        public const long MIN_GROUP_ID = 990000000000;
        public long GroupId { get; set; }
    }
}
