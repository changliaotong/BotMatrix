using System.IO;
using System.Threading.Tasks;
using System.Collections.Generic;
using System.Linq;

namespace BotWorker.Services.Rag.Parsers
{
    public class CodeParser : IContentParser
    {
        public string Format => "Code";
        private readonly string[] _extensions = { ".cs", ".go", ".py", ".js", ".ts", ".java", ".c", ".cpp" };

        public async Task<string> ParseAsync(Stream stream)
        {
            using var reader = new StreamReader(stream);
            var text = await reader.ReadToEndAsync();
            return $"```\n{text}\n```";
        }
    }
}
