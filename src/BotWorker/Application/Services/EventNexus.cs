using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Application.Services
{
    /// <summary>
    /// BotMatrix 事件中枢实现
    /// 提供高性能的进程内事件发布与订阅机制
    /// </summary>
    public class EventNexus : IEventNexus
    {
        private readonly ConcurrentDictionary<Type, List<Delegate>> _subscriptions = new();

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
