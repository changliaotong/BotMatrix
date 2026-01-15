using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Entities
{
    public partial class Chengyu
    {
        private static IChengyuRepository? _repository;
        private static IChengyuRepository Repository => _repository ??= BotMessage.ServiceProvider?.GetRequiredService<IChengyuRepository>() ?? throw new InvalidOperationException("IChengyuRepository not registered");

        public static async Task<long> GetOidAsync(string text)
        {
            return await Repository.GetOidAsync(text);
        }

        public static async Task<bool> ExistsAsync(string text)
        {
            return await GetOidAsync(text) != 0;
        }

        public static async Task<string> PinYinAsync(string text)
        {
            var cy = await Repository.GetByNameAsync(text);
            return cy?.Pingyin ?? string.Empty;
        }

        public static async Task<string> PinYinAsciiAsync(string text)
        {
            var cy = await Repository.GetByNameAsync(text);
            return cy?.Pinyin ?? string.Empty;
        }

        public static async Task<string> GetCyInfoAsync(string text, long oid = 0)
        {
            return await Repository.GetCyInfoAsync(text, oid);
        }

        //一次获得多个成语的解释网页版
        public static async Task<Dictionary<string, string>> GetCyInfoAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetCyInfoAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //成语解释网页版 拼音部分更详细
        public static async Task<string> GetInfoHtmlAsync(string text, long oid = 0)
        {
            return await Repository.GetInfoHtmlAsync(text, oid);
        }

        //一次获得多个成语的解释网页版
        public static async Task<Dictionary<string, string>> GetInfoHtmlAsync(IEnumerable<string> cys)
        {
            Dictionary<string, string> res = [];
            foreach (var cy in cys)
            {
                string cyInfo = await GetInfoHtmlAsync(cy);
                res.TryAdd(cy, cyInfo);
            }
            return res;
        }

        //首字拼音
        public static async Task<string> PinYinFirstAsync(string textCy)
        {
            var pinyin = await PinYinAsciiAsync(textCy);
            return pinyin[..pinyin.IndexOf(' ')];
        }

        //尾字拼音
        public static async Task<string> PinYinLastAsync(string text)
        {
            var pinyin = await PinYinAsciiAsync(text);
            return pinyin.Substring(pinyin.LastIndexOf(' ') + 1, pinyin.Length - pinyin.LastIndexOf(" ") - 1);
        }


        //成语解释
        public static async Task<string> GetCyResAsync(BotMessage bm)
        {
            if (bm.CmdPara.Contains("接龙"))
            {
                if (BotCmd.IsClosedCmd(bm.GroupId, "接龙"))
                    return "接龙功能已关闭";
                else
                {
                    bm.Answer = bm.Answer.Replace("接龙", "");
                    return await bm.GetJielongRes();
                }
            }

            if (bm.CmdPara.IsNull())
                return "📚 格式：成语 + 关键字\n📌 例如：成语 德高望重";
            
            var count = await Repository.CountBySearchAsync(bm.CmdPara);
            if (count == 0)
                return "没有找到相关成语";
            
            string res = count == 1
                ? await Repository.GetCyInfoAsync("", await Repository.GetOidBySearchAsync(bm.CmdPara))
                : "📚" + await Repository.SearchCysAsync(bm.CmdPara, 50);
            
            return res + await bm.MinusCreditResAsync(10, "成语扣分");
        }

        // 反查 根据释义反查成语
        public static async Task<string> GetFanChaResAsync(BotMessage bm)
        {
            if (bm.CmdPara.IsNullOrWhiteSpace())
                return "📚 格式：反查 + 关键字\n例如：反查 坚强 ";
            
            var count = await Repository.CountByFanChaAsync(bm.CmdPara);
            if (count == 0)
                return "没有找到相关成语";
            
            string res = count == 1
                ? await Repository.GetCyInfoAsync("", await Repository.GetOidBySearchAsync(bm.CmdPara))
                : await Repository.SearchByFanChaAsync(bm.CmdPara, 50);
            
            res += await bm.MinusCreditResAsync(10, "成语扣分");
            return res;
        }
    }
}
