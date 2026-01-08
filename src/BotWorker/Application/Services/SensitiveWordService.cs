using System.Collections.Generic;
using System.Linq;

namespace BotWorker.Services
{
    public interface ISensitiveWordService
    {
        string Filter(string text);
        bool ContainsSensitiveWord(string text);
    }

    public class SensitiveWordService : ISensitiveWordService
    {
        private readonly HashSet<string> _sensitiveWords = new();

        public string Filter(string text)
        {
            if (string.IsNullOrEmpty(text)) return text;
            
            var result = text;
            foreach (var word in _sensitiveWords)
            {
                result = result.Replace(word, new string('*', word.Length));
            }
            return result;
        }

        public bool ContainsSensitiveWord(string text)
        {
            if (string.IsNullOrEmpty(text)) return false;
            return _sensitiveWords.Any(text.Contains);
        }
    }
}


