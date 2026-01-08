using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IPlugin
    {
        string Name { get; }
        string Description { get; }
        Task InitAsync(IRobot robot);
    }
}


