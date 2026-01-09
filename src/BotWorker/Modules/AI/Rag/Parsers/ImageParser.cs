using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Rag.Parsers
{
    public class ImageParser : IContentParser
    {
        public string Format => "Image";

        public async Task<string> ParseAsync(Stream stream)
        {
            return await Task.FromResult("[Image Data]");
        }
    }
}


