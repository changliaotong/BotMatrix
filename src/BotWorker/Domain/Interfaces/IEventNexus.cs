using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Models;

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
        Task PublishAsync<T>(T eventData) where T : class;

        /// <summary>
        /// 订阅事件
        /// </summary>
        void Subscribe<T>(Func<T, Task> handler) where T : class;

        /// <summary>
        /// 取消订阅
        /// </summary>
        void Unsubscribe<T>(Func<T, Task> handler) where T : class;

        /// <summary>
        /// 获取最近的系统审计日志
        /// </summary>
        List<SystemAuditEvent> GetRecentAudits();

        /// <summary>
        /// 获取当前的全局 Buff 倍率
        /// </summary>
        double GetActiveBuff(BuffType type);
    }
}
