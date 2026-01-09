using System.ComponentModel;
using System.Diagnostics.CodeAnalysis;
using Microsoft.SemanticKernel;
using BotWorker.Modules.AI.Interfaces;

namespace BotWorker.Modules.AI.Plugins
{
    public class KnowledgeBasePlugin(IKnowledgeBaseService knowledgeBaseService, long groupId): KernelPlugin("knowledge", "当用户的问题与本群所配置的知识库内容有关时，调用此函数（如学校政策、公司制度等）")
    {
        private readonly IKnowledgeBaseService _knowledgeBaseService = knowledgeBaseService;
        private readonly long _groupId = groupId;
        private readonly List<KernelFunction> _functions = [];

        public override int FunctionCount => _functions.Count;

        public override IEnumerator<KernelFunction> GetEnumerator() => _functions.GetEnumerator();

        public override bool TryGetFunction(string name, [NotNullWhen(true)] out KernelFunction? function)
        {
            function = _functions.FirstOrDefault(f => f.Name == name);
            return function != null;
        }

        [KernelFunction(name: "get_knowledge")]
        [Description("当用户的问题与本群所配置的知识库内容有关时，调用此函数（如学校政策、公司制度等）")]
        public async Task<string> GetKnowledgeAsync(string question)
        {
            if (string.IsNullOrWhiteSpace(question))
                return "未提供问题，无法检索知识库。";

            return await _knowledgeBaseService.BuildPrompt(question, _groupId);
        }
    }

}
