using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
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
            using var conn = CreateConnection();
            string cleanText = text.Replace("，", "").Replace("。", "").Replace("？", "").Replace("！", ""); // Simplified RemoveBiaodian
            return await conn.ExecuteScalarAsync<long>(
                $"SELECT oid FROM {_tableName} WHERE replace(chengyu, '，', '') = @text LIMIT 1",
                new { text = cleanText });
        }

        public async Task<Chengyu?> GetByNameAsync(string name)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Chengyu>(
                $"SELECT * FROM {_tableName} WHERE chengyu = @name",
                new { name });
        }

        public async Task<string> GetCyInfoAsync(string text, long oid = 0)
        {
            if (oid == 0)
                oid = await GetOidAsync(text);
            
            if (oid == 0) return string.Empty;

            using var conn = CreateConnection();
            var cy = await conn.QueryFirstOrDefaultAsync<Chengyu>(
                $"SELECT chengyu as Name, pingyin, diangu, chuchu, lizi FROM {_tableName} WHERE oid = @oid",
                new { oid });

            if (cy == null) return string.Empty;

            string dianguStr = !string.IsNullOrEmpty(cy.Diangu) ? $"\n💡【释义】{cy.Diangu}" : "";
            string chuchuStr = !string.IsNullOrEmpty(cy.Chuchu) ? $"\n📜【出处】{cy.Chuchu}" : "";
            string liziStr = !string.IsNullOrEmpty(cy.Lizi) ? $"\n📝【例子】{cy.Lizi}" : "";

            return $"📚【成语】{cy.Name}\n🔤【拼音】{cy.Pingyin}{dianguStr}{chuchuStr}{liziStr}";
        }

        public async Task<string> GetInfoHtmlAsync(string text, long oid = 0)
        {
            if (oid == 0)
                oid = await GetOidAsync(text);
            
            if (oid == 0) return string.Empty;

            using var conn = CreateConnection();
            var cy = await conn.QueryFirstOrDefaultAsync<Chengyu>(
                $"SELECT chengyu as Name, pingyin, pinyin, spinyin, diangu, chuchu, lizi FROM {_tableName} WHERE oid = @oid",
                new { oid });

            if (cy == null) return string.Empty;

            string pingyinStr = $"{cy.Pingyin} <span>|</span> {cy.Pinyin} <span>|</span> {cy.Spinyin}";
            string dianguStr = !string.IsNullOrEmpty(cy.Diangu) ? $"\n【释义】{cy.Diangu}" : "";
            string chuchuStr = !string.IsNullOrEmpty(cy.Chuchu) ? $"\n【出处】{cy.Chuchu}" : "";
            string liziStr = !string.IsNullOrEmpty(cy.Lizi) ? $"\n【例子】{cy.Lizi}" : "";

            return $"📚【成语】{cy.Name}\n🔤【拼音】{pingyinStr}{dianguStr}{chuchuStr}{liziStr}";
        }

        public async Task<long> CountBySearchAsync(string search)
        {
            using var conn = CreateConnection();
            string pattern = $"%{search}%";
            string cleanSearch = search.Replace(" ", "");
            return await conn.ExecuteScalarAsync<long>(
                $"SELECT count(*) FROM {_tableName} WHERE chengyu LIKE @pattern OR replace(pinyin, ' ', '') LIKE @cleanPattern OR spinyin LIKE @pattern",
                new { pattern, cleanPattern = $"%{cleanSearch}%" });
        }

        public async Task<string> SearchCysAsync(string search, int top = 50)
        {
            using var conn = CreateConnection();
            string pattern = $"%{search}%";
            string cleanSearch = search.Replace(" ", "");
            var names = await conn.QueryAsync<string>(
                $"SELECT chengyu FROM {_tableName} WHERE chengyu LIKE @pattern OR replace(pinyin, ' ', '') LIKE @cleanPattern OR spinyin LIKE @pattern ORDER BY random() LIMIT @top",
                new { pattern, cleanPattern = $"%{cleanSearch}%", top });
            
            if (!names.Any()) return string.Empty;
            return string.Join("", names.Select(n => $"【{n}】")) + $"共{names.Count()}条";
        }

        public async Task<long> GetOidBySearchAsync(string search)
        {
            using var conn = CreateConnection();
            string pattern = $"%{search}%";
            string cleanSearch = search.Replace(" ", "");
            return await conn.ExecuteScalarAsync<long>(
                $"SELECT oid FROM {_tableName} WHERE chengyu LIKE @pattern OR replace(pinyin, ' ', '') LIKE @cleanPattern OR spinyin LIKE @pattern LIMIT 1",
                new { pattern, cleanPattern = $"%{cleanSearch}%" });
        }

        public async Task<long> CountByFanChaAsync(string search)
        {
            using var conn = CreateConnection();
            string pattern = $"%{search}%";
            return await conn.ExecuteScalarAsync<long>(
                $"SELECT count(*) FROM {_tableName} WHERE diangu LIKE @pattern",
                new { pattern });
        }

        public async Task<string> SearchByFanChaAsync(string search, int top = 50)
        {
            using var conn = CreateConnection();
            string pattern = $"%{search}%";
            var names = await conn.QueryAsync<string>(
                $"SELECT chengyu FROM {_tableName} WHERE diangu LIKE @pattern ORDER BY random() LIMIT @top",
                new { pattern, top });
            
            if (!names.Any()) return string.Empty;
            return string.Join("", names.Select(n => $"【{n}】")) + $"共{names.Count()}条";
        }
    }

    public class CidianRepository : BaseRepository<Cidian>, ICidianRepository
    {
        public CidianRepository() : base("cidian", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string> GetDescriptionAsync(string keyword)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(
                $"SELECT description FROM {_tableName} WHERE keyword = @keyword",
                new { keyword }) ?? string.Empty;
        }

        public async Task<IEnumerable<Cidian>> SearchAsync(string keyword, int limit = 20)
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<Cidian>(
                $"SELECT keyword, description FROM {_tableName} WHERE keyword LIKE @keyword ORDER BY keyword LIMIT @limit",
                new { keyword = $"{keyword}%", limit });
        }
    }

    public class CityRepository : BaseRepository<City>, ICityRepository
    {
        public CityRepository() : base("city", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<City?> GetByNameAsync(string cityName)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<City>(
                $"SELECT * FROM {_tableName} WHERE city_name = @cityName",
                new { cityName });
        }
    }
}
