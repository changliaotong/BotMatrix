using System.Collections.Concurrent;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.Ranking
{
    public class RankingManager : IRankingService
    {
        // key 结构: keyName -> groupId -> (userId -> value)
        private readonly ConcurrentDictionary<string, ConcurrentDictionary<string, ConcurrentDictionary<long, int>>> _rankings = new();

        public List<RankingEntry> GetTop(string key, string groupId, int topN = 10)
        {
            if (!_rankings.TryGetValue(key, out var groupMap) ||
                !groupMap.TryGetValue(groupId, out var userMap))
                return new List<RankingEntry>();

            return userMap.OrderByDescending(kv => kv.Value)
                          .Take(topN)
                          .Select(kv => new RankingEntry { UserId = kv.Key, Value = kv.Value })
                          .ToList();
        }

        public void UpdateValue(string key, string groupId, long userId, int value)
        {
            var groupMap = _rankings.GetOrAdd(key, _ => new ConcurrentDictionary<string, ConcurrentDictionary<long, int>>());
            var userMap = groupMap.GetOrAdd(groupId, _ => new ConcurrentDictionary<long, int>());
            userMap[userId] = value;
        }

        public void Increment(string key, string groupId, long userId, int delta = 1)
        {
            var groupMap = _rankings.GetOrAdd(key, _ => new ConcurrentDictionary<string, ConcurrentDictionary<long, int>>());
            var userMap = groupMap.GetOrAdd(groupId, _ => new ConcurrentDictionary<long, int>());
            userMap.AddOrUpdate(userId, delta, (_, old) => old + delta);
        }
    }
}
