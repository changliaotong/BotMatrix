using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IMusicVideoRepository : IBaseRepository<MusicVideo>
    {
        Task<string> GetContentByVidAsync(string vid);
        Task<bool> ExistsByVidAsync(string vid);
        Task<int> AddAsync(MusicVideo musicVideo);
    }
}
