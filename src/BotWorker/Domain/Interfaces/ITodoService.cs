using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface ITodoService
    {
        Task<string> GetTodoResAsync(long groupId, string groupName, long qq, string name, string cmdName, string cmdPara);
    }
}
