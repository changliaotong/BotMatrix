using System.Threading.Tasks;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Application.Messaging.Handlers;

namespace BotWorker.Application.Services
{
    public interface IHotCmdService
    {
        /// <summary>
        /// 判定是否为热门指令
        /// </summary>
        bool IsHot(BotMessage botMsg);

        /// <summary>
        /// 处理热门指令并返回结果
        /// </summary>
        Task<CommandResult> HandleHotCmdAsync(BotMessage botMsg);
    }
}


