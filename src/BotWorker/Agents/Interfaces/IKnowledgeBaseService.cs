namespace BotWorker.Agents.Interfaces
{
    // 你需要自己实现的知识库服务接口
    public interface IKnowledgeBaseService
    {
        Task<string> BuildPrompt(string query, long groupId);
    }
}
