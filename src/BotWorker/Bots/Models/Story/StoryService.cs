using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Core.Interfaces;

namespace BotWorker.Bots.Models.Story
{
    public class StoryService : IBotModule
    {
        private readonly Dictionary<string, int> _chapterProgress = new();
        public IModuleMetadata Metadata => new StoryMetadata();

        public string AdvanceStory(string userId)
        {
            int chapter = _chapterProgress.TryGetValue(userId, out var c) ? c + 1 : 1;
            _chapterProgress[userId] = chapter;

            return $"📖 你进入了第 {chapter} 章：{GenerateChapterTitle(chapter)}";
        }

        private static string GenerateChapterTitle(int chapter)
        {
            return $"《虚拟世界 第{chapter}章》";
        }        

        public void ConfigureDbContext(ModelBuilder modelBuilder)
        {
            throw new NotImplementedException();
        }

        public void RegisterServices(IServiceCollection services, IConfiguration config)
        {
            services.AddSingleton<StoryService>();
            Console.WriteLine("✅ [StoryModule] 注册成功");
        }
    }

}
