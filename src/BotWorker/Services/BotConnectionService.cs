using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Services
{
    public interface IBotConnectionService
    {
        Task AddConnectionAsync(BotConnectionConfig config);
        Task RemoveConnectionAsync(string name);
        IEnumerable<BotConnectionConfig> GetConnections();
    }

    public class BotConnectionConfig
    {
        public string Name { get; set; } = string.Empty;
        public string Type { get; set; } = string.Empty; // OneBot, WeChat, etc.
        public string Endpoint { get; set; } = string.Empty;
        public string? Token { get; set; }
        public bool Enabled { get; set; } = true;
    }

    public class BotConnectionService : IBotConnectionService
    {
        private readonly ConcurrentDictionary<string, BotConnectionConfig> _connections = new();

        public Task AddConnectionAsync(BotConnectionConfig config)
        {
            if (string.IsNullOrEmpty(config.Name)) throw new ArgumentException("Connection name cannot be empty");
            _connections[config.Name] = config;
            return Task.CompletedTask;
        }

        public Task RemoveConnectionAsync(string name)
        {
            _connections.TryRemove(name, out _);
            return Task.CompletedTask;
        }

        public IEnumerable<BotConnectionConfig> GetConnections()
        {
            return _connections.Values.ToList();
        }
    }
}

