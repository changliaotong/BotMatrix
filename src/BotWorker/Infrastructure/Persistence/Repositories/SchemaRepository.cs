using System.Threading.Tasks;
using BotWorker.Domain.Constants.AppsInfo;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Database;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class SchemaRepository : ISchemaRepository
    {
        public async Task<string> GetFieldCaptionAsync(string tableId, string fieldName)
        {
            using var conn = DbProviderFactory.CreateConnection();
            const string sql = "SELECT field_caption FROM field_info WHERE table_id = @tableId AND field_name = @fieldName";
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { tableId, fieldName }) ?? fieldName;
        }

        public async Task<string> GetReportFieldCaptionAsync(string rptId, string fieldName)
        {
            using var conn = DbProviderFactory.CreateConnection();
            const string sql = "SELECT field_caption FROM report_field_info WHERE rpt_id = @rptId AND field_name = @fieldName";
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { rptId, fieldName }) ?? fieldName;
        }

        public async Task<TableInfo?> GetTableInfoAsync(int tableId)
        {
            using var conn = DbProviderFactory.CreateConnection();
            const string sql = "SELECT * FROM table_info WHERE table_id = @tableId";
            return await conn.QueryFirstOrDefaultAsync<TableInfo>(sql, new { tableId });
        }
    }
}
