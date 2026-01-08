using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IFissionService
    {
        Task<string> CreateFissionLinkAsync(long userId);
        Task ProcessFissionAsync(long invitedUserId, string code);
    }

    public class FissionService : IFissionService
    {
        public async Task<string> CreateFissionLinkAsync(long userId)
        {
            return await Task.FromResult($"FISSION-{userId}");
        }

        public async Task ProcessFissionAsync(long invitedUserId, string code)
        {
            await Task.CompletedTask;
        }
    }
}


