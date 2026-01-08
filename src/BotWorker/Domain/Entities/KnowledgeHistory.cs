
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class KnowledgeHistory : MetaData<KnowledgeHistory>
    {
        public override string TableName => "KnowledgeHistory";

        public override string KeyField => "Id";

        public static int AddKnowledgeHistroy(string question, string targetQuestion, long targetQuestionId, float Similarity, long answerId, string answer)
        {
            var (sql, paras) = SqlInsert([
                                new Cov("Question", question),
                                new Cov("TargetQuestion", targetQuestion),
                                new Cov("TargetQuestionId", targetQuestionId),
                                new Cov("Similarity", Similarity),
                                new Cov("AnswerId", answerId),
                                new Cov("Answer", answer),
                            ]);
            return Exec(sql, paras);
        }
    }
}
