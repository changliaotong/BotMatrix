using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IUserRepository : IBaseRepository<UserInfo>
    {
        Task<UserInfo?> GetByOpenIdAsync(string openId, long botUin);
        Task<UserInfo?> GetBySz84UidAsync(int sz84Uid);
        Task<long> AddAsync(UserInfo user);
        Task<bool> UpdateAsync(UserInfo user);
        Task<long> GetCreditAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<bool> AddCreditAsync(long botUin, long groupId, long qq, long amount, string reason, IDbTransaction? trans = null);
        Task<long> GetCreditForUpdateAsync(long botUin, long groupId, long qq, IDbTransaction? trans = null);
        Task<long> GetTokensAsync(long qq, IDbTransaction? trans = null);
        Task<long> GetTokensRankingAsync(long groupId, long qq);
        Task<long> GetDayTokensGroupAsync(long groupId, long userId);
        Task<long> GetDayTokensAsync(long userId);
        Task<bool> AddTokensAsync(long qq, long amount, IDbTransaction? trans = null);
        Task<long> GetTokensForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<string> GetTokensListAsync(long groupId, int top);
        Task<decimal> GetBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<bool> AddBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceAsync(long qq, IDbTransaction? trans = null);
        Task<decimal> GetFreezeBalanceForUpdateAsync(long qq, IDbTransaction? trans = null);
        Task<bool> FreezeBalanceAsync(long qq, decimal amount, IDbTransaction? trans = null);
        Task<string> GetBalanceListAsync(long groupId, long qq);
        Task<string> GetRankAsync(long groupId, long qq);
        Task<bool> GetIsBlackAsync(long qq);
        Task<bool> GetIsFreezeAsync(long qq);
        Task<bool> GetIsShutupAsync(long qq);
        Task<bool> GetIsSuperAsync(long qq);
        Task<bool> UpdateCszGameAsync(long qq, int cszRes, long cszCredit, int cszTimes);
        Task<long> GetCoinsAsync(long userId);
        Task<long> GetBotUinByOpenidAsync(string userOpenid);
        Task<long> GetTargetUserIdAsync(string userOpenid);
        Task<long> GetMaxIdInRangeAsync(long min, long max);
        Task<string> GetUserOpenidAsync(long botUin, long userId);
        Task<long> GetSourceQQAsync(long botUin, long userId);
        Task SyncCacheFieldAsync(long userId, string field, object value);
        Task<string> GetCreditRankingAsync(long groupId, int top, string format);
    }

    public interface IPartnerRepository : IBaseRepository<BotWorker.Modules.Office.Partner>
    {
        Task<bool> IsPartnerAsync(long userId);
        Task<bool> IsNotPartnerAsync(long userId);
        Task<int> AppendAsync(long userId, long refUserId = 0);
        Task<long> GetUnsettledCreditAsync(long userId);
        Task<int> SettleAsync(long userId, IDbTransaction? trans = null);
        Task<string> GetCreditListAsync(long userId);
        Task<string> GetCreditTodayAsync(long userId);
    }

    public interface IPriceRepository : IBaseRepository<BotWorker.Modules.Office.Price>
    {
        Task<decimal> GetRobotPriceAsync(long month);
    }

    public interface IWeatherRepository : IBaseRepository<BotWorker.Domain.Entities.Weather>
    {
        Task<string?> GetRecentWeatherAsync(string cityName, int hours);
    }

    public interface ISchemaRepository
    {
        Task<string> GetFieldCaptionAsync(string tableId, string fieldName);
        Task<string> GetReportFieldCaptionAsync(string rptId, string fieldName);
        Task<BotWorker.Domain.Constants.AppsInfo.TableInfo?> GetTableInfoAsync(int tableId);
    }

    public interface IToolService
    {
        Task<Dictionary<string, string>> GetWeatherAsync(IEnumerable<string> cities);
        Task<string> GetCountDownAsync();
        string GetMonthRes(DateTime dt, bool isYinli = false, int spaceCount = 3, int spaceCount2 = 1);
        Task<string> GetTranslateAsync(string text);
        Task<string> CalculateAsync(string expression);
        Task<string> GetCidResAsync(BotWorker.Domain.Models.BotMessages.BotMessage msg, bool isMinus = true);
        Task<string> GetAreaNameAsync(string areaCode);
    }

    public interface IPropRepository : IBaseRepository<BotWorker.Domain.Entities.Prop>
    {
        Task<int> GetPropPriceAsync(long propId);
        Task<long> GetIdAsync(string name);
        Task<string> GetPropListAsync();
    }

    public interface IGroupPropsRepository : IBaseRepository<BotWorker.Domain.Entities.GroupProps>
    {
        Task<long> GetIdAsync(long groupId, long qq, long propId);
        Task<bool> HavePropAsync(long groupId, long userId, long propId);
        Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp);
        Task<string> GetMyPropListAsync(long groupId, long userId);
        Task<int> InsertAsync(long groupId, long qq, long propId, IDbTransaction? trans = null);
    }

    public interface IGroupRepository : IBaseRepository<GroupInfo>
    {
        Task<GroupInfo?> GetByOpenIdAsync(string openId, long botUin);
        Task<long> AddAsync(GroupInfo group);
        Task<bool> UpdateAsync(GroupInfo group);
        Task<long> GetGroupOwnerAsync(long groupId, long def = 0, IDbTransaction? trans = null);
        Task<bool> GetIsCreditAsync(long groupId);
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
        Task<int> AppendAsync(long group, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "");
        Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0);
        Task<bool> GetIsNoLogAsync(long groupId);
        Task<bool> GetIsNoCheckAsync(long groupId);
        Task<bool> GetIsHintCloseAsync(long groupId);
        Task<long> GetSourceGroupIdAsync(long groupId);
        Task<long> GetSourceGroupIdAsync(long botUin, long groupId);
        Task<int> SetIsOpenAsync(bool isOpen, long groupId);
        Task<int> SetPowerOnAsync(bool isOpen, long groupId);
        Task<bool> GetPowerOnAsync(long groupId);
        Task<string> GetSystemPromptStatusAsync(long groupId);
        Task<string> GetVipResAsync(long groupId);
        Task<int> StartCyGameAsync(long groupId);
        Task<T> GetValueAsync<T>(string field, long groupId, IDbTransaction? trans = null);
        Task<int> GetIntAsync(string field, long groupId, IDbTransaction? trans = null);
        Task<int> SetValueAsync(string field, object value, long groupId, IDbTransaction? trans = null);
    }

    public interface IGroupVipRepository : IBaseRepository<GroupVip>
    {
        Task<int> BuyRobotAsync(long botUin, long groupId, string groupName, long qqBuyer, string buyerName, long month, decimal payMoney, string payMethod, string trade, string memo, int insertBy, IDbTransaction? trans = null);
        Task<int> ChangeGroupAsync(long groupId, long newGroupId, long qq, IDbTransaction? trans = null);
        Task<int> RestDaysAsync(long groupId);
        Task<int> RestMonthsAsync(long groupId);
        Task<bool> IsYearVIPAsync(long groupId);
        Task<bool> IsVipAsync(long groupId);
        Task<bool> IsForeverAsync(long groupId);
        Task<bool> IsVipOnceAsync(long groupId);
        Task<bool> IsClientVipAsync(long qq);
        Task<string> GetVipListByUserIdAsync(long userId);
    }

    public interface IGroupMemberRepository : IBaseRepository<GroupMember>
    {
        Task<GroupMember?> GetAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> ExistsAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<long> AddAsync(GroupMember member);
        Task<bool> UpdateAsync(GroupMember member);
        Task<long> GetCoinsAsync(int coinsType, long groupId, long userId, IDbTransaction? trans = null);
        Task<bool> AddCoinsAsync(long botUin, long groupId, long userId, int coinsType, long amount, string reason);
        Task<long> GetCoinsForUpdateAsync(int coinsType, long groupId, long userId, IDbTransaction trans);
        Task<bool> UpdateSignInfoAsync(long groupId, long userId, int signTimes, int signLevel, IDbTransaction? trans = null);
        Task<int> GetSignDateDiffAsync(long groupId, long userId);
        Task<string> GetSignListAsync(long groupId, int topN = 10);
        Task<long> GetLongAsync(string field, long groupId, long userId, IDbTransaction? trans = null);
        Task<T> GetValueAsync<T>(string field, long groupId, long userId, IDbTransaction? trans = null);
        Task<int> SetValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null);
        Task<int> IncrementValueAsync(string field, object value, long groupId, long userId, IDbTransaction? trans = null);
        Task<long> GetForUpdateAsync(string field, long groupId, long userId, IDbTransaction trans);
        Task<int> AppendAsync(long groupId, long userId, string name, string displayName = "", long groupCredit = 0, string confirmCode = "", IDbTransaction? trans = null);
        Task<long> GetCoinsRankingAsync(long groupId, long userId);
        Task<long> GetCoinsRankingAllAsync(long userId);
        Task<int> GetIntAsync(string field, long groupId, long userId, IDbTransaction? trans = null);
        Task<string> GetCreditRankingAsync(long groupId, int top, string format);
    }

    public interface IJielongRepository : IBaseRepository<BotWorker.Modules.Games.Jielong>
    {
        Task<string?> GetRandomChengyuAsync();
        Task<string?> GetChengYuByPinyinAsync(string pinyin, long groupId);
        Task<bool> IsDupAsync(long groupId, long userId, string chengYu);
        Task<int> GetMaxIdAsync(long groupId);
        Task<int> GetCountAsync(long groupId, long userId);
        Task<long> GetCreditAddAsync(long userId);
        Task<bool> InGameAsync(long groupId, long userId);
        Task<string> GetCurrentChengYuAsync(long groupId, long userId);
        Task<int> AppendAsync(long groupId, long userId, string userName, string chengYu, int gameNo);
        Task<int> EndGameAsync(long groupId, long userId);
    }

    public interface IBlackListRepository : IBaseRepository<BlackList>
    {
        Task<IEnumerable<long>> GetSystemBlackListAsync();
        Task<bool> IsExistsAsync(long groupId, long userId);
        Task<int> AddAsync(BlackList blackList);
        Task<int> AddBlackListAsync(long botUin, long groupId, string groupName, long qq, string name, long blackQQ, string blackInfo);
        Task<bool> IsSystemBlackAsync(long userId);
        Task<int> DeleteAsync(long groupId, long userId);
        Task<int> ClearGroupAsync(long groupId);
    }

    public interface IWhiteListRepository : IBaseRepository<WhiteList>
    {
        Task<bool> IsExistsAsync(long groupId, long userId);
        Task<int> AddAsync(WhiteList whiteList);
        Task<int> DeleteAsync(long groupId, long userId);
    }

    public interface IGreyListRepository : IBaseRepository<GreyList>
    {
        Task<IEnumerable<long>> GetSystemGreyListAsync();
        Task<bool> IsExistsAsync(long groupId, long userId);
        Task<int> AddAsync(GreyList greyList);
        Task<int> DeleteAsync(long groupId, long userId);
    }

    public interface IBugRepository : IBaseRepository<Bug>
    {
        Task<int> AddAsync(Bug bug);
    }

    public interface IBotHintsRepository : IBaseRepository<BotHints>
    {
        Task<string> GetHintAsync(string cmd);
    }

    public interface ITokenRepository : IBaseRepository<Token>
    {
        Task<Token?> GetByUserIdAsync(long userId);
        Task<string> GetTokenByUserIdAsync(long userId);
        Task<bool> ExistsTokenAsync(string token);
        Task<bool> ExistsTokenAsync(long userId, string token);
        Task<int> UpsertTokenAsync(long userId, string token);
        Task<int> UpsertRefreshTokenAsync(long userId, string token, string refreshToken, DateTime expiryDate);
        Task<string> GetRefreshTokenAsync(long userId);
        Task<bool> IsTokenValidAsync(long userId, string token, long seconds);
    }

    public interface IGroupOfficalRepository : IBaseRepository<GroupOffical>
    {
        Task<bool> IsOfficalAsync(long groupId);
        Task<long> GetTargetGroupAsync(string groupOpenid);
        Task<long> GetMaxGroupIdAsync();
        Task<string> GetGroupOpenidAsync(long botUin, long groupId);
    }

    public interface IGroupEventRepository : IBaseRepository<GroupEvent>
    {
        Task<int> AddAsync(GroupEvent groupEvent);
    }

    public interface IFriendRepository : IBaseRepository<Friend>
    {
        Task<int> AddAsync(Friend friend);
        Task<int> AppendAsync(long botUin, long friendId, string friendName);
        Task<string> GetCreditRankingAsync(long botUin, long groupId, int top, string format);
        Task<bool> UpdateCreditAsync(long botUin, long friendId, long credit);
        Task<bool> UpdateSaveCreditAsync(long botUin, long friendId, long saveCredit);
    }

    public interface IFishingUserRepository : IBaseRepository<BotWorker.Modules.Games.FishingUser>
    {
        Task UpdateStateAsync(long userId, int state, int waitMinutes);
        Task UpdateStateAsync(long userId, int state);
        Task AddExpAndResetStateAsync(long userId, int exp);
        Task UpgradeRodAsync(long userId, long cost);
        Task SellFishAsync(long userId, long totalGold);
    }

    public interface IFishingBagRepository : IBaseRepository<BotWorker.Modules.Games.FishingBag>
    {
        Task<IEnumerable<BotWorker.Modules.Games.FishingBag>> GetByUserIdAsync(long userId, int limit);
        Task<IEnumerable<BotWorker.Modules.Games.FishingBag>> GetAllByUserIdAsync(long userId);
    }

    public interface ICreditLogRepository : IBaseRepository<CreditLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long creditAdd, long creditValue, string creditInfo, IDbTransaction? trans = null);
        Task<int> CreditCountAsync(long userId, string creditInfo, int second = 60);
    }

    public interface IBalanceLogRepository : IBaseRepository<BalanceLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, decimal balanceAdd, decimal balanceValue, string balanceInfo, IDbTransaction? trans = null);
    }

    public interface ITokensLogRepository : IBaseRepository<TokensLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null);
        Task<long> GetDayTokensGroupAsync(long groupId, long userId);
        Task<long> GetDayTokensAsync(long userId);
    }

    public interface IIncomeRepository : IBaseRepository<BotWorker.Modules.Office.Income>
    {
        Task<long> AddAsync(BotWorker.Modules.Office.Income income, IDbTransaction? trans = null);
        Task<float> GetTotalAsync(long userId);
        Task<float> GetTotalLastYearAsync(long userId);
        Task<bool> IsVipOnceAsync(long groupId);
        Task<int> GetClientLevelAsync(long userId);
        Task<string> GetLevelListAsync(long groupId);
        Task<string> GetLeverOrderAsync(long groupId, long userId);
        Task<string> GetStatAsync(string range);
    }

    public interface IBotRepository : IBaseRepository<BotInfo>
    {
        Task<BotInfo?> GetByBotUinAsync(long botUin);
        Task<bool> IsRobotAsync(long userId);
    }

    public interface IBotPublicRepository : IBaseRepository<BotPublic>
    {
        Task<BotPublic?> GetByPublicKeyAsync(string publicKey);
        Task<long> GetRobotQQAsync(string botKey);
        Task<long> GetGroupIdAsync(string botKey);
        Task<string> GetBotNameAsync(string botKey);
    }

    public interface IPublicUserRepository : IBaseRepository<PublicUser>
    {
        Task<long> GetUserIdAsync(string botKey, string userKey);
        Task<bool> IsBindAsync(long userId);
        Task<bool> IsSubscribedToOfficialPublicAsync(long userId);
        Task<string> GetBindTokenAsync(string botKey, string userKey);
        Task<string> GetInviteCodeAsync(string botKey, string userKey);
        Task<string> GetRecResAsync(long botUin, long groupId, string groupName, long userId, string name, string botKey, string clientKey, string message);
        Task<string> GetBindTokenResAsync(BotMessage bm, string tokenType, string bindToken);
    }

    public interface IFaceRepository
    {
        string ConvertFacesBack(string text);
    }
}
