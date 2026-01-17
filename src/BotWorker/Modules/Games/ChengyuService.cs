using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Games
{
    public class ChengyuService : IChengyuService
    {
        private readonly IChengyuRepository _repository;
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly ILogger<ChengyuService> _logger;

        public ChengyuService(
            IChengyuRepository repository,
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            ILogger<ChengyuService> logger)
        {
            _repository = repository;
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _logger = logger;
        }

        public async Task<long> GetOidAsync(string text)
        {
            return await _repository.GetOidAsync(text);
        }

        public async Task<bool> ExistsAsync(string text)
        {
            return await GetOidAsync(text) != 0;
        }

        public async Task<string> PinYinAsync(string text)
        {
            var cy = await _repository.GetByNameAsync(text);
            return cy?.Pingyin ?? string.Empty;
        }

        public async Task<string> PinYinAsciiAsync(string text)
        {
            var cy = await _repository.GetByNameAsync(text);
            return cy?.Pinyin ?? string.Empty;
        }

        public async Task<string> GetCyInfoAsync(string text, long oid = 0)
        {
            return await _repository.GetCyInfoAsync(text, oid);
        }

        public async Task<Dictionary<string, string>> GetCyInfoAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetCyInfoAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        public async Task<string> GetInfoHtmlAsync(string text, long oid = 0)
        {
            return await _repository.GetInfoHtmlAsync(text, oid);
        }

        public async Task<Dictionary<string, string>> GetInfoHtmlAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetInfoHtmlAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        public async Task<string> PinYinFirstAsync(string textCy)
        {
            var pinyin = await PinYinAsciiAsync(textCy);
            if (string.IsNullOrEmpty(pinyin)) return string.Empty;
            int idx = pinyin.IndexOf(' ');
            return idx > 0 ? pinyin[..idx] : pinyin;
        }

        public async Task<string> PinYinLastAsync(string text)
        {
            var pinyin = await PinYinAsciiAsync(text);
            if (string.IsNullOrEmpty(pinyin)) return string.Empty;
            int idx = pinyin.LastIndexOf(' ');
            return idx > 0 ? pinyin.Substring(idx + 1) : pinyin;
        }

        public async Task<string> GetCyResAsync(IPluginContext ctx, string cmdPara)
        {
            if (string.IsNullOrEmpty(cmdPara))
                return "📚 格式：成语 + 关键字\n📌 例如：成语 德高望重";

            var count = await _repository.CountBySearchAsync(cmdPara);
            if (count == 0)
                return "没有找到相关成语";

            string res = count == 1
                ? await _repository.GetCyInfoAsync("", await _repository.GetOidBySearchAsync(cmdPara))
                : "📚" + await _repository.SearchCysAsync(cmdPara, 50);

            var creditRes = await MinusCreditResAsync(ctx, 10, "成语扣分");
            return res + creditRes;
        }

        public async Task<string> GetFanChaResAsync(IPluginContext ctx, string cmdPara)
        {
            if (string.IsNullOrWhiteSpace(cmdPara))
                return "📚 格式：反查 + 关键字\n例如：反查 坚强 ";

            var count = await _repository.CountByFanChaAsync(cmdPara);
            if (count == 0)
                return "没有找到相关成语";

            string res = count == 1
                ? await _repository.GetCyInfoAsync("", await _repository.GetOidBySearchAsync(cmdPara))
                : await _repository.SearchByFanChaAsync(cmdPara, 50);

            var creditRes = await MinusCreditResAsync(ctx, 10, "成语扣分");
            return res + creditRes;
        }

        public async Task<string> GetRandomAsync(string category)
        {
            return await _repository.GetRandomAsync(category);
        }

        private async Task<string> MinusCreditResAsync(IPluginContext ctx, long creditMinus, string creditInfo)
        {
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var group = await _groupRepo.GetAsync(groupId);
            if (group == null || !group.IsCreditSystem) return "";

            var userId = long.Parse(ctx.UserId);
            var botId = long.Parse(ctx.BotId);

            var res = await _userRepo.AddCreditAsync(botId, groupId, group.GroupName, userId, ctx.UserName, -creditMinus, creditInfo);
            return res.Success ? $"\n💎 {{积分类型}}：-{creditMinus}，累计：{res.CreditValue:N0}" : "";
        }
    }
}
