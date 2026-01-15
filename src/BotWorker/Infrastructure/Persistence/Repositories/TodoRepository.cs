using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Tools;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class TodoRepository : BaseRepository<Todo>, ITodoRepository
    {
        public async Task<Todo?> GetByNoAsync(long userId, int todoNo)
        {
            string sql = $"SELECT * FROM \"Todo\" WHERE \"UserId\" = @userId AND \"TodoNo\" = @todoNo";
            return await Connection.QueryFirstOrDefaultAsync<Todo>(sql, new { userId, todoNo });
        }

        public async Task<int> DeleteByNoAsync(long userId, int todoNo)
        {
            string sql = $"DELETE FROM \"Todo\" WHERE \"UserId\" = @userId AND \"TodoNo\" = @todoNo";
            return await Connection.ExecuteAsync(sql, new { userId, todoNo });
        }

        public async Task<int> GetMaxNoAsync(long userId)
        {
            string sql = $"SELECT COALESCE(MAX(\"TodoNo\"), 0) FROM \"Todo\" WHERE \"UserId\" = @userId";
            return await Connection.ExecuteScalarAsync<int>(sql, new { userId });
        }

        public async Task<IEnumerable<Todo>> GetListAsync(long userId)
        {
            string sql = $"SELECT * FROM \"Todo\" WHERE \"UserId\" = @userId ORDER BY \"TodoNo\" DESC";
            return await Connection.QueryAsync<Todo>(sql, new { userId });
        }
    }
}
