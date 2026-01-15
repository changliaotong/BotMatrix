using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BlackListRepository : BaseRepository<BlackList>, IBlackListRepository
    {
        public BlackListRepository(string? connectionString = null)
            : base("black_list", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<IEnumerable<long>> GetSystemBlackListAsync()
        {
            const string sql = "SELECT black_id FROM black_list WHERE group_id = @groupId";
            using var conn = CreateConnection();
            return await conn.QueryAsync<long>(sql, new { groupId = BotInfo.GroupIdDef });
        }

        public async Task<bool> IsExistsAsync(long groupId, long userId)
        {
            const string sql = "SELECT COUNT(1) FROM black_list WHERE group_id = @groupId AND black_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, userId }) > 0;
        }

        public async Task<int> AddAsync(BlackList blackList)
        {
            const string sql = @"
                INSERT INTO black_list (bot_uin, group_id, group_name, user_id, user_name, black_id, black_info)
                VALUES (@BotUin, @GroupId, @GroupName, @UserId, @UserName, @BlackId, @BlackInfo)
                ON CONFLICT (group_id, black_id) DO NOTHING";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, blackList);
        }

        public async Task<int> DeleteAsync(long groupId, long userId)
        {
            const string sql = "DELETE FROM black_list WHERE group_id = @groupId AND black_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }

        public async Task<int> ClearGroupAsync(long groupId)
        {
            const string sql = "DELETE FROM black_list WHERE group_id = @groupId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId });
        }
    }

    public class WhiteListRepository : BaseRepository<WhiteList>, IWhiteListRepository
    {
        public WhiteListRepository(string? connectionString = null)
            : base("white_list", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<bool> IsExistsAsync(long groupId, long userId)
        {
            const string sql = "SELECT COUNT(1) FROM white_list WHERE group_id = @groupId AND white_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, userId }) > 0;
        }

        public async Task<int> AddAsync(WhiteList whiteList)
        {
            const string sql = @"
                INSERT INTO white_list (bot_uin, group_id, group_name, user_id, user_name, white_id)
                VALUES (@BotUin, @GroupId, @GroupName, @UserId, @UserName, @WhiteId)
                ON CONFLICT (group_id, white_id) DO NOTHING";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, whiteList);
        }

        public async Task<int> DeleteAsync(long groupId, long userId)
        {
            const string sql = "DELETE FROM white_list WHERE group_id = @groupId AND white_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }
    }

    public class GreyListRepository : BaseRepository<GreyList>, IGreyListRepository
    {
        public GreyListRepository(string? connectionString = null)
            : base("grey_list", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<IEnumerable<long>> GetSystemGreyListAsync()
        {
            const string sql = "SELECT grey_id FROM grey_list WHERE group_id = @groupId";
            using var conn = CreateConnection();
            return await conn.QueryAsync<long>(sql, new { groupId = BotInfo.GroupIdDef });
        }

        public async Task<bool> IsExistsAsync(long groupId, long userId)
        {
            const string sql = "SELECT COUNT(1) FROM grey_list WHERE group_id = @groupId AND grey_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, userId }) > 0;
        }

        public async Task<int> AddAsync(GreyList greyList)
        {
            const string sql = @"
                INSERT INTO grey_list (bot_uin, group_id, group_name, user_id, user_name, grey_id, grey_info)
                VALUES (@BotUin, @GroupId, @GroupName, @UserId, @UserName, @GreyId, @GreyInfo)
                ON CONFLICT (group_id, grey_id) DO NOTHING";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, greyList);
        }

        public async Task<int> DeleteAsync(long groupId, long userId)
        {
            const string sql = "DELETE FROM grey_list WHERE group_id = @groupId AND grey_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { groupId, userId });
        }
    }

    public class BugRepository : BaseRepository<Bug>, IBugRepository
    {
        public BugRepository(string? connectionString = null)
            : base("bug", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddAsync(Bug bug)
        {
            const string sql = @"
                INSERT INTO bug (bug_group, bug_info)
                VALUES (@BugGroup, @BugInfo)";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, bug);
        }
    }

    public class BotHintsRepository : BaseRepository<BotHints>, IBotHintsRepository
    {
        public BotHintsRepository(string? connectionString = null)
            : base("bot_hints", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<string> GetHintAsync(string cmd)
        {
            const string sql = "SELECT hint FROM bot_hints WHERE cmd = @cmd";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { cmd }) ?? string.Empty;
        }
    }

    public class TokenRepository : BaseRepository<Token>, ITokenRepository
    {
        public TokenRepository(string? connectionString = null)
            : base("token", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<Token?> GetByUserIdAsync(long userId)
        {
            const string sql = "SELECT * FROM token WHERE user_id = @userId";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Token>(sql, new { userId });
        }

        public async Task<string> GetTokenByUserIdAsync(long userId)
        {
            const string sql = "SELECT token FROM token WHERE user_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { userId }) ?? string.Empty;
        }

        public async Task<bool> ExistsTokenAsync(string token)
        {
            const string sql = "SELECT COUNT(1) FROM token WHERE token = @token";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { token }) > 0;
        }

        public async Task<bool> ExistsTokenAsync(long userId, string token)
        {
            const string sql = "SELECT COUNT(1) FROM token WHERE user_id = @userId AND token = @token";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId, token }) > 0;
        }

        public async Task<int> UpsertTokenAsync(long userId, string token)
        {
            const string sql = @"
                INSERT INTO token (user_id, token, token_date)
                VALUES (@userId, @token, CURRENT_TIMESTAMP)
                ON CONFLICT (user_id) DO UPDATE SET token = @token, token_date = CURRENT_TIMESTAMP";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { userId, token });
        }

        public async Task<int> UpsertRefreshTokenAsync(long userId, string token, string refreshToken, DateTime expiryDate)
        {
            const string sql = @"
                INSERT INTO token (user_id, token, refresh_token, expiry_date)
                VALUES (@userId, @token, @refreshToken, @expiryDate)
                ON CONFLICT (user_id) DO UPDATE SET refresh_token = @refreshToken, expiry_date = @expiryDate";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { userId, token, refreshToken, expiryDate });
        }

        public async Task<string> GetRefreshTokenAsync(long userId)
        {
            const string sql = "SELECT refresh_token FROM token WHERE user_id = @userId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { userId }) ?? string.Empty;
        }

        public async Task<bool> IsTokenValidAsync(long userId, string token, long seconds)
        {
            const string sql = @"
                SELECT COUNT(1) 
                FROM token 
                WHERE user_id = @userId 
                AND token = @token 
                AND ABS(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - token_date))) < @seconds";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId, token, seconds }) > 0;
        }
    }

    public class GroupOfficalRepository : BaseRepository<GroupOffical>, IGroupOfficalRepository
    {
        public GroupOfficalRepository(string? connectionString = null)
            : base("group_offical", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<bool> IsOfficalAsync(long groupId)
        {
            const string sql = "SELECT COUNT(1) FROM group_offical WHERE group_id = @groupId";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId }) > 0;
        }
    }

    public class GroupEventRepository : BaseRepository<GroupEvent>, IGroupEventRepository
    {
        public GroupEventRepository(string? connectionString = null)
            : base("group_event", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddAsync(GroupEvent groupEvent)
        {
            const string sql = @"
                INSERT INTO group_event (bot_uin, group_id, event_type, event_msg)
                VALUES (@BotUin, @GroupId, @EventType, @EventMsg)";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, groupEvent);
        }
    }

    public class FriendRepository : BaseRepository<Friend>, IFriendRepository
    {
        public FriendRepository(string? connectionString = null)
            : base("friend", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<int> AddAsync(Friend friend)
        {
            const string sql = @"
                INSERT INTO friend (bot_uin, friend_id, friend_name)
                VALUES (@BotUin, @FriendId, @FriendName)
                ON CONFLICT (bot_uin, friend_id) DO NOTHING";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, friend);
        }
    }
}
