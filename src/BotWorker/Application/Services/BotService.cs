using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Services
{
    public class BotService : IBotService
    {
        public bool IsSuperAdmin(long userId)
        {
            return userId == BotInfo.AdminUin || userId == BotInfo.AdminUin2;
        }
    }
}
