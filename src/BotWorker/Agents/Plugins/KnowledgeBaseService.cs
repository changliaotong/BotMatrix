using System.Net.Http.Json;
using System.Text;
using BotWorker.Agents.Interfaces;
using BotWorker.Domain.Entities;

namespace BotWorker.Agents.Plugins
{
    public class KnowledgeBaseService(HttpClient httpClient, string kbApiUrl = "/ask") : IKnowledgeBaseService
    {
        private readonly HttpClient _httpClient = httpClient;
        private readonly string _kbApiUrl = kbApiUrl;

        public class KnowledgeResult
        {
            public string Content { get; set; } = string.Empty;
            public string Source { get; set; } = string.Empty;
            public string Question { get; set; } = string.Empty;
            public float Score { get; set; }
        }

        public class KnowledgeResponse
        {
            public List<KnowledgeResult> Results { get; set; } = [];
        }

        public async Task<List<KnowledgeResult>?> GetKnowledgesAsync(long groupId, string question)
        {
            try
            {
                var response = await _httpClient.PostAsJsonAsync(_kbApiUrl, new { question, group_id = groupId.ToString() });
                if (!response.IsSuccessStatusCode)
                {
                    var responseBody = await response.Content.ReadAsStringAsync();
                    Console.WriteLine($"请求 {_kbApiUrl} 失败，状态码：{response.StatusCode}，错误信息：{responseBody}");
                    return null;
                }

                var knowledgeResponse = await response.Content.ReadFromJsonAsync<KnowledgeResponse>();
                if (knowledgeResponse == null)
                    return null;

                return knowledgeResponse.Results;
            }
            catch (Exception ex)
            { 
                InfoMessage($"获取知识库失败{ex.Message}");
                return null;
            };
        }

        public static string BuildPrompt(long groupId, string question)
        {
            var systemPrompt = GroupInfo.GetValue("ai_system", groupId);

            var promptBuilder = new StringBuilder(systemPrompt);
            promptBuilder.AppendLine();

            promptBuilder.AppendLine("请你根据用户的多轮提问以及资料内容，优先结合资料来回答当前问题；如果资料无相关内容，请合理补充但不得编造。回答应自然清晰。");

            return promptBuilder.ToString();
        }

        public async Task<string> BuildPrompt(string query, long groupId)
        {
            var result = await GetKnowledgesAsync(groupId, query);
            if (result == null)
                return string.Empty;

            var promptBuilder = new StringBuilder();
            promptBuilder.AppendLine();
           
            var systemResults = result.Where(r => r.Source == "group").Select(r => r.Content); 
            if (systemResults.Any())
                promptBuilder.AppendLine($"【本群知识库】:\n{string.Join("\n\n", systemResults)}");

            systemResults = result.Where(r => r.Source != "group").Select(r => r.Content);

            if (systemResults.Any())
                promptBuilder.AppendLine($"【系统知识库】:\n{string.Join("\n\n", systemResults)}");

            return promptBuilder.ToString();
        }
    }

}
