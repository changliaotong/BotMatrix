using System;
using System.Collections.Generic;

namespace BotWorker.Modules.AI.Rag
{
    public class Chunk
    {
        public string Id { get; set; } = Guid.NewGuid().ToString();
        public string Content { get; set; } = string.Empty;
        public float[]? Embedding { get; set; }
        public Dictionary<string, object> Metadata { get; set; } = new();
        public string Source { get; set; } = string.Empty;
    }
}


