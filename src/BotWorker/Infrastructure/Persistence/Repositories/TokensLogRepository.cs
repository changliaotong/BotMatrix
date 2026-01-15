using System.Threading.Tasks;
using System.Data;
using BotWorker.Common;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class TokensLogRepository : BaseRepository<TokensLog>, ITokensLogRepository
    {
        public TokensLogRepository() : base("token_log", GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null)
        {
            var log = new TokensLog
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = userName,
                TokensAdd = tokensAdd,
                TokensValue = tokensValue,
                TokensInfo = tokensInfo,
                InsertDate = System.DateTime.Now
            };

            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(
                    $"INSERT INTO {TableName} (BotUin, GroupId, GroupName, UserId, UserName, TokensAdd, TokensValue, TokensInfo, InsertDate) VALUES (@BotUin, @GroupId, @GroupName, @UserId, @UserName, @TokensAdd, @TokensValue, @TokensInfo, @InsertDate)",
                    log, trans);
            }
            else
            {
                return await AddAsync(log);
            }
        }
    }
}
