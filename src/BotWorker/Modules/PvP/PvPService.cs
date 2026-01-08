using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Interfaces;

namespace BotWorker.Modules.PvP
{
    public class PvPService : IBotModule
    {
        private readonly Dictionary<long, int> _score = new();
        private readonly Dictionary<long, int> _win = new();
        private readonly Dictionary<long, int> _lose = new();

        public bool? IsEnable { get => throw new NotImplementedException(); set => throw new NotImplementedException(); }

        public IModuleMetadata Metadata => new PvPMetadata();

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            throw new NotImplementedException();
        }

        public string Duel(long userId, long opponentId)
        {
            _score.TryAdd(userId, 1000);
            _score.TryAdd(opponentId, 1000);

            bool win = new Random().NextDouble() > 0.5;
            int delta = new Random().Next(5, 15);

            if (win)
            {
                _score[userId] += delta;
                _score[opponentId] -= delta;
                _win[userId] = _win.GetValueOrDefault(userId) + 1;
                _lose[opponentId] = _lose.GetValueOrDefault(opponentId) + 1;
                ///_credit.AddCredit(userId, delta, "战胜");
                //_credit.MinusCredit(opponentId, delta, "战败");
                return $"🏆 {userId} 战胜了 {opponentId}，获得 {delta} 积分！";
            }
            else
            {
                _score[opponentId] += delta;
                _score[userId] -= delta;
                _win[opponentId] = _win.GetValueOrDefault(opponentId) + 1;
                _lose[userId] = _lose.GetValueOrDefault(userId) + 1;
                //_credit.AddCredit(opponentId, delta, "战败");
                //_credit.MinusCredit(userId, delta, "战胜");
                return $"💀 {opponentId} 战胜了 {userId}，获得 {delta} 积分！";
            }
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<PvPService>();
            InfoMessage($"✅ [{nameof(PvPService)}] 注册成功\n" +
                $"插件名称：{Metadata.Name}\n" +
                $"插件版本：{Metadata.Version}\n" +
                $"插件描述：{Metadata.Description}\n" +
                $"插件作者：{Metadata.Author}");
        }
    }

}
