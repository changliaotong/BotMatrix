using System.Collections.Concurrent;
using BotWorker.Domain.Models;

namespace BotWorker.Application.Services
{
    /// <summary>
    /// BotMatrix 事件中枢实现
    /// 提供高性能的进程内事件发布与订阅机制
    /// </summary>
    public class EventNexus : IEventNexus
    {
        private readonly ConcurrentDictionary<Type, List<Delegate>> _subscriptions = new();
        private readonly ConcurrentQueue<SystemAuditEvent> _auditLog = new();
        private readonly ConcurrentDictionary<BuffType, GlobalBuffEvent> _activeBuffs = new();
        private const int MaxAuditLogSize = 50;

        public EventNexus()
        {
            // 自动订阅审计事件，记录到内存队列
            Subscribe<SystemAuditEvent>(ev => {
                _auditLog.Enqueue(ev);
                while (_auditLog.Count > MaxAuditLogSize)
                {
                    _auditLog.TryDequeue(out _);
                }
                return Task.CompletedTask;
            });

            // 自动订阅全局 Buff 事件
            Subscribe<GlobalBuffEvent>(ev => {
                _activeBuffs[ev.Type] = ev;
                return Task.CompletedTask;
            });
        }

        public double GetActiveBuff(BuffType type)
        {
            if (_activeBuffs.TryGetValue(type, out var buff) && buff.IsActive)
            {
                return buff.Multiplier;
            }
            return 1.0; // 默认倍率为 1.0
        }

        public List<SystemAuditEvent> GetRecentAudits()
        {
            return _auditLog.Reverse().ToList();
        }

        public async Task PublishAsync<T>(T eventData) where T : class
        {
            if (eventData == null) return;

            var eventType = typeof(T);
            if (_subscriptions.TryGetValue(eventType, out var handlers))
            {
                List<Delegate> handlersCopy;
                lock (handlers)
                {
                    handlersCopy = handlers.ToList();
                }

                var tasks = handlersCopy
                    .Select(handler => ((Func<T, Task>)handler)(eventData))
                    .ToList();

                await Task.WhenAll(tasks);
            }
        }

        public void Subscribe<T>(Func<T, Task> handler) where T : class
        {
            var eventType = typeof(T);
            var handlers = _subscriptions.GetOrAdd(eventType, _ => new List<Delegate>());
            
            lock (handlers)
            {
                if (!handlers.Contains(handler))
                {
                    handlers.Add(handler);
                }
            }
        }

        public void Unsubscribe<T>(Func<T, Task> handler) where T : class
        {
            var eventType = typeof(T);
            if (_subscriptions.TryGetValue(eventType, out var handlers))
            {
                lock (handlers)
                {
                    handlers.Remove(handler);
                }
            }
        }
    }
}
