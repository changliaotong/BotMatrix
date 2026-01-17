using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Modules.Games;
using Dapper.Contrib.Extensions;

namespace BotWorker.Domain.Repositories
{
    public interface IUserMarriageRepository : IBaseRepository<UserMarriage>
    {
        Task<UserMarriage?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
        Task<UserMarriage> GetOrCreateAsync(string userId, IDbTransaction? trans = null);
        Task UpdateMarriageStatusAsync(string userId, string spouseId, string status, DateTime marriageDate, IDbTransaction? trans = null);
        Task DivorceAsync(string userId, string spouseId, DateTime divorceDate, IDbTransaction? trans = null);
    }

    public interface IMarriageProposalRepository : IBaseRepository<MarriageProposal>
    {
        Task<MarriageProposal?> GetPendingAsync(string recipientId, IDbTransaction? trans = null);
        Task UpdateStatusAsync(Guid id, string status, IDbTransaction? trans = null);
    }

    public interface IWeddingItemRepository : IBaseRepository<WeddingItem>
    {
        Task<WeddingItem?> GetByUserAndTypeAsync(string userId, string type, IDbTransaction? trans = null);
    }

    public interface ISweetHeartRepository : IBaseRepository<SweetHeart>
    {
    }

    public interface IBabyRepository : IBaseRepository<Baby>
    {
        Task<Baby?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IBabyEventRepository : IBaseRepository<BabyEvent>
    {
    }

    public interface IBabyConfigRepository : IBaseRepository<BabyConfig>
    {
        Task<BabyConfig> GetAsync(IDbTransaction? trans = null);
    }

