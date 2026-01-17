using System;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Entities
{
    public class UserGuild
    {
        public const long MIN_USER_ID = 980000000000;
        public const long MAX_USER_ID = 990000000000;
    }
}
