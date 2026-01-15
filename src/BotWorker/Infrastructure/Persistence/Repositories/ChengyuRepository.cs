using System.Collections.Generic;
using System.Threading.Tasks;
using System.Text;
using System.Linq;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class ChengyuRepository : BaseRepository<Chengyu>, IChengyuRepository
    {
        public ChengyuRepository() : base("chengyu", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<long> GetOidAsync(string text)
        {
            // Simplified query, assuming text is already cleaned or exact match
            var oid = await GetValueAsync<long?>("Oid", "WHERE Name = @text", new { text });
            return oid ?? 0;
        }

        public async Task<Chengyu?> GetByNameAsync(string name)
        {
            return await GetFirstOrDefaultAsync("WHERE Name = @name", new { name });
        }

        public async Task<string> GetCyInfoAsync(string text, long oid = 0)
        {
            Chengyu? cy;
            if (oid > 0)
                cy = await GetByIdAsync(oid);
            else
                cy = await GetByNameAsync(text);
            
            if (cy == null) return string.Empty;
            
            var sb = new StringBuilder();
            sb.Append($"📚【成语】{cy.Name}\n🔤【拼音】{cy.Pingyin}");
            if (!string.IsNullOrEmpty(cy.Diangu)) sb.Append($"\n💡【释义】{cy.Diangu}");
            if (!string.IsNullOrEmpty(cy.Chuchu)) sb.Append($"\n📜【出处】{cy.Chuchu}");
            if (!string.IsNullOrEmpty(cy.Lizi)) sb.Append($"\n📝【例子】{cy.Lizi}");
            sb.Append(")");
            return sb.ToString();
        }

        public async Task<string> GetInfoHtmlAsync(string text, long oid = 0)
        {
            Chengyu? cy;
            if (oid > 0)
                cy = await GetByIdAsync(oid);
            else
                cy = await GetByNameAsync(text);
            
            if (cy == null) return string.Empty;

            var sb = new StringBuilder();
            sb.Append($"📚【成语】{cy.Name}\n🔤【拼音】{cy.Pingyin} <span>|</span> {cy.Pinyin} <span>|</span> {cy.Spinyin}");
            if (!string.IsNullOrEmpty(cy.Diangu)) sb.Append($"\n【释义】{cy.Diangu}");
            if (!string.IsNullOrEmpty(cy.Chuchu)) sb.Append($"\n【出处】{cy.Chuchu}");
            if (!string.IsNullOrEmpty(cy.Lizi)) sb.Append($"\n【例子】{cy.Lizi}");
            sb.Append(")");
            return sb.ToString();
        }

        public async Task<long> CountBySearchAsync(string search)
        {
            return await GetCountAsync("WHERE Name LIKE @search", new { search = $"%{search}%" });
        }

        public async Task<string> SearchCysAsync(string search, int top = 50)
        {
            var list = await GetListAsync("WHERE Name LIKE @search LIMIT @top", new { search = $"%{search}%", top });
            return string.Join(" ", list.Select(x => x.Name));
        }

        public async Task<long> GetOidBySearchAsync(string search)
        {
            var oid = await GetValueAsync<long?>("Oid", "WHERE Name LIKE @search LIMIT 1", new { search = $"%{search}%" });
            return oid ?? 0;
        }

        public async Task<long> CountByFanChaAsync(string search)
        {
            return await GetCountAsync("WHERE Diangu LIKE @search OR Lizi LIKE @search", new { search = $"%{search}%" });
        }

        public async Task<string> SearchByFanChaAsync(string search, int top = 50)
        {
            var list = await GetListAsync("WHERE Diangu LIKE @search OR Lizi LIKE @search LIMIT @top", new { search = $"%{search}%", top });
            return string.Join(" ", list.Select(x => x.Name));
        }
    }
}
