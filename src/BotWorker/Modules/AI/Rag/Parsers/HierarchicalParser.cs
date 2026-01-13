using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Rag.Parsers
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


