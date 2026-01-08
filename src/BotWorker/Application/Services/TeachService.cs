using System.Threading.Tasks;

namespace BotWorker.Services
{
    public interface ITeachService
    {
        Task AddTeachAsync(string question, string answer, long creatorId);
        Task<string?> GetAnswerAsync(string question);
    }

    public class TeachService : ITeachService
    {
        public async Task AddTeachAsync(string question, string answer, long creatorId)
        {
            await Task.CompletedTask;
        }

        public async Task<string?> GetAnswerAsync(string question)
        {
            return await Task.FromResult<string?>(null);
        }
    }
}


