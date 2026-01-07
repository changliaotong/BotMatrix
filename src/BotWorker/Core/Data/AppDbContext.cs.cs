using Microsoft.EntityFrameworkCore;
using sz84.Bots.Models.Achievement;
using sz84.Bots.Models.Limiter;

namespace sz84.Core.Data
{
    public partial class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
    {
        public DbSet<LimiterLog> LimiterLog => Set<LimiterLog>();
        public DbSet<Achievement> Achievements => Set<Achievement>();
        public DbSet<UserAchievement> UserAchievement => Set<UserAchievement>();

        // 可选：配置表名、索引等
        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);
        }    
    }
}
