using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class Friend
    {
        public long BotUin { get; set; }
        public long FriendId { get; set; }
        public string FriendName { get; set; } = string.Empty;
        public long Credit { get; set; }
        public long SaveCredit { get; set; }
        public DateTime InsertDate { get; set; }
    }
}
