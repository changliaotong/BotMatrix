using Microsoft.EntityFrameworkCore;
using BotWorker.Modules.Achievement;
using BotWorker.Models;

namespace BotWorker.Core.Data
{
    public partial class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
    {
        public DbSet<LimiterLog> LimiterLog => Set<LimiterLog>();
        public DbSet<Achievement> Achievements => Set<Achievement>();
        public DbSet<UserAchievement> UserAchievement => Set<UserAchievement>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            base.OnModelCreating(modelBuilder);
        }    
    }
}


