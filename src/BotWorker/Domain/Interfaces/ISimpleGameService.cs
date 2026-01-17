using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface ISimpleGameService
    {
        string RobBuilding(long userId);
        string DaFeiji(long userId);
        string DaDishu(long userId);
        string DaQunzhu(long userId);
        string QiangjiuQunzhu(long userId);
        string AiQunzhu(long userId);
    }
}