    public interface IGroupRepository : IBaseRepository<GroupInfo>
    {
        Task<GroupInfo?> GetByGroupIdAsync(long groupId, IDbTransaction? trans = null);
        Task<string> GetValueAsync(string field, long groupId, IDbTransaction? trans = null);
        Task<int> SetValueAsync(string field, string value, long groupId, IDbTransaction? trans = null);
        Task<long> GetGroupOwnerAsync(long groupId, long defaultValue = 0, IDbTransaction? trans = null);
        Task<bool> GetIsCreditAsync(long groupId, IDbTransaction? trans = null);
        Task<GroupInfo?> GetByOpenIdAsync(string openId, long botUin);
        Task<long> AddAsync(GroupInfo group);
        Task<bool> UpdateAsync(GroupInfo group);
        Task<bool> GetIsPetAsync(long groupId);
        Task<int> SetPowerOnAsync(long groupId, IDbTransaction? trans = null);
        Task<int> SetPowerOffAsync(long groupId, IDbTransaction? trans = null);
        Task<int> StartCyGameAsync(int state, string lastChengyu, long groupId);
        Task<int> GetChengyuIdleMinutesAsync(long groupId);
        Task<bool> GetPowerOnAsync(long groupId, IDbTransaction? trans = null);
        Task<int> SetRobotOwnerAsync(long groupId, long ownerId, IDbTransaction? trans = null);
        Task<long> GetRobotOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null);
        Task<bool> IsOwnerAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> IsPowerOffAsync(long groupId, IDbTransaction? trans = null);
        Task<bool> GetIsValidAsync(long groupId, IDbTransaction? trans = null);
        Task<string> GetRobotOwnerNameAsync(long groupId, string botName = "");
        Task<bool> IsCanTrialAsync(long groupId);
        Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0);
        Task<int> SetHintDateAsync(long groupId);
        Task<bool> GetIsWhiteAsync(long groupId);
        Task<string> GetIsBlockResAsync(long groupId);
        Task<bool> GetIsBlockAsync(long groupId);
        Task<int> GetIsOpenAsync(long groupId);
        Task<int> GetLastHintTimeAsync(long groupId);
        Task<int> CloudAnswerAsync(long groupId);
        Task<string> CloudAnswerResAsync(long groupId);
        Task<bool> GetIsBlackExitAsync(long groupId);
        Task<bool> GetIsBlackKickAsync(long groupId);
        Task<string> GetClosedFuncAsync(long groupId);
        Task<string> GetClosedRegexAsync(long groupId);
        Task<bool> GetIsExitHintAsync(long groupId);
        Task<bool> GetIsKickHintAsync(long groupId);
        Task<bool> GetIsRequirePrefixAsync(long groupId);
        Task<string> GetJoinResAsync(long groupId);
        Task<string> GetSystemPromptAsync(long groupId);
        Task<string> GetAdminRightResAsync(long groupId);
        Task<string> GetRightResAsync(long groupId);
        Task<string> GetTeachRightResAsync(long groupId);
        Task<int> SetInGameAsync(int isInGame, long groupId);
        Task<string> GetWelcomeResAsync(long groupId);
        Task<string> GetGroupNameAsync(long groupId);
        Task<string> GetGroupOwnerNicknameAsync(long groupId);
        Task<bool> GetIsAIAsync(long groupId);
        Task<bool> GetIsOwnerPayAsync(long groupId);
        Task<int> GetContextCountAsync(long groupId);
        Task<bool> GetIsMultAIAsync(long groupId);
        Task<bool> GetIsUseKnowledgebaseAsync(long groupId);
        Task<int> AppendAsync(long groupId, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "");
        Task<bool> GetIsNoLogAsync(long groupId);
        Task<bool> GetIsNoCheckAsync(long groupId);
        Task<bool> GetIsHintCloseAsync(long groupId);
        Task<long> GetSourceGroupIdAsync(long groupId);
        Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0);
        Task<long> GetSourceGroupIdAsync(long botUin, long groupId);
        Task<int> SetIsOpenAsync(bool isOpen, long groupId);
        Task<int> SetPowerOnAsync(bool isOpen, long groupId);
        Task<bool> GetPowerOnAsync(long groupId);
        Task<string> GetSystemPromptStatusAsync(long groupId);
        Task<string> GetVipResAsync(long groupId);
        Task<int> StartCyGameAsync(long groupId);
        Task<int> UpdateIsPowerOnAsync(long groupId, bool isPowerOn, IDbTransaction? trans = null);
        Task<int> UpdateAdminRightAsync(long groupId, int adminRight);
        Task<int> UpdateUseRightAsync(long groupId, int useRight);
        Task<int> UpdateTeachRightAsync(long groupId, int teachRight);
        Task<int> UpdateBlockMinAsync(long groupId, int blockMin);
        Task<int> UpdateJoinGroupSettingsAsync(long groupId, int isAccept, string rejectMessage, string regexRequestJoin);
        Task<int> UpdateIsChangeHintAsync(long groupId, bool isChangeHint);
        Task<int> UpdateWelcomeMessageAsync(long groupId, string message);
        Task<int> UpdateIsWelcomeHintAsync(long groupId, bool isWelcomeHint);
        Task<int> UpdateSystemPromptAsync(long groupId, string systemPrompt);
        Task<int> UpdateReplyModeAsync(long groupId, int modeReply);
        Task<int> UpdateCloseRegexAsync(long groupId, string closeRegex);
        Task<string> GetCreditRankingAsync(long groupId, int top, string format);
    }

    public interface IDigitalStaffRepository : IBaseRepository<DigitalStaff>
    {
        Task<DigitalStaff?> GetByGuidAsync(Guid guid, IDbTransaction? trans = null);
        Task<List<DigitalStaff>> GetByOwnerAsync(string ownerUserId, string? status = null, IDbTransaction? trans = null);
    }

    public interface ICognitiveMemoryRepository : IBaseRepository<CognitiveMemory>
    {
        Task<IEnumerable<CognitiveMemory>> GetByStaffAsync(string staffId, IDbTransaction? trans = null);
        Task DeleteByStaffAsync(string staffId, IDbTransaction? trans = null);
    }

    public interface IStaffKpiRepository : IBaseRepository<StaffKpi>
    {
        Task<List<StaffKpi>> GetByStaffAsync(string staffId, IDbTransaction? trans = null);
    }

    public interface IStaffTaskRepository : IBaseRepository<StaffTask>
    {
        Task<StaffTask?> GetByGuidAsync(Guid guid, IDbTransaction? trans = null);
        Task<List<StaffTask>> GetPendingTasksAsync(IDbTransaction? trans = null);
    }

    public interface IRobberyRecordRepository : IBaseRepository<RobberyRecord>
    {
        Task<DateTime> GetLastRobTimeAsync(string userId, IDbTransaction? trans = null);
        Task<DateTime> GetProtectionEndTimeAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IUserPairingProfileRepository : IBaseRepository<UserPairingProfile>
    {
        Task<UserPairingProfile?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
        Task<List<UserPairingProfile>> GetActiveSeekersAsync(int limit = 10, IDbTransaction? trans = null);
    }

    public interface IPairingRecordRepository : IBaseRepository<PairingRecord>
    {
        Task<PairingRecord?> GetCurrentPairAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IBrickRecordRepository : IBaseRepository<BrickRecord>
    {
        Task<DateTime> GetLastActionTimeAsync(string userId, IDbTransaction? trans = null);
        Task<List<(string UserId, int Count)>> GetTopAttackersAsync(int limit = 10, IDbTransaction? trans = null);
    }

    public interface IUserMetricRepository : IBaseRepository<UserMetric>
    {
        Task<UserMetric?> GetAsync(string userId, string key, IDbTransaction? trans = null);
        Task<UserMetric> GetOrCreateAsync(string userId, string key, IDbTransaction? trans = null);
    }

    public interface IUserAchievementRepository : IBaseRepository<UserAchievement>
    {
        Task<bool> IsUnlockedAsync(string userId, string achievementId, IDbTransaction? trans = null);
        Task<IEnumerable<UserAchievement>> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IGiftRepository : IBaseRepository<Gift>
    {
        Task<long> GetGiftIdAsync(string giftName, IDbTransaction? trans = null);
        Task<long> GetRandomGiftAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<string> GetGiftListAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
    }

    public interface IGiftLogRepository : IBaseRepository<GiftRecord>
    {
    }

    public interface IGiftStoreItemRepository : IBaseRepository<GiftStoreItem>
    {
        Task<List<GiftStoreItem>> GetValidGiftsAsync(IDbTransaction? trans = null);
        Task<GiftStoreItem?> GetByNameAsync(string name, IDbTransaction? trans = null);
    }

    public interface IGiftBackpackRepository : IBaseRepository<GiftBackpack>
    {
        Task<List<GiftBackpack>> GetUserBackpackAsync(string userId, IDbTransaction? trans = null);
        Task<GiftBackpack?> GetItemAsync(string userId, long giftId, IDbTransaction? trans = null);
    }

    public interface IGroupGiftRepository : IBaseRepository<GroupGift>
    {
    }

    public interface IVehicleRepository : IBaseRepository<Vehicle>
    {
        Task<Vehicle?> GetActiveVehicleAsync(string userId, IDbTransaction? trans = null);
        Task<List<Vehicle>> GetUserVehiclesAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IPetRepository : IBaseRepository<Pet>
    {
        Task<Pet?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
    }

    public interface IPetInventoryRepository : IBaseRepository<PetInventory>
    {
        Task<List<PetInventory>> GetByUserAsync(string userId, IDbTransaction? trans = null);
        Task<PetInventory?> GetItemAsync(string userId, string itemId, IDbTransaction? trans = null);
    }

    public interface IMountRepository : IBaseRepository<Mount>
    {
        Task<Mount?> GetActiveMountAsync(string userId, IDbTransaction? trans = null);
        Task<List<Mount>> GetUserMountsAsync(string userId, IDbTransaction? trans = null);
    }

    public interface ICultivationProfileRepository : IBaseRepository<CultivationProfile>
    {
        Task<CultivationProfile?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
        Task<List<CultivationProfile>> GetTopCultivatorsAsync(int limit = 10, IDbTransaction? trans = null);
    }

    public interface ICultivationRecordRepository : IBaseRepository<CultivationRecord>
    {
    }

    public interface IBuyFriendsRepository : IBaseRepository<BuyFriends>
    {
        Task<long> GetCurrMasterAsync(long groupId, long friendId, IDbTransaction? trans = null);
        Task<long> GetSellPriceAsync(long groupId, long friendId, IDbTransaction? trans = null);
        Task<long> GetBuyPriceAsync(long groupId, long friendId, IDbTransaction? trans = null);
        Task<int> GetBuyIdAsync(long groupId, long friendId, IDbTransaction? trans = null);
        Task<long> GetPetCountAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<List<(long FriendId, long SellPrice)>> GetPriceListAsync(long groupId, int limit, IDbTransaction? trans = null);
        Task<List<(long GroupId, long SellPrice)>> GetMyPriceListAsync(long userId, int limit, IDbTransaction? trans = null);
        Task<int> GetRankAsync(long groupId, long sellPrice, IDbTransaction? trans = null);
        Task<List<(long FriendId, long SellPrice)>> GetMyPetListAsync(long groupId, long userId, int limit, IDbTransaction? trans = null);
    }

    public interface IUserModuleAccessRepository : IBaseRepository<UserModuleAccess>
    {
        Task<List<UserModuleAccess>> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
        Task<UserModuleAccess?> GetAsync(string userId, string moduleId, IDbTransaction? trans = null);
    }

    public interface IUserLevelRepository : IBaseRepository<UserLevel>
    {
        Task<UserLevel?> GetByUserIdAsync(string userId, IDbTransaction? trans = null);
        Task<List<UserLevel>> GetTopRankingsAsync(int limit = 10, IDbTransaction? trans = null);
    }

    public interface ISecretLoveRepository : IBaseRepository<SecretLove>
    {
        Task<string> GetLoveStatusAsync(IDbTransaction? trans = null);
        Task<bool> IsLoveEachotherAsync(long userId, long loveId, IDbTransaction? trans = null);
        Task<int> GetCountLoveAsync(long userId, IDbTransaction? trans = null);
        Task<int> GetCountLoveMeAsync(long userId, IDbTransaction? trans = null);
    }

    public interface IShuffledDeckRepository : IBaseRepository<ShuffledDeck>
    {
        Task ClearShuffledDeckAsync(long groupId, IDbTransaction? trans = null);
        Task ClearShuffledDeckAsync(long groupId, long id, IDbTransaction? trans = null);
        Task ClearShuffledDeckAsync(long groupId, List<int> ids, IDbTransaction? trans = null);
        Task<List<Card>> ReadShuffledDeckAsync(long groupId, IDbTransaction? trans = null, bool lockRow = false);
        Task<bool> IsShuffledDeckExistsAsync(long groupId, IDbTransaction? trans = null);
    }

    public interface IBlockRepository : IBaseRepository<Block>
    {
        Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null);
        Task<long> GetBlockIdAsync(string hash, IDbTransaction? trans = null);
        Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null);
        Task<string> GetValueAsync(string field, long blockId, IDbTransaction? trans = null);
        Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null);
    }

    public interface IBlockTypeRepository : IBaseRepository<BlockType>
    {
        Task<int> GetTypeIdAsync(string typeName, IDbTransaction? trans = null);
        Task<decimal> GetOddsAsync(int typeId, IDbTransaction? trans = null);
    }

    public interface IBlockWinRepository : IBaseRepository<BlockWin>
    {
        Task<bool> IsWinAsync(int typeId, int blockNum, IDbTransaction? trans = null);
    }

    public interface IBlockRandomRepository : IBaseRepository<BlockRandom>
    {
        Task<int> GetRandomNumAsync(IDbTransaction? trans = null);
    }

    [Table("groups")]
    public class GroupInfo
    {
        [ExplicitKey]
        public long Id { get; set; }
        public string Game2048 { get; set; } = string.Empty;
    }
}