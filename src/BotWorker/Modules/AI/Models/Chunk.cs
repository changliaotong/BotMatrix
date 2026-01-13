using System;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Models
{
    public class KnowledgeBase
    {
        public long Id { get; set; }
        public string Name { get; set; } = string.Empty;
        public string? Description { get; set; }
        public long UserId { get; set; }
        public bool IsPublic { get; set; }
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }

    public class Chunk
    {
        public long Id { get; set; }
        public long KbId { get; set; } // 对应 knowledge_chunks.kb_id
        public string Content { get; set; } = string.Empty;
        public float[]? Embedding { get; set; } // 对应 knowledge_chunks.embedding
        public string Metadata { get; set; } = "{}"; // JSONB
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

        // 辅助属性：用于计算分数的相似度
        public double? Score { get; set; }
    }
}
