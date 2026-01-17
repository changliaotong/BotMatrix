using BotWorker.Domain.Models.BotMessages;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IAgentService
    {
        Task<bool> TryParseAgentCallAsync(BotMessage botMsg);
        Task GetAgentResAsync(BotMessage botMsg);
        Task<string> ChangeAgentAsync(BotMessage botMsg);
        Task GetImageResAsync(BotMessage botMsg);
        Task<bool> IsEnoughAsync(BotMessage botMsg);
    }
}
