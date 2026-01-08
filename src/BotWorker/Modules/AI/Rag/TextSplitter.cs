using System.Collections.Generic;
using System.Linq;

namespace BotWorker.Services.Rag
{
    public class TextSplitter
    {
        private readonly int _chunkSize;
        private readonly int _overlap;

        public TextSplitter(int chunkSize = 500, int overlap = 50)
        {
            _chunkSize = chunkSize;
            _overlap = overlap;
        }

        public List<string> Split(string text)
        {
            if (string.IsNullOrWhiteSpace(text)) return new List<string>();

            var chunks = new List<string>();
            int start = 0;
            while (start < text.Length)
            {
                int length = Math.Min(_chunkSize, text.Length - start);
                chunks.Add(text.Substring(start, length));
                start += _chunkSize - _overlap;
                if (start >= text.Length) break;
            }
            return chunks;
        }
    }
}


