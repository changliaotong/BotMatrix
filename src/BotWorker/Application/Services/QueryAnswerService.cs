using BotWorker.Modules.AI.Plugins;

namespace BotWorker.Application.Services
{
    public class QueryAnswerService(KnowledgeBaseService knowledgeBaseService)
    {
        private readonly KnowledgeBaseService _knowledgeBaseService = knowledgeBaseService;

        public async Task<(long, float, string)> GetTargetQuestionAsync(string question)
        {            
            //调用知识库，获取问题的目标问�? 例如：你是谁呀 可能会匹�?你是�?           
            return await GetKnowledgeAsync(86433316, question);
        }

        public async Task<(long, float, string)> GetKnowledgeAsync(long groupId, string question)
        {
            var knowledges = await _knowledgeBaseService.GetKnowledgesAsync(groupId, question);
            if (knowledges == null || knowledges.Count == 0)
                return (0, 0, "");

            //每个群定义不同的使用标准
            if (groupId == 86433316)
            {
                //获得评分最高的答案 并且分数要在 0.85 以上 取第一条知�? 
                //要同时返�?id �?相似�?
                var res = knowledges.Where(k => k.Score > 0.85)
                                    .OrderByDescending(k => k.Score)
                                    .Take(1)
                                    .Select(k => (k.Question.AsLong(), k.Score, k.Content))
                                    .FirstOrDefault();

                return res;  // 返回的是 (string question, double Score, string Content)
            }
            else
            {
                //其它�?获取所有知识库答案供参�?
                var res = string.Empty;
                foreach ( var k in knowledges) 
                {
                    res += k.Content;
                }
                return (0, 0, res);
            }
        }     
    }
}


