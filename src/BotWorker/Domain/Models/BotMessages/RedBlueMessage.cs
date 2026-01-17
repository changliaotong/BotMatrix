using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Games;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Domain.Models.BotMessages
{
    //红和蓝
    public partial class BotMessage
    {
        public async Task<string> GetRedBlueResAsync(bool isDetail = true)
        {
            var redBlueService = ServiceProvider!.GetRequiredService<IRedBlueService>();
            return await redBlueService.GetRedBlueResAsync(this, isDetail);
        }
    }
}
