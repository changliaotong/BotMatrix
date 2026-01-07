using System;
using System.Threading.Tasks;
using BotWorker.Bots.BotMessages;

namespace BotWorker.Core.Pipeline
{
    /// <summary>
    /// 机器人消息处理中间件接口
    /// </summary>
    public interface IMiddleware
    {
        /// <summary>
        /// 执行中间件逻辑
        /// </summary>
        /// <param name="context">BotMessage 上下文</param>
        /// <param name="next">指向下一个中间件的委托</param>
        /// <returns>返回 true 表示管道继续，false 表示中断</returns>
        Task<bool> InvokeAsync(BotMessage context, Func<Task<bool>> next);
    }
}
