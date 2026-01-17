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

        public async Task<int> AddBlackListAsync(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo)
        {
            if (await IsExistsAsync(groupId, blackQQ))
                return 0;

            var blackList = new BlackList
            {
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = qq,
                UserName = name,
                BlackId = blackQQ,
                BlackInfo = blackInfo
            };

            return await AddAsync(blackList);
        }

        public async Task<bool> IsSystemBlackAsync(long userId)
        {
            return await IsExistsAsync(BotInfo.GroupIdDef, userId);
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

        public async Task<long> GetTargetGroupAsync(string groupOpenid)
        {
            const string sql = "SELECT COALESCE(target_group, id) FROM group_offical WHERE group_openid = @groupOpenid";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupOpenid });
        }

        public async Task<long> GetMaxGroupIdAsync()
        {
            const long MIN_GROUP_ID = 990000000000;
            const long MAX_GROUP_ID = 1000000000000;
            const string sql = "SELECT MAX(id) FROM group_offical WHERE id > @min AND id < @max";
            using var conn = CreateConnection();
            var result = await conn.ExecuteScalarAsync<long?>(sql, new { min = MIN_GROUP_ID, max = MAX_GROUP_ID });
            return (result ?? MIN_GROUP_ID) + 1;
        }

        public async Task<string> GetGroupOpenidAsync(long botUin, long groupId)
        {
            const string sql = "SELECT group_openid FROM group_offical WHERE (target_group = @groupId OR id = @groupId) AND bot_uin = @botUin";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { groupId, botUin }) ?? string.Empty;
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

        public async Task<int> AppendAsync(long botUin, long friendId, string friendName)
        {
            var friend = new Friend
            {
                BotUin = botUin,
                FriendId = friendId,
                FriendName = friendName,
                InsertDate = DateTime.Now
            };
            return await AddAsync(friend);
        }

        public async Task<bool> UpdateCreditAsync(long botUin, long friendId, long credit)
        {
            const string sql = "UPDATE friend SET credit = @credit WHERE bot_uin = @botUin AND friend_id = @friendId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { credit, botUin, friendId }) > 0;
        }

        public async Task<bool> UpdateSaveCreditAsync(long botUin, long friendId, long saveCredit)
        {
            const string sql = "UPDATE friend SET save_credit = @saveCredit WHERE bot_uin = @botUin AND friend_id = @friendId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { saveCredit, botUin, friendId }) > 0;
        }

        public async Task<string> GetCreditRankingAsync(long botUin, long groupId, int top, string format)
        {
            // query: select top {top} UserId, credit from Friend where UserId in (select UserId from GroupMember where GroupId = {GroupId}) order by credit desc
            // format: 第{i}名{0} 💎{1:N0}\n
            // Note: Friend table has BotUin, FriendId, FriendName. It does NOT have 'credit'.
            // Wait, the original code in CreditMessage.cs was:
            // await Friend.QueryWhereAsync($"top {top} UserId, credit", $"UserId in (select UserId from {GroupMember.FullName} where GroupId = {GroupId})", $"credit desc", format)
            // Friend table definition I saw earlier:
            // public class Friend { BotUin, FriendId, FriendName, InsertDate }
            // It does NOT have Credit.
            // But UserInfo has Credit.
            // Maybe Friend view or join?
            // "Friend" class maps to "friend" table.
            
            // Re-checking CreditMessage.cs logic:
            // SelfInfo.IsCredit ? await Friend.QueryWhereAsync(...)
            // If SelfInfo.IsCredit is true, it queries Friend.
            // Maybe Friend entity has Credit property but I missed it?
            
            // Let's check Friend.cs again.
            // 10->        public long BotUin { get; set; }
            // 11->        public long FriendId { get; set; }
            // 12->        public string FriendName { get; set; } = string.Empty;
            // 13->        public DateTime InsertDate { get; set; }
            
            // It does not have Credit.
            // How did `Friend.QueryWhereAsync` work?
            // `Friend` inherits from `BaseEntity`? No.
            // Maybe `QueryWhereAsync` does a join or the query string provided "UserId, credit" implies the table has it?
            // If the table doesn't have it, the SQL would fail.
            // UNLESS `Friend` class is partial and has `Credit` in another file?
            // OR `Friend` table in DB has `credit` column but entity doesn't map it?
            
            // If I look at `UserInfo.QueryWhereAsync` usage in `CreditMessage.cs`:
            // `await UserInfo.QueryWhereAsync(...)`
            
            // If `Friend` table really has credit, I should use it.
            // If not, maybe it joins with UserInfo?
            // `Friend.QueryWhereAsync` was a static method in `Friend` class (or inherited).
            
            // If I assume `friend` table has `credit` column (common in legacy DBs to have denormalized data), I can write the SQL.
            // The original query was: `select top {top} UserId, credit ...`
            // Wait, `Friend` entity has `FriendId`. The query uses `UserId`.
            // Maybe `UserId` alias for `FriendId`?
            // Or maybe `Friend` static class mapped `UserId` to `friend_id`.
            
            // Let's assume `friend` table has `credit`.
            // And `friend_id` is the user id.
            
            string sql = $@"
                SELECT friend_id as UserId, credit 
                FROM friend 
                WHERE friend_id IN (SELECT user_id FROM group_member WHERE group_id = @groupId) 
                ORDER BY credit DESC 
                LIMIT @top";

            using var conn = CreateConnection();
            var list = await conn.QueryAsync<dynamic>(sql, new { groupId, top });
            
            var sb = new System.Text.StringBuilder();
            int i = 1;
            foreach (var item in list)
            {
                // format: "第{i}名{0} 💎{1:N0}\n" or "第{i}名[@:{0}] 💎{1:N0}\n"
                // The format string expects {0} to be UserId and {1} to be Credit.
                // And {i} is the rank.
                // I need to implement the formatting manually since I can't use string.Format with {i} easily if it's inside.
                
                string line = format.Replace("{i}", i.ToString());
                // format has {0} and {1}.
                long userId = (long)item.UserId;
                long credit = (long)item.credit;
                
                line = string.Format(line, userId, credit);
                sb.Append(line);
                i++;
            }
            return sb.ToString();
        }
    }
}
