using System;
using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    /// <summary>
    /// BotMatrix 核心事件中枢接口
    /// 支持插件间的解耦通信与联动
    /// </summary>
    public interface IEventNexus
    {
        /// <summary>
        /// 发布事件
        /// </summary>
        /// <typeparam name="T">事件类型</typeparam>
        /// <param name="eventData">事件数据</param>
        Task PublishAsync<T>(T eventData) where T : class;

        /// <summary>
        /// 订阅事件
        /// </summary>
        /// <typeparam name="T">事件类型</typeparam>
        /// <param name="handler">处理函数</param>
        void Subscribe<T>(Func<T, Task> handler) where T : class;

        /// <summary>
        /// 取消订阅
        /// </summary>
        /// <typeparam name="T">事件类型</typeparam>
        /// <param name="handler">处理函数</param>
        void Unsubscribe<T>(Func<T, Task> handler) where T : class;
    }
}
