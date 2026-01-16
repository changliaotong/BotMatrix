using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using System.Data;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class HandleQuestionRepository : IHandleQuestionRepository
    {
        private readonly IDbConnection _connection;

        public HandleQuestionRepository(IDbConnection connection)
        {
            _connection = connection;
        }

        public async Task<long> InsertAsync(HandleQuestion question, IDbTransaction? trans = null)
        {
            return await _connection.InsertAsync(question, trans);
        }
    }
}
