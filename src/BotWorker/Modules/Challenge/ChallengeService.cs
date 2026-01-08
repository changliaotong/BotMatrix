using Microsoft.Extensions.DependencyInjection;
using BotWorker.Application.Services;

namespace BotWorker.Modules.Challenge
{
    public class ChallengeService
    {
        private readonly LimiterService _dailyLimit;
        private readonly Dictionary<(long, long), (int Level, DateTime? LastFail)> _progress = [];

        public ChallengeService(IServiceProvider provider)
        {
            _dailyLimit = provider.GetRequiredService<LimiterService>();
        }

        public async Task<string> ChallengeAsync(long groupId, long userId)
        {
            var now = DateTime.Now;
            var data = _progress.GetValueOrDefault((groupId, userId));

            if (await _dailyLimit.HasUsedAsync(groupId, userId, "challenge_max") && data.LastFail?.Date == now.Date)
                return "📛 今日挑战次数已用尽";

            if (data.LastFail.HasValue && (now - data.LastFail.Value).TotalMinutes < 10)
                return "⌛ 挑战失败后冷却中，请稍后再试";

            bool success = new Random().NextDouble() > 0.5;

            if (success)
            {
                int reward = 5 + data.Level;
                //_credit?.AddCredit(reward, "成功通关");
                _progress[(groupId, userId)] = (data.Level + 1, null);
                return $"✅ 成功通关第 {data.Level + 1} 关，奖励 {reward} 积分！";
            }
            else
            {
                _progress[(groupId, userId)] = (data.Level, now);
                await _dailyLimit.MarkUsedAsync(groupId, userId, "challenge_max");
                //_credit?.MinusCredit(3, "挑战失败");
                return "❌ 挑战失败，进入10分钟冷却，扣除3积分";
            }
        }
    }
}
