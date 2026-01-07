using Microsoft.EntityFrameworkCore;
using sz84.Bots.Models.Limiter;

namespace sz84.Bots.Plugins
{
    public class AppDbContext(DbContextOptions<AppDbContext> options, IEnumerable<IAppModuleDbContext>? moduleDbContexts = null) : DbContext(options)
    {
        private readonly IEnumerable<IAppModuleDbContext>? _moduleDbContexts = moduleDbContexts;

        public DbSet<LimiterLog> DailyLimitLogs => Set<LimiterLog>();

        protected override void OnModelCreating(ModelBuilder modelBuilder)
        {
            modelBuilder.Entity<LimiterLog>()
                .HasIndex(x => new { x.GroupId, x.UserId, x.ActionKey })
                .IsUnique();

            if (_moduleDbContexts != null)
            {
                foreach (var moduleContext in _moduleDbContexts)
                {
                    moduleContext.RegisterModels(modelBuilder);
                }
            }
        }
    }

}
