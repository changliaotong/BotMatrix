using Microsoft.EntityFrameworkCore;
using sz84.Bots.Models.Achievement;
using sz84.Bots.Models.Limiter;
using sz84.Bots.Models.Title;

namespace BotWorker.Domain.Entities
{
    public partial class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
    {
        public DbSet<LimiterLog> LimiterLog => Set<LimiterLog>();
        public DbSet<Achievement> Achievements => Set<Achievement>();
        public DbSet<UserAchievement> UserAchievement => Set<UserAchievement>();
        public DbSet<UserTitle> Titles => Set<UserTitle>();
        public DbSet<UserTitle> UserTitles => Set<UserTitle>();

        // 可选：配置表名、索引等
        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);
        }    
    }
}

