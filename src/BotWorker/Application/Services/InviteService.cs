using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IInviteService
    {
        Task<string> GenerateInviteCodeAsync(long userId);
        Task<bool> UseInviteCodeAsync(long userId, string code);
    }

    public class InviteService : IInviteService
    {
        public async Task<string> GenerateInviteCodeAsync(long userId)
        {
            return await Task.FromResult($"INVITE-{userId}-{Guid.NewGuid().ToString().Substring(0, 8)}");
        }

        public async Task<bool> UseInviteCodeAsync(long userId, string code)
        {
            // 简化实�?
            return await Task.FromResult(true);
        }
    }
}


