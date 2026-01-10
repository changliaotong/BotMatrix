using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IContextManager
    {
        Task<Dictionary<string, object>> GetContextAsync(string key);
        Task SetContextAsync(string key, Dictionary<string, object> context);
        Task RemoveContextAsync(string key);
    }

    public class ContextManager : IContextManager
    {
        private readonly ConcurrentDictionary<string, Dictionary<string, object>> _contexts = new();

        public Task<Dictionary<string, object>> GetContextAsync(string key)
        {
            return Task.FromResult(_contexts.GetOrAdd(key, _ => new Dictionary<string, object>()));
        }

        public Task SetContextAsync(string key, Dictionary<string, object> context)
        {
            _contexts[key] = context;
            return Task.CompletedTask;
        }

        public Task RemoveContextAsync(string key)
        {
            _contexts.TryRemove(key, out _);
            return Task.CompletedTask;
        }
    }
}


