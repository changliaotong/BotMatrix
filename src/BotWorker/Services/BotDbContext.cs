using Microsoft.EntityFrameworkCore;
using sz84.Bots.Users;
using sz84.Bots.Entries;
using sz84.Groups;
using sz84.Bots.Games.Gift;
using BotWorker.Models;

namespace BotWorker.Services
{
    public class BotDbContext : DbContext
    {
        public BotDbContext(DbContextOptions<BotDbContext> options) : base(options) { }

        public DbSet<LimiterLog> LimiterLogs => Set<LimiterLog>();
        public DbSet<sz84.Bots.Users.UserInfo> Users => Set<sz84.Bots.Users.UserInfo>();
        public DbSet<sz84.Bots.Entries.GroupInfo> Groups => Set<sz84.Bots.Entries.GroupInfo>();
        public DbSet<sz84.Groups.GroupMember> GroupMembers => Set<sz84.Groups.GroupMember>();
        public DbSet<Gift> Gifts => Set<Gift>();
        public DbSet<GiftLog> GiftLogs => Set<GiftLog>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            // 配置索引和主键
            modelBuilder.Entity<sz84.Bots.Users.UserInfo>()
                .HasKey(x => x.Id);

            modelBuilder.Entity<sz84.Bots.Entries.GroupInfo>()
                .HasKey(x => x.Id);
        }
    }
}
