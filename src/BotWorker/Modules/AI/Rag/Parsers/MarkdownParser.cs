using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Rag.Parsers
{
    public class MarkdownParser : IContentParser
    {
        public string Format => ".md";

        public async Task<string> ParseAsync(Stream stream)
        {
            using var reader = new StreamReader(stream);
            return await reader.ReadToEndAsync();
        }
    }
}


