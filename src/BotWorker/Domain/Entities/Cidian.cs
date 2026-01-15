using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Entities
{
    public partial class Cidian
    {
        private static ICidianRepository? _repository;
        private static ICidianRepository Repository => _repository ??= BotMessage.ServiceProvider?.GetRequiredService<ICidianRepository>() ?? throw new InvalidOperationException("ICidianRepository not registered");

        // 翻译功能 先从数据库读取单词翻译，不存在的再调用有道翻译
        public static string GetCiDianRes(string text)
        {
            // Note: This is synchronous in the original code, but our repository is async.
            // For now, we use .Result to match the original signature, but in a real refactor we should make this async.
            string res = Repository.GetDescriptionAsync(text).Result;

            res = res.ReplaceInvalid();

            return res;
        }

        public static string GetCiba(string? text)
        {
            if (text == null) return "";
            var results = Repository.SearchAsync(text, 20).Result;
            if (!results.Any())
            {
                if (text.Trim().Contains(' '))
                    return GetCiba(text.Trim().Split(new char[] { '\u002C', ' ', '，', '、', '\n' }, StringSplitOptions.RemoveEmptyEntries).Last());
                return "";
            }

            return string.Join("<br />", results.Select(r => $"{r.Keyword} {r.Description}"));
        }
    }
}
