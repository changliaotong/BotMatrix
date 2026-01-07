using Microsoft.EntityFrameworkCore;
using BotWorker.Bots.Users;
using BotWorker.Bots.Entries;
using BotWorker.Groups;
using BotWorker.Bots.Games.Gift;
using BotWorker.Models;

namespace BotWorker.Services
{
    public class BotDbContext : DbContext
    {
        public BotDbContext(DbContextOptions<BotDbContext> options) : base(options) { }

        public DbSet<LimiterLog> LimiterLogs => Set<LimiterLog>();
        public DbSet<BotWorker.Bots.Users.UserInfo> Users => Set<BotWorker.Bots.Users.UserInfo>();
        public DbSet<BotWorker.Bots.Entries.GroupInfo> Groups => Set<BotWorker.Bots.Entries.GroupInfo>();
        public DbSet<BotWorker.Groups.GroupMember> GroupMembers => Set<BotWorker.Groups.GroupMember>();
        public DbSet<Gift> Gifts => Set<Gift>();
        public DbSet<GiftLog> GiftLogs => Set<GiftLog>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            // 配置索引和主键
            modelBuilder.Entity<BotWorker.Bots.Users.UserInfo>()
                .HasKey(x => x.Id);

            modelBuilder.Entity<BotWorker.Bots.Entries.GroupInfo>()
                .HasKey(x => x.Id);
        }
    }
}
