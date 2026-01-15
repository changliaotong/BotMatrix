using System;
using System.Collections.Generic;
using System.ComponentModel.DataAnnotations.Schema;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Modules.AI.Models
{
    public class KnowledgeFile
    {
        public long Id { get; set; }
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

        private static IKnowledgeFileRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IKnowledgeFileRepository>() 
            ?? throw new InvalidOperationException("IKnowledgeFileRepository not registered");

        // 获取指定群组的文件列表
        public static async Task<IEnumerable<KnowledgeFile>> GetFilesByGroupAsync(long groupId)
        {
            return await Repository.GetFilesByGroupAsync(groupId);
        }

        // 新增文件记录
        public static async Task<long> AddAsync(KnowledgeFile file)
        {
            return await Repository.AddAsync(file);
        }

        public static void MarkFileEmbedded(long fileId)
        {
            Repository.MarkFileEmbeddedAsync(fileId).GetAwaiter().GetResult();
        }

        public static void MarkEmbeddingFailed(long fileId, string error)
        {
            Repository.MarkEmbeddingFailedAsync(fileId, error).GetAwaiter().GetResult();
        }

        public static async Task<IEnumerable<KnowledgeFile>> GetPendingEmbeddingFilesAsync(long groupId)
        {
            return await Repository.GetPendingEmbeddingFilesAsync(groupId);
        }

        // 修改启用状态
        public static async Task<bool> SetEnabledAsync(long fileId, bool enabled)
        {
            return await Repository.SetEnabledAsync(fileId, enabled);
        }
    }
}
