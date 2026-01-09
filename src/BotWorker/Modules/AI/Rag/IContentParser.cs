using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Rag
{
    public interface IContentParser
    {
        string Format { get; }
        Task<string> ParseAsync(Stream stream);
    }
}


