using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Services.Rag.Parsers
{
    public class HierarchicalParser : IContentParser
    {
        public string Format => "Hierarchical";

        public async Task<string> ParseAsync(Stream stream)
        {
            using var reader = new StreamReader(stream);
            return await reader.ReadToEndAsync();
        }
    }
}


