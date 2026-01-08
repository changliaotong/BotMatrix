using Microsoft.EntityFrameworkCore;

namespace BotWorker.Application.Services
{
    public class BotDbContext : DbContext
    {
        public BotDbContext(DbContextOptions<BotDbContext> options) : base(options) { }

        public DbSet<LimiterLog> LimiterLogs => Set<LimiterLog>();
        public DbSet<UserInfo> Users => Set<UserInfo>();
        public DbSet<GroupInfo> Groups => Set<GroupInfo>();
        public DbSet<GroupMember> GroupMembers => Set<GroupMember>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            // 配置索引和主键
            modelBuilder.Entity<UserInfo>()
                .HasKey(x => x.Id);

            modelBuilder.Entity<GroupInfo>()
                .HasKey(x => x.Id);
        }
    }
}


