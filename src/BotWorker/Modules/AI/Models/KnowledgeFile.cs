using System.ComponentModel.DataAnnotations.Schema;
using System.Data;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Persistence.Database;

namespace BotWorker.Modules.AI.Models
{
    public class KnowledgeFile : MetaDataGuid<KnowledgeFile>
    {
        public override string TableName => "KnowledgeFiles";

        public override string KeyField => "Id";

        public long GroupId { get; set; }
        public string FileName { get; set; } = string.Empty;
        public string DisplayName { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string StoragePath { get; set; } = string.Empty;
        public bool Enabled { get; set; } = true;
        public DateTime UploadTime { get; set; } = DateTime.Now;
        public string FileHash { get; set; } = string.Empty;
        public bool IsEmbedded { get; set; } = false;
        public DateTime? EmbeddedTime { get; set; }
        public string EmbeddingError { get; set; } = string.Empty;
        public long UserId { get; set; } = 0;

        [NotMapped]
        public string BuildStatus
        {
            get
            {
                if (IsEmbedded)
                    return "Success";
                if (!string.IsNullOrWhiteSpace(EmbeddingError))
                    return "Failed";
                if (EmbeddedTime == null)
                    return "Pending"; // 构建中
                return "Unknown";
            }
        }

        // 获取指定群组的文件列表
        public static async Task<List<KnowledgeFile>> GetFilesByGroupAsync(long groupId)
        {
            var sql = $"SELECT * FROM {FullName} WHERE GroupId = {groupId} ORDER BY UploadTime DESC";
            return await QueryListAsync<KnowledgeFile>(sql);
        }

        // 新增文件记录
        public static async Task<Dictionary<string, object>?> AddAsync(KnowledgeFile file)
        {
            return await InsertReturnFieldsAsync(new
            {
                file.GroupId,
                file.FileName,
                file.StoragePath,
                file.Enabled,
                file.UploadTime,
                file.FileHash,
                file.UserId
            }, "Id", "Guid");
        }

        public static void MarkFileEmbedded(string fileId)
        {
            var sql = $"UPDATE {FullName} SET IsEmbedded = 1, EmbeddedTime = {SqlDateTime} WHERE Id = @fileId";
            var parameters = new[]
            {
                DbProviderFactory.CreateParameter("@fileId", fileId),
            };
            Exec(sql, null, parameters);
        }

        public static void MarkEmbeddingFailed(long fileId, string error)
        {
            var sql = $"UPDATE {FullName} SET EmbeddingError = @Error, EmbeddedTime = {SqlDateTime} WHERE Id = @fileId";
            var parameters = new[]
            {
                DbProviderFactory.CreateParameter("@error", error),
                DbProviderFactory.CreateParameter("@fileId", fileId),
            };

            Exec(sql, null, parameters);
        }

        public static async Task<List<KnowledgeFile>> GetPendingEmbeddingFilesAsync(long groupId)
        {
            string sql = $"SELECT * FROM {FullName} WHERE GroupId = @GroupId AND IsEmbedded = 0";
            IDataParameter[] paras = {  DbProviderFactory.CreateParameter("@GroupId", groupId) };
            return await QueryListAsync<KnowledgeFile>(sql, null, paras);
        }

        // 修改启用状态
        public static async Task<int?> SetEnabledAsync(string id, bool enabled)
        {
            var sql = $"UPDATE {FullName} SET Enabled = @Enabled WHERE Id = @Id";
            var parameters = new[]
            {
                DbProviderFactory.CreateParameter("@Enabled", enabled),
                DbProviderFactory.CreateParameter("@Id", id),
            };
            return await ExecAsync(sql, null, parameters);
        }
    }

    public class KnowledgeVectors : MetaData<KnowledgeVectors>
    {
        public override string TableName => "KnowledgeVectors";

        public override string KeyField => "Id";

    }
}
