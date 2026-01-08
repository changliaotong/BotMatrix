using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Agents.Entries
{
    public class AgentSubs : MetaData<AgentSubs>
    {
        public override string TableName => "AgentSubs";
        public override string KeyField => "UserId";
        public override string KeyField2 => "AgentId";

        public long UserId { get; set; }
        public long AgentId { get; set; }

        public static async Task<int> AppendAsync(long userId, long id, bool isSub = true)
        {
            return await InsertObjectAsync(new
            {
                userId,
                AgentId = id,
                IsSub = isSub.AsInt(),
                SubDate = DateTime.MinValue
            });
        }

        public static bool IsSub(long userId, Guid guid)
        {
            return GetBool("IsSub", userId, Agent.GetId(guid));
        } 

        public static async Task<int> SubAsync(long userId, long id, bool isSub = true)
        {
            int i = 0;

            if (!Exists(userId, id)) 
                i = await AppendAsync(userId, id, isSub);
            else
                i = await UpdateAsync(new { IsSub = isSub.AsInt(), UnsubDate = DateTime.MinValue }, userId, id);

            if (i == 0) return i;
            var (sql, parameters) = Agent.GetSqlPlusCount(id, isSub ? 1 : -1);
            return await ExecAsync(sql, parameters);
        }
    }
}
