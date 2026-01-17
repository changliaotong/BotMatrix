using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Interfaces
{
    public interface IBrickService
    {
        Task<string> GetBrickResAsync(BotMessage botMsg);
    }
}
