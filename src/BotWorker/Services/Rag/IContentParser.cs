using System.IO;
using System.Threading.Tasks;

namespace BotWorker.Services.Rag
{
    public interface IContentParser
    {
        string Format { get; }
        Task<string> ParseAsync(Stream stream);
    }
}
