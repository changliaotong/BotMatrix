using System;
using System.Threading.Tasks;
using Microsoft.EntityFrameworkCore;

namespace BotWorker.Services
{
    public interface IDatabaseService
    {
        Task<int> SaveChangesAsync();
        BotDbContext Context { get; }
    }

    public class DatabaseService : IDatabaseService
    {
        private readonly BotDbContext _context;

        public DatabaseService(BotDbContext context)
        {
            _context = context;
        }

        public BotDbContext Context => _context;

        public async Task<int> SaveChangesAsync()
        {
            try
            {
                return await _context.SaveChangesAsync();
            }
            catch (Exception ex)
            {
                // 记录日志
                Console.WriteLine($"Database save error: {ex.Message}");
                throw;
            }
        }
    }
}


