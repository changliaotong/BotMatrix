using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Games;
using Dapper;
using Dapper.Contrib.Extensions;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class UserMarriageRepository : BaseRepository<UserMarriage>, IUserMarriageRepository
    {
        public UserMarriageRepository() : base("UserMarriages") { }
        protected override string KeyField => "Id";

        public async Task<UserMarriage?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserMarriages WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<UserMarriage>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserMarriage>(sql, new { userId });
        }

        public async Task<UserMarriage> GetOrCreateAsync(string userId, IDbTransaction? trans = null)
        {
            var m = await GetByUserIdAsync(userId, trans);
            if (m == null)
            {
                m = new UserMarriage { UserId = userId, Status = "single" };
                if (trans != null) await trans.Connection.InsertAsync(m, trans);
                else
                {
                    using var conn = CreateConnection();
                    await conn.InsertAsync(m);
                }
            }
            return m;
        }

        public async Task UpdateMarriageStatusAsync(string userId, string spouseId, string status, DateTime marriageDate, IDbTransaction? trans = null)
        {
            const string sql = "UPDATE UserMarriages SET Status = @status, SpouseId = @spouseId, MarriageDate = @marriageDate, UpdatedAt = @now WHERE UserId = @userId";
            var now = DateTime.Now;
            if (trans != null) await trans.Connection.ExecuteAsync(sql, new { userId, spouseId, status, marriageDate, now }, trans);
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { userId, spouseId, status, marriageDate, now });
            }
        }

        public async Task DivorceAsync(string userId, string spouseId, DateTime divorceDate, IDbTransaction? trans = null)
        {
            const string sql = "UPDATE UserMarriages SET Status = 'divorced', SpouseId = '', DivorceDate = @divorceDate, UpdatedAt = @now WHERE UserId = @userId";
            var now = DateTime.Now;
            if (trans != null)
            {
                await trans.Connection.ExecuteAsync(sql, new { userId, divorceDate, now }, trans);
            }
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { userId, divorceDate, now });
            }
        }
    }

    public class MarriageProposalRepository : BaseRepository<MarriageProposal>, IMarriageProposalRepository
    {
        public MarriageProposalRepository() : base("MarriageProposals") { }
        protected override string KeyField => "Id";

        public async Task<MarriageProposal?> GetPendingAsync(string recipientId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM MarriageProposals WHERE RecipientId = @recipientId AND Status = 'pending' ORDER BY CreatedAt DESC";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<MarriageProposal>(sql, new { recipientId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<MarriageProposal>(sql, new { recipientId });
        }

        public async Task UpdateStatusAsync(Guid id, string status, IDbTransaction? trans = null)
        {
            const string sql = "UPDATE MarriageProposals SET Status = @status, UpdatedAt = @now WHERE Id = @id";
            var now = DateTime.Now;
            if (trans != null) await trans.Connection.ExecuteAsync(sql, new { id, status, now }, trans);
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { id, status, now });
            }
        }
    }

    public class WeddingItemRepository : BaseRepository<WeddingItem>, IWeddingItemRepository
    {
        public WeddingItemRepository() : base("WeddingItems") { }
        protected override string KeyField => "Id";

        public async Task<WeddingItem?> GetByUserAndTypeAsync(string userId, string type, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM WeddingItems WHERE UserId = @userId AND ItemType = @type LIMIT 1";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<WeddingItem>(sql, new { userId, type }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<WeddingItem>(sql, new { userId, type });
        }
    }

    public class SweetHeartRepository : BaseRepository<SweetHeart>, ISweetHeartRepository
    {
        public SweetHeartRepository() : base("SweetHearts") { }
        protected override string KeyField => "Id";
    }

    public class BabyRepository : BaseRepository<Baby>, IBabyRepository
    {
        public BabyRepository() : base("Babies") { }
        protected override string KeyField => "Id";

        public async Task<Baby?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM Babies WHERE UserId = @userId AND Status = 'active'";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<Baby>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Baby>(sql, new { userId });
        }
    }

    public class BabyEventRepository : BaseRepository<BabyEvent>, IBabyEventRepository
    {
        public BabyEventRepository() : base("BabyEvents") { }
        protected override string KeyField => "Id";
    }

    public class BabyConfigRepository : BaseRepository<BabyConfig>, IBabyConfigRepository
    {
        public BabyConfigRepository() : base("BabyConfig") { }
        protected override string KeyField => "Id";

        public async Task<BabyConfig> GetAsync(IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM BabyConfig WHERE Id = 1";
            BabyConfig? config;
            if (trans != null)
            {
                config = await trans.Connection.QueryFirstOrDefaultAsync<BabyConfig>(sql, null, trans);
            }
            else
            {
                using var conn = CreateConnection();
                config = await conn.QueryFirstOrDefaultAsync<BabyConfig>(sql);
            }
            return config ?? new BabyConfig();
        }
    }

    public class GroupRepository : BaseRepository<GroupInfo>, IGroupRepository
    {
        public GroupRepository() : base("Groups") { }
        protected override string KeyField => "Id";

        public async Task<GroupInfo?> GetByGroupIdAsync(long groupId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM Groups WHERE Id = @groupId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<GroupInfo>(sql, new { groupId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<GroupInfo>(sql, new { groupId });
        }

        public async Task<string> GetValueAsync(string field, long groupId, IDbTransaction? trans = null)
        {
            string sql = $"SELECT {field} FROM Groups WHERE Id = @groupId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<string>(sql, new { groupId }, trans) ?? string.Empty;
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<string>(sql, new { groupId }) ?? string.Empty;
        }

        public async Task<int> SetValueAsync(string field, string value, long groupId, IDbTransaction? trans = null)
        {
            string sql = $"UPDATE Groups SET {field} = @value WHERE Id = @groupId";
            if (trans != null) return await trans.Connection.ExecuteAsync(sql, new { value, groupId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { value, groupId });
        }

        public async Task<long> GetGroupOwnerAsync(long groupId, long defaultValue = 0, IDbTransaction? trans = null)
        {
            const string sql = "SELECT GroupOwner FROM Groups WHERE Id = @groupId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long?>(sql, new { groupId }, trans) ?? defaultValue;
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long?>(sql, new { groupId }) ?? defaultValue;
        }

        public async Task<bool> GetIsCreditAsync(long groupId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT IsCredit FROM Groups WHERE Id = @groupId";
            if (trans != null) return (await trans.Connection.ExecuteScalarAsync<int?>(sql, new { groupId }, trans) ?? 0) == 1;
            using var conn = CreateConnection();
            return (await conn.ExecuteScalarAsync<int?>(sql, new { groupId }) ?? 0) == 1;
        }
    }

    public class RobberyRecordRepository : BaseRepository<RobberyRecord>, IRobberyRecordRepository
    {
        public RobberyRecordRepository() : base("RobberyRecords") { }
        protected override string KeyField => "Id";

        public async Task<DateTime> GetLastRobTimeAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT RobTime FROM RobberyRecords WHERE RobberId = @userId ORDER BY RobTime DESC LIMIT 1";
            DateTime? lastTime;
            if (trans != null) lastTime = await trans.Connection.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                lastTime = await conn.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId });
            }
            return lastTime ?? DateTime.MinValue;
        }

        public async Task<DateTime> GetProtectionEndTimeAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT RobTime FROM RobberyRecords WHERE VictimId = @userId AND IsSuccess = 1 ORDER BY RobTime DESC LIMIT 1";
            DateTime? lastTime;
            if (trans != null) lastTime = await trans.Connection.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                lastTime = await conn.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId });
            }
            if (lastTime == null) return DateTime.MinValue;
            return lastTime.Value.AddMinutes(30);
        }
    }

    public class UserPairingProfileRepository : BaseRepository<UserPairingProfile>, IUserPairingProfileRepository
    {
        public UserPairingProfileRepository() : base("UserPairingProfiles") { }
        protected override string KeyField => "Id";

        public async Task<UserPairingProfile?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserPairingProfiles WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<UserPairingProfile>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserPairingProfile>(sql, new { userId });
        }

        public async Task<List<UserPairingProfile>> GetActiveSeekersAsync(int limit = 10, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM UserPairingProfiles WHERE IsLooking = 1 ORDER BY LastActive DESC LIMIT {limit}";
            if (trans != null) return (await trans.Connection.QueryAsync<UserPairingProfile>(sql, null, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<UserPairingProfile>(sql)).ToList();
        }
    }

    public class PairingRecordRepository : BaseRepository<PairingRecord>, IPairingRecordRepository
    {
        public PairingRecordRepository() : base("PairingRecords") { }
        protected override string KeyField => "Id";

        public async Task<PairingRecord?> GetCurrentPairAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM PairingRecords WHERE (User1Id = @userId OR User2Id = @userId) AND Status = 'coupled'";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<PairingRecord>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<PairingRecord>(sql, new { userId });
        }
    }

    public class BrickRecordRepository : BaseRepository<BrickRecord>, IBrickRecordRepository
    {
        public BrickRecordRepository() : base("BrickRecords") { }
        protected override string KeyField => "Id";

        public async Task<DateTime> GetLastActionTimeAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT ActionTime FROM BrickRecords WHERE AttackerId = @userId ORDER BY ActionTime DESC LIMIT 1";
            DateTime? lastTime;
            if (trans != null) lastTime = await trans.Connection.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                lastTime = await conn.QueryFirstOrDefaultAsync<DateTime?>(sql, new { userId });
            }
            return lastTime ?? DateTime.MinValue;
        }

        public async Task<List<(string UserId, int Count)>> GetTopAttackersAsync(int limit = 10, IDbTransaction? trans = null)
        {
            string sql = $"SELECT AttackerId as UserId, COUNT(*) as Count FROM BrickRecords WHERE IsSuccess = 1 GROUP BY AttackerId ORDER BY Count DESC LIMIT {limit}";
            IEnumerable<dynamic> results;
            if (trans != null) results = await trans.Connection.QueryAsync(sql, null, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync(sql);
            }
            return results.Select(r => ((string)r.UserId, (int)r.Count)).ToList();
        }
    }

    public class GiftRepository : BaseRepository<Gift>, IGiftRepository
    {
        public GiftRepository() : base("Gift") { }
        protected override string KeyField => "Id";

        public async Task<long> GetGiftIdAsync(string giftName, IDbTransaction? trans = null)
        {
            const string sql = "SELECT Id FROM Gift WHERE GiftName = @giftName";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { giftName }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { giftName });
        }

        public async Task<long> GetRandomGiftAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            const string sql = "SELECT Id FROM Gift WHERE GiftCredit < (SELECT Credit FROM user_info WHERE BotUin = @botUin AND GroupId = @groupId AND Id = @qq) ORDER BY RANDOM() LIMIT 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { botUin, groupId, qq }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { botUin, groupId, qq });
        }

        public async Task<string> GetGiftListAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null)
        {
            const string sql = "SELECT GiftName, GiftCredit FROM Gift WHERE IsValid = 1 AND GiftCredit <= (SELECT Credit FROM user_info WHERE BotUin = @botUin AND GroupId = @groupId AND Id = @qq) ORDER BY RANDOM() LIMIT 5";
            IEnumerable<dynamic> results;
            if (trans != null) results = await trans.Connection.QueryAsync(sql, new { botUin, groupId, qq }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync(sql, new { botUin, groupId, qq });
            }
            if (!results.Any())
            {
                const string sqlFallback = "SELECT GiftName, GiftCredit FROM Gift WHERE IsValid = 1 AND GiftCredit < 10000 ORDER BY RANDOM() LIMIT 5";
                if (trans != null) results = await trans.Connection.QueryAsync(sqlFallback, null, trans);
                else
                {
                    using var conn = CreateConnection();
                    results = await conn.QueryAsync(sqlFallback);
                }
            }
            return string.Join("\n", results.Select(r => $"{r.GiftName}={r.GiftCredit}分"));
        }
    }

    public class GiftLogRepository : BaseRepository<GiftRecord>, IGiftLogRepository
    {
        public GiftLogRepository() : base("GiftLog") { }
        protected override string KeyField => "Id";
    }

    public class GiftStoreItemRepository : BaseRepository<GiftStoreItem>, IGiftStoreItemRepository
    {
        public GiftStoreItemRepository() : base("GiftStoreItem") { }
        protected override string KeyField => "Id";

        public async Task<List<GiftStoreItem>> GetValidGiftsAsync(IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM GiftStoreItem WHERE IsValid = 1";
            IEnumerable<GiftStoreItem> results;
            if (trans != null) results = await trans.Connection.QueryAsync<GiftStoreItem>(sql, null, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<GiftStoreItem>(sql);
            }
            return results.ToList();
        }

        public async Task<GiftStoreItem?> GetByNameAsync(string name, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM GiftStoreItem WHERE GiftName = @name";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<GiftStoreItem>(sql, new { name }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<GiftStoreItem>(sql, new { name });
        }
    }

    public class GiftBackpackRepository : BaseRepository<GiftBackpack>, IGiftBackpackRepository
    {
        public GiftBackpackRepository() : base("GiftBackpack") { }
        protected override string KeyField => "Id";

        public async Task<List<GiftBackpack>> GetUserBackpackAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM GiftBackpack WHERE UserId = @userId AND ItemCount > 0";
            IEnumerable<GiftBackpack> results;
            if (trans != null) results = await trans.Connection.QueryAsync<GiftBackpack>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<GiftBackpack>(sql, new { userId });
            }
            return results.ToList();
        }

        public async Task<GiftBackpack?> GetItemAsync(string userId, long giftId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM GiftBackpack WHERE UserId = @userId AND GiftId = @giftId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<GiftBackpack>(sql, new { userId, giftId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<GiftBackpack>(sql, new { userId, giftId });
        }
    }

    public class GroupGiftRepository : BaseRepository<GroupGift>, IGroupGiftRepository
    {
        public GroupGiftRepository() : base("GroupGift") { }
        protected override string KeyField => "Id";
    }

    public class VehicleRepository : BaseRepository<Vehicle>, IVehicleRepository
    {
        public VehicleRepository() : base("UserVehicles") { }
        protected override string KeyField => "Id";

        public async Task<Vehicle?> GetActiveVehicleAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserVehicles WHERE UserId = @userId AND Status = @status";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<Vehicle>(sql, new { userId, status = (int)VehicleStatus.Driving }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Vehicle>(sql, new { userId, status = (int)VehicleStatus.Driving });
        }

        public async Task<List<Vehicle>> GetUserVehiclesAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserVehicles WHERE UserId = @userId";
            IEnumerable<Vehicle> results;
            if (trans != null) results = await trans.Connection.QueryAsync<Vehicle>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<Vehicle>(sql, new { userId });
            }
            return results.ToList();
        }
    }

    public class PetRepository : BaseRepository<Pet>, IPetRepository
    {
        public PetRepository() : base("UserPets") { }
        protected override string KeyField => "Id";

        public async Task<Pet?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserPets WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<Pet>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Pet>(sql, new { userId });
        }
    }

    public class PetInventoryRepository : BaseRepository<PetInventory>, IPetInventoryRepository
    {
        public PetInventoryRepository() : base("UserPetInventory") { }
        protected override string KeyField => "Id";

        public async Task<List<PetInventory>> GetByUserAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserPetInventory WHERE UserId = @userId AND Count > 0";
            IEnumerable<PetInventory> results;
            if (trans != null) results = await trans.Connection.QueryAsync<PetInventory>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<PetInventory>(sql, new { userId });
            }
            return results.ToList();
        }

        public async Task<PetInventory?> GetItemAsync(string userId, string itemId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserPetInventory WHERE UserId = @userId AND ItemId = @itemId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<PetInventory>(sql, new { userId, itemId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<PetInventory>(sql, new { userId, itemId });
        }
    }

    public class MountRepository : BaseRepository<Mount>, IMountRepository
    {
        public MountRepository() : base("UserMounts") { }
        protected override string KeyField => "Id";

        public async Task<Mount?> GetActiveMountAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserMounts WHERE UserId = @userId AND Status = @status";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<Mount>(sql, new { userId, status = (int)MountStatus.Riding }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<Mount>(sql, new { userId, status = (int)MountStatus.Riding });
        }

        public async Task<List<Mount>> GetUserMountsAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserMounts WHERE UserId = @userId";
            IEnumerable<Mount> results;
            if (trans != null) results = await trans.Connection.QueryAsync<Mount>(sql, new { userId }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<Mount>(sql, new { userId });
            }
            return results.ToList();
        }
    }

    public class CultivationProfileRepository : BaseRepository<CultivationProfile>, ICultivationProfileRepository
    {
        public CultivationProfileRepository() : base("CultivationProfiles") { }
        protected override string KeyField => "Id";

        public async Task<CultivationProfile?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM CultivationProfiles WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<CultivationProfile>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<CultivationProfile>(sql, new { userId });
        }

        public async Task<List<CultivationProfile>> GetTopCultivatorsAsync(int limit = 10, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM CultivationProfiles ORDER BY Level DESC, Exp DESC LIMIT @limit";
            IEnumerable<CultivationProfile> results;
            if (trans != null) results = await trans.Connection.QueryAsync<CultivationProfile>(sql, new { limit }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync<CultivationProfile>(sql, new { limit });
            }
            return results.ToList();
        }
    }

    public class CultivationRecordRepository : BaseRepository<CultivationRecord>, ICultivationRecordRepository
    {
        public CultivationRecordRepository() : base("CultivationRecords") { }
        protected override string KeyField => "Id";
    }

    public class BuyFriendsRepository : BaseRepository<BuyFriends>, IBuyFriendsRepository
    {
        public BuyFriendsRepository() : base("BuyFriends") { }
        protected override string KeyField => "Id";

        public async Task<long> GetCurrMasterAsync(long groupId, long friendId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT UserId FROM BuyFriends WHERE GroupId = @groupId AND FriendId = @friendId AND IsValid = 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { groupId, friendId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, friendId });
        }

        public async Task<long> GetSellPriceAsync(long groupId, long friendId, IDbTransaction? trans = null)
        {
            const long minPrice = 100;
            const string sql = "SELECT get_sell_price(SellPrice, InsertDate) FROM BuyFriends WHERE GroupId = @groupId AND FriendId = @friendId AND IsValid = 1";
            long? price;
            if (trans != null) price = await trans.Connection.ExecuteScalarAsync<long?>(sql, new { groupId, friendId }, trans);
            else
            {
                using var conn = CreateConnection();
                price = await conn.ExecuteScalarAsync<long?>(sql, new { groupId, friendId });
            }
            return (price ?? minPrice) < minPrice ? minPrice : (price ?? minPrice);
        }

        public async Task<long> GetBuyPriceAsync(long groupId, long friendId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT BuyPrice FROM BuyFriends WHERE GroupId = @groupId AND FriendId = @friendId AND IsValid = 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { groupId, friendId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, friendId });
        }

        public async Task<int> GetBuyIdAsync(long groupId, long friendId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT Id FROM BuyFriends WHERE GroupId = @groupId AND FriendId = @friendId AND IsValid = 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { groupId, friendId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, friendId });
        }

        public async Task<long> GetPetCountAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) FROM BuyFriends WHERE GroupId = @groupId AND UserId = @userId AND IsValid = 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { groupId, userId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<List<(long FriendId, long SellPrice)>> GetPriceListAsync(long groupId, int limit, IDbTransaction? trans = null)
        {
            string sql = $"SELECT FriendId, get_sell_price(SellPrice, InsertDate) AS SellPrice FROM BuyFriends WHERE GroupId = @groupId AND IsValid = 1 ORDER BY SellPrice DESC LIMIT @limit";
            IEnumerable<dynamic> results;
            if (trans != null) results = await trans.Connection.QueryAsync(sql, new { groupId, limit }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync(sql, new { groupId, limit });
            }
            return results.Select(r => ((long)r.friendid, (long)r.sellprice)).ToList();
        }

        public async Task<List<(long GroupId, long SellPrice)>> GetMyPriceListAsync(long userId, int limit, IDbTransaction? trans = null)
        {
            string sql = $"SELECT GroupId, get_sell_price(SellPrice, InsertDate) AS SellPrice FROM BuyFriends WHERE IsValid = 1 AND FriendId = @userId ORDER BY SellPrice DESC LIMIT @limit";
            IEnumerable<dynamic> results;
            if (trans != null) results = await trans.Connection.QueryAsync(sql, new { userId, limit }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync(sql, new { userId, limit });
            }
            return results.Select(r => ((long)r.groupid, (long)r.sellprice)).ToList();
        }

        public async Task<int> GetRankAsync(long groupId, long sellPrice, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) + 1 FROM BuyFriends WHERE GroupId = @groupId AND IsValid = 1 AND get_sell_price(SellPrice, InsertDate) > @sellPrice";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { groupId, sellPrice }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { groupId, sellPrice });
        }

        public async Task<List<(long FriendId, long SellPrice)>> GetMyPetListAsync(long groupId, long userId, int limit, IDbTransaction? trans = null)
        {
            string sql = $"SELECT FriendId, get_sell_price(SellPrice, InsertDate) AS SellPrice FROM BuyFriends WHERE GroupId = @groupId AND UserId = @userId AND IsValid = 1 ORDER BY SellPrice DESC LIMIT @limit";
            IEnumerable<dynamic> results;
            if (trans != null) results = await trans.Connection.QueryAsync(sql, new { groupId, userId, limit }, trans);
            else
            {
                using var conn = CreateConnection();
                results = await conn.QueryAsync(sql, new { groupId, userId, limit });
            }
            return results.Select(r => ((long)r.friendid, (long)r.sellprice)).ToList();
        }
    }

    public class UserModuleAccessRepository : BaseRepository<UserModuleAccess>, IUserModuleAccessRepository
    {
        public UserModuleAccessRepository() : base("UserModuleAccess") { }
        protected override string KeyField => "Id";

        public async Task<List<UserModuleAccess>> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserModuleAccess WHERE UserId = @userId";
            if (trans != null) return (await trans.Connection.QueryAsync<UserModuleAccess>(sql, new { userId }, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<UserModuleAccess>(sql, new { userId })).ToList();
        }

        public async Task<UserModuleAccess?> GetAsync(string userId, string moduleId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserModuleAccess WHERE UserId = @userId AND ModuleId = @moduleId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<UserModuleAccess>(sql, new { userId, moduleId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserModuleAccess>(sql, new { userId, moduleId });
        }
    }

    public class UserLevelRepository : BaseRepository<UserLevel>, IUserLevelRepository
    {
        public UserLevelRepository() : base("UserLevels") { }
        protected override string KeyField => "Id";

        public async Task<UserLevel?> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserLevels WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<UserLevel>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserLevel>(sql, new { userId });
        }

        public async Task<List<UserLevel>> GetTopRankingsAsync(int limit = 10, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM UserLevels ORDER BY Experience DESC LIMIT {limit}";
            if (trans != null) return (await trans.Connection.QueryAsync<UserLevel>(sql, null, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<UserLevel>(sql)).ToList();
        }
    }

    public class SecretLoveRepository : BaseRepository<SecretLove>, ISecretLoveRepository
    {
        public SecretLoveRepository() : base("Love") { }
        protected override string KeyField => "UserId";

        public async Task<string> GetLoveStatusAsync(IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(DISTINCT UserId) as UserCount, COUNT(LoveId) as LoveCount FROM Love";
            dynamic? result;
            if (trans != null) result = await trans.Connection.QueryFirstOrDefaultAsync(sql, null, trans);
            else
            {
                using var conn = CreateConnection();
                result = await conn.QueryFirstOrDefaultAsync(sql);
            }
            return $"已有{result?.UserCount ?? 0}人登记暗恋对象{result?.LoveCount ?? 0}个。";
        }

        public async Task<bool> IsLoveEachotherAsync(long userId, long loveId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) FROM Love WHERE (UserId = @userId AND LoveId = @loveId) OR (UserId = @loveId AND LoveId = @userId)";
            int count;
            if (trans != null) count = await trans.Connection.ExecuteScalarAsync<int>(sql, new { userId, loveId }, trans);
            else
            {
                using var conn = CreateConnection();
                count = await conn.ExecuteScalarAsync<int>(sql, new { userId, loveId });
            }
            return count >= 2;
        }

        public async Task<int> GetCountLoveAsync(long userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) FROM Love WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId });
        }

        public async Task<int> GetCountLoveMeAsync(long userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) FROM Love WHERE LoveId = @userId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { userId });
        }
    }

    public class ShuffledDeckRepository : BaseRepository<ShuffledDeck>, IShuffledDeckRepository
    {
        public ShuffledDeckRepository() : base("ShuffledDeck") { }
        protected override string KeyField => "DeckId";

        public async Task ClearShuffledDeckAsync(long groupId, IDbTransaction? trans = null)
        {
            const string sql = "DELETE FROM ShuffledDeck WHERE GroupId = @groupId";
            if (trans != null) await trans.Connection.ExecuteAsync(sql, new { groupId }, trans);
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { groupId });
            }
        }

        public async Task ClearShuffledDeckAsync(long groupId, long id, IDbTransaction? trans = null)
        {
            const string sql = "DELETE FROM ShuffledDeck WHERE GroupId = @groupId AND Id = @id";
            if (trans != null) await trans.Connection.ExecuteAsync(sql, new { groupId, id }, trans);
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { groupId, id });
            }
        }

        public async Task ClearShuffledDeckAsync(long groupId, List<int> ids, IDbTransaction? trans = null)
        {
            if (ids == null || ids.Count == 0) return;
            const string sql = "DELETE FROM ShuffledDeck WHERE GroupId = @groupId AND Id IN @ids";
            if (trans != null) await trans.Connection.ExecuteAsync(sql, new { groupId, ids }, trans);
            else
            {
                using var conn = CreateConnection();
                await conn.ExecuteAsync(sql, new { groupId, ids });
            }
        }

        public async Task<List<Card>> ReadShuffledDeckAsync(long groupId, IDbTransaction? trans = null, bool lockRow = false)
        {
            string lockSql = lockRow ? " WITH (UPDLOCK, ROWLOCK) " : "";
            string sql = $"SELECT Id, Rank, Suit FROM ShuffledDeck{lockSql} WHERE GroupId = @groupId ORDER BY DeckOrder";
            
            IEnumerable<Card> result;
            if (trans != null) result = await trans.Connection.QueryAsync<Card>(sql, new { groupId }, trans);
            else
            {
                using var conn = CreateConnection();
                result = await conn.QueryAsync<Card>(sql, new { groupId });
            }
            return result.ToList();
        }

        public async Task<bool> IsShuffledDeckExistsAsync(long groupId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(*) FROM ShuffledDeck WHERE GroupId = @groupId";
            int count;
            if (trans != null) count = await trans.Connection.ExecuteScalarAsync<int>(sql, new { groupId }, trans);
            else
            {
                using var conn = CreateConnection();
                count = await conn.ExecuteScalarAsync<int>(sql, new { groupId });
            }
            return count >= 6;
        }
    }

    public class BlockRepository : BaseRepository<Block>, IBlockRepository
    {
        public BlockRepository() : base("Block") { }
        protected override string KeyField => "Id";

        public async Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            string sql = "SELECT COALESCE(MAX(Id), 0) FROM Block WHERE GroupId = @groupId AND IsOpen = 0";
            if (groupId == 0) sql += " AND UserId = @userId";

            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { groupId, userId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { groupId, userId });
        }

        public async Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null)
        {
            if (blockId == 0) return string.Empty;
            const string sql = "SELECT BlockHash FROM Block WHERE Id = @blockId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<string>(sql, new { blockId }, trans) ?? string.Empty;
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { blockId }) ?? string.Empty;
        }

        public async Task<long> GetBlockIdAsync(string hash, IDbTransaction? trans = null)
        {
            const string sql = "SELECT Id FROM Block WHERE BlockHash = @hash";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, new { hash }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, new { hash });
        }

        public async Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT BlockNum FROM Block WHERE Id = @blockId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { blockId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { blockId });
        }

        public async Task<string> GetValueAsync(string field, long blockId, IDbTransaction? trans = null)
        {
            string sql = $"SELECT {field} FROM Block WHERE Id = @blockId";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<string>(sql, new { blockId }, trans) ?? string.Empty;
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<string>(sql, new { blockId }) ?? string.Empty;
        }

        public async Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT IsOpen FROM Block WHERE Id = @blockId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<bool>(sql, new { blockId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<bool>(sql, new { blockId });
        }
    }

    public class BlockTypeRepository : BaseRepository<BlockType>, IBlockTypeRepository
    {
        public BlockTypeRepository() : base("BlockType") { }
        protected override string KeyField => "Id";

        public async Task<int> GetTypeIdAsync(string typeName, IDbTransaction? trans = null)
        {
            const string sql = "SELECT Id FROM BlockType WHERE TypeName = @typeName";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, new { typeName }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { typeName });
        }

        public async Task<decimal> GetOddsAsync(int typeId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT BlockOdds FROM BlockType WHERE Id = @typeId";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<decimal>(sql, new { typeId }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<decimal>(sql, new { typeId });
        }
    }

    public class BlockWinRepository : BaseRepository<BlockWin>, IBlockWinRepository
    {
        public BlockWinRepository() : base("BlockWin") { }
        protected override string KeyField => "Id";

        public async Task<bool> IsWinAsync(int typeId, int blockNum, IDbTransaction? trans = null)
        {
            const string sql = "SELECT IsWin FROM BlockWin WHERE TypeId = @typeId AND BlockNum = @blockNum";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<bool>(sql, new { typeId, blockNum }, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<bool>(sql, new { typeId, blockNum });
        }
    }

    public class BlockRandomRepository : BaseRepository<BlockRandom>, IBlockRandomRepository
    {
        public BlockRandomRepository() : base("BlockRandom") { }
        protected override string KeyField => "Id";

        public async Task<int> GetRandomNumAsync(IDbTransaction? trans = null)
        {
            const string sql = "SELECT BlockNum FROM BlockRandom WHERE Id = (ABS(RANDOM()) % 216) + 1";
            if (trans != null) return await trans.Connection.ExecuteScalarAsync<int>(sql, null, trans);
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql);
        }
    }

    public class DigitalStaffRepository : BaseRepository<DigitalStaff>, IDigitalStaffRepository
    {
        public DigitalStaffRepository() : base("DigitalStaff") { }
        protected override string KeyField => "Id";

        public async Task<DigitalStaff?> GetByGuidAsync(Guid guid, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM DigitalStaff WHERE Guid = @guid";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<DigitalStaff>(sql, new { guid }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<DigitalStaff>(sql, new { guid });
        }

        public async Task<List<DigitalStaff>> GetByOwnerAsync(string ownerUserId, string? status = null, IDbTransaction? trans = null)
        {
            string sql = "SELECT * FROM DigitalStaff WHERE OwnerUserId = @ownerUserId";
            if (!string.IsNullOrEmpty(status)) sql += " AND CurrentStatus = @status";
            
            if (trans != null) return (await trans.Connection.QueryAsync<DigitalStaff>(sql, new { ownerUserId, status }, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<DigitalStaff>(sql, new { ownerUserId, status })).ToList();
        }
    }

    public class CognitiveMemoryRepository : BaseRepository<CognitiveMemory>, ICognitiveMemoryRepository
    {
        public CognitiveMemoryRepository() : base("CognitiveMemories") { }
        protected override string KeyField => "Id";

        public async Task<IEnumerable<CognitiveMemory>> GetByStaffAsync(string staffId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM CognitiveMemories WHERE StaffId = @staffId ORDER BY Category, CreateTime ASC";
            if (trans != null) return await trans.Connection.QueryAsync<CognitiveMemory>(sql, new { staffId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryAsync<CognitiveMemory>(sql, new { staffId });
        }

        public async Task DeleteByStaffAsync(string staffId, IDbTransaction? trans = null)
        {
            const string sql = "DELETE FROM CognitiveMemories WHERE StaffId = @staffId";
            if (trans != null)
            {
                await trans.Connection.ExecuteAsync(sql, new { staffId }, trans);
                return;
            }
            using var conn = CreateConnection();
            await conn.ExecuteAsync(sql, new { staffId });
        }
    }

    public class StaffKpiRepository : BaseRepository<StaffKpi>, IStaffKpiRepository
    {
        public StaffKpiRepository() : base("StaffKpis") { }
        protected override string KeyField => "Id";

        public async Task<List<StaffKpi>> GetByStaffAsync(string staffId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM StaffKpis WHERE StaffId = @staffId ORDER BY CreateTime DESC";
            if (trans != null) return (await trans.Connection.QueryAsync<StaffKpi>(sql, new { staffId }, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<StaffKpi>(sql, new { staffId })).ToList();
        }
    }

    public class StaffTaskRepository : BaseRepository<StaffTask>, IStaffTaskRepository
    {
        public StaffTaskRepository() : base("StaffTasks") { }
        protected override string KeyField => "Id";

        public async Task<StaffTask?> GetByGuidAsync(Guid guid, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM StaffTasks WHERE Guid = @guid";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<StaffTask>(sql, new { guid }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<StaffTask>(sql, new { guid });
        }

        public async Task<List<StaffTask>> GetPendingTasksAsync(IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM StaffTasks WHERE Status = 'Pending' ORDER BY CreateTime ASC";
            if (trans != null) return (await trans.Connection.QueryAsync<StaffTask>(sql, null, trans)).ToList();
            using var conn = CreateConnection();
            return (await conn.QueryAsync<StaffTask>(sql)).ToList();
        }
    }

    public class UserMetricRepository : BaseRepository<UserMetric>, IUserMetricRepository
    {
        public UserMetricRepository() : base("UserMetrics") { }
        protected override string KeyField => "Id";

        public async Task<UserMetric?> GetAsync(string userId, string key, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserMetrics WHERE UserId = @userId AND MetricKey = @key";
            if (trans != null) return await trans.Connection.QueryFirstOrDefaultAsync<UserMetric>(sql, new { userId, key }, trans);
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<UserMetric>(sql, new { userId, key });
        }

        public async Task<UserMetric> GetOrCreateAsync(string userId, string key, IDbTransaction? trans = null)
        {
            var metric = await GetAsync(userId, key, trans);
            if (metric == null)
            {
                metric = new UserMetric
                {
                    Id = $"{userId}_{key}",
                    UserId = userId,
                    MetricKey = key,
                    Value = 0,
                    LastUpdateTime = DateTime.Now
                };
                await InsertAsync(metric, trans);
            }
            return metric;
        }
    }

    public class UserAchievementRepository : BaseRepository<UserAchievement>, IUserAchievementRepository
    {
        public UserAchievementRepository() : base("UserAchievements") { }
        protected override string KeyField => "Id";

        public async Task<bool> IsUnlockedAsync(string userId, string achievementId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT COUNT(1) FROM UserAchievements WHERE UserId = @userId AND AchievementId = @achievementId";
            int count;
            if (trans != null) count = await trans.Connection.ExecuteScalarAsync<int>(sql, new { userId, achievementId }, trans);
            else
            {
                using var conn = CreateConnection();
                count = await conn.ExecuteScalarAsync<int>(sql, new { userId, achievementId });
            }
            return count > 0;
        }

        public async Task<IEnumerable<UserAchievement>> GetByUserIdAsync(string userId, IDbTransaction? trans = null)
        {
            const string sql = "SELECT * FROM UserAchievements WHERE UserId = @userId";
            if (trans != null) return await trans.Connection.QueryAsync<UserAchievement>(sql, new { userId }, trans);
            using var conn = CreateConnection();
            return await conn.QueryAsync<UserAchievement>(sql, new { userId });
        }
    }
}
