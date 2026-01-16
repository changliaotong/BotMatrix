using System.Data;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using Dapper;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotMessageRepository : BaseRepository<BotMessage>, IBotMessageRepository
    {
        public BotMessageRepository() : base("SendMessage")
        {
        }

        protected override string KeyField => "MsgId";

        public async Task<string> InsertAsync(BotMessage message, IDbTransaction? trans = null)
        {
            if (trans != null)
            {
                await trans.Connection.InsertAsync(message, trans);
                return message.MsgId;
            }
            using var conn = CreateConnection();
            await conn.InsertAsync(message);
            return message.MsgId;
        }

        public async Task<bool> UpdateAsync(BotMessage message, IDbTransaction? trans = null)
        {
            if (trans != null)
            {
                return await trans.Connection.UpdateAsync(message, trans);
            }
            using var conn = CreateConnection();
            return await conn.UpdateAsync(message);
        }

        public async Task<bool> DeleteAsync(string msgId, IDbTransaction? trans = null)
        {
            string sql = $"DELETE FROM {_tableName} WHERE {KeyField} = @msgId";
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { msgId }, trans) > 0;
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { msgId }) > 0;
        }

        public async Task<BotMessage?> GetByMsgIdAsync(string msgId, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE {KeyField} = @msgId";
            if (trans != null)
            {
                return await trans.Connection.QueryFirstOrDefaultAsync<BotMessage>(sql, new { msgId }, trans);
            }
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<BotMessage>(sql, new { msgId });
        }
    }
}
