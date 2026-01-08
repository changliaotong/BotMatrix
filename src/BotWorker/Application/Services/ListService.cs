using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IListService
    {
        Task<IEnumerable<string>> GetListAsync(string listName);
        Task AddToListAsync(string listName, string item);
    }

    public class ListService : IListService
    {
        private readonly Dictionary<string, List<string>> _lists = new();

        public Task<IEnumerable<string>> GetListAsync(string listName)
        {
            if (_lists.TryGetValue(listName, out var list))
            {
                return Task.FromResult<IEnumerable<string>>(list);
            }
            return Task.FromResult<IEnumerable<string>>(new List<string>());
        }

        public Task AddToListAsync(string listName, string item)
        {
            if (!_lists.ContainsKey(listName))
            {
                _lists[listName] = new List<string>();
            }
            _lists[listName].Add(item);
            return Task.CompletedTask;
        }
    }
}


