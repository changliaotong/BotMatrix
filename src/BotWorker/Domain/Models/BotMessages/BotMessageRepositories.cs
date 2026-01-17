using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Domain.Interfaces;
using BotWorker.Modules.AI.Interfaces;
using System.Data;
using Microsoft.Extensions.DependencyInjection;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence;
using BotWorker.Infrastructure.Persistence.Repositories;
using Dapper;

using BotWorker.Infrastructure.Tools;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage
    {
        // 通过 ServiceProvider 获取仓储，消除静态引用
        public IUserRepository UserRepository => ServiceProvider?.GetRequiredService<IUserRepository>() ?? throw new InvalidOperationException("IUserRepository not registered or ServiceProvider not set");
        public IGroupRepository GroupRepository => ServiceProvider?.GetRequiredService<IGroupRepository>() ?? throw new InvalidOperationException("IGroupRepository not registered or ServiceProvider not set");
        public IGroupMemberRepository GroupMemberRepository => ServiceProvider?.GetRequiredService<IGroupMemberRepository>() ?? throw new InvalidOperationException("IGroupMemberRepository not registered or ServiceProvider not set");
        public ISignInRepository SignInRepository => ServiceProvider?.GetRequiredService<ISignInRepository>() ?? throw new InvalidOperationException("ISignInRepository not registered or ServiceProvider not set");
        public IBotRepository BotRepository => ServiceProvider?.GetRequiredService<IBotRepository>() ?? throw new InvalidOperationException("IBotRepository not registered or ServiceProvider not set");
        public IGroupService GroupService => ServiceProvider?.GetRequiredService<IGroupService>() ?? throw new InvalidOperationException("IGroupService not registered or ServiceProvider not set");
        public IUserService UserService => ServiceProvider?.GetRequiredService<IUserService>() ?? throw new InvalidOperationException("IUserService not registered or ServiceProvider not set");
        public IBotCmdService BotCmdService => ServiceProvider?.GetRequiredService<IBotCmdService>() ?? throw new InvalidOperationException("IBotCmdService not registered or ServiceProvider not set");
        public ITokensLogRepository TokenLogRepository => ServiceProvider?.GetRequiredService<ITokensLogRepository>() ?? throw new InvalidOperationException("ITokensLogRepository not registered or ServiceProvider not set");
        public ICreditLogRepository CreditLogRepository => ServiceProvider?.GetRequiredService<ICreditLogRepository>() ?? throw new InvalidOperationException("ICreditLogRepository not registered or ServiceProvider not set");
        public IBotMessageRepository BotMessageRepository => ServiceProvider?.GetRequiredService<IBotMessageRepository>() ?? throw new InvalidOperationException("IBotMessageRepository not registered or ServiceProvider not set");
        public IGroupWarnRepository GroupWarnRepository => ServiceProvider?.GetRequiredService<IGroupWarnRepository>() ?? throw new InvalidOperationException("IGroupWarnRepository not registered or ServiceProvider not set");
        public IAnswerRepository AnswerRepository => ServiceProvider?.GetRequiredService<IAnswerRepository>() ?? throw new InvalidOperationException("IAnswerRepository not registered or ServiceProvider not set");
        public IQuestionInfoRepository QuestionInfoRepository => ServiceProvider?.GetRequiredService<IQuestionInfoRepository>() ?? throw new InvalidOperationException("IQuestionInfoRepository not registered or ServiceProvider not set");
        public ISystemSettingRepository SystemSettingRepository => ServiceProvider?.GetRequiredService<ISystemSettingRepository>() ?? throw new InvalidOperationException("ISystemSettingRepository not registered or ServiceProvider not set");
        public IKnowledgeHistoryRepository KnowledgeHistoryRepository => ServiceProvider?.GetRequiredService<IKnowledgeHistoryRepository>() ?? throw new InvalidOperationException("IKnowledgeHistoryRepository not registered or ServiceProvider not set");
        public IGuildEventRepository GuildEventRepository => ServiceProvider?.GetRequiredService<IGuildEventRepository>() ?? throw new InvalidOperationException("IGuildEventRepository not registered or ServiceProvider not set");
        public IGroupVipRepository GroupVipRepository => ServiceProvider?.GetRequiredService<IGroupVipRepository>() ?? throw new InvalidOperationException("IGroupVipRepository not registered or ServiceProvider not set");
        public IGroupSendMessageRepository GroupSendMessageRepository => ServiceProvider?.GetRequiredService<IGroupSendMessageRepository>() ?? throw new InvalidOperationException("IGroupSendMessageRepository not registered or ServiceProvider not set");
        public IBlackListRepository BlackListRepository => ServiceProvider?.GetRequiredService<IBlackListRepository>() ?? throw new InvalidOperationException("IBlackListRepository not registered or ServiceProvider not set");
        public IWhiteListRepository WhiteListRepository => ServiceProvider?.GetRequiredService<IWhiteListRepository>() ?? throw new InvalidOperationException("IWhiteListRepository not registered or ServiceProvider not set");
        public IGroupEventRepository GroupEventRepository => ServiceProvider?.GetRequiredService<IGroupEventRepository>() ?? throw new InvalidOperationException("IGroupEventRepository not registered or ServiceProvider not set");
        public IGreyListRepository GreyListRepository => ServiceProvider?.GetRequiredService<IGreyListRepository>() ?? throw new InvalidOperationException("IGreyListRepository not registered or ServiceProvider not set");
        public IBotEventLogRepository BotEventLogRepository => ServiceProvider?.GetRequiredService<IBotEventLogRepository>() ?? throw new InvalidOperationException("IBotEventLogRepository not registered or ServiceProvider not set");
        public IBotPublicRepository BotPublicRepository => ServiceProvider?.GetRequiredService<IBotPublicRepository>() ?? throw new InvalidOperationException("IBotPublicRepository not registered or ServiceProvider not set");
        public IPublicUserRepository PublicUserRepository => ServiceProvider?.GetRequiredService<IPublicUserRepository>() ?? throw new InvalidOperationException("IPublicUserRepository not registered or ServiceProvider not set");
        public IFaceRepository FaceRepository => ServiceProvider?.GetRequiredService<IFaceRepository>() ?? throw new InvalidOperationException("IFaceRepository not registered or ServiceProvider not set");
        public IMusicRepository MusicRepository => ServiceProvider?.GetRequiredService<IMusicRepository>() ?? throw new InvalidOperationException("IMusicRepository not registered or ServiceProvider not set");
        public BotWorker.Modules.Games.IPetService PetService => ServiceProvider?.GetRequiredService<BotWorker.Modules.Games.IPetService>() ?? throw new InvalidOperationException("IPetService not registered or ServiceProvider not set");
        public IPartnerService PartnerService => ServiceProvider?.GetRequiredService<IPartnerService>() ?? throw new InvalidOperationException("IPartnerService not registered or ServiceProvider not set");
        public IPartnerRepository PartnerRepository => ServiceProvider?.GetRequiredService<IPartnerRepository>() ?? throw new InvalidOperationException("IPartnerRepository not registered or ServiceProvider not set");
        public IPriceRepository PriceRepository => ServiceProvider?.GetRequiredService<IPriceRepository>() ?? throw new InvalidOperationException("IPriceRepository not registered or ServiceProvider not set");
        public IPropRepository PropRepository => ServiceProvider?.GetRequiredService<IPropRepository>() ?? throw new InvalidOperationException("IPropRepository not registered or ServiceProvider not set");
        public IGroupPropsRepository GroupPropsRepository => ServiceProvider?.GetRequiredService<IGroupPropsRepository>() ?? throw new InvalidOperationException("IGroupPropsRepository not registered or ServiceProvider not set");
        public IGroupMemberService GroupMemberService => ServiceProvider?.GetRequiredService<IGroupMemberService>() ?? throw new InvalidOperationException("IGroupMemberService not registered or ServiceProvider not set");
        public IGoodsTransService GoodsTransService => ServiceProvider?.GetRequiredService<IGoodsTransService>() ?? throw new InvalidOperationException("IGoodsTransService not registered or ServiceProvider not set");
        public IGroupPropsService GroupPropsService => ServiceProvider?.GetRequiredService<IGroupPropsService>() ?? throw new InvalidOperationException("IGroupPropsService not registered or ServiceProvider not set");
        public IBlockService BlockService => ServiceProvider?.GetRequiredService<IBlockService>() ?? throw new InvalidOperationException("IBlockService not registered or ServiceProvider not set");
        public ITodoService TodoService => ServiceProvider?.GetRequiredService<ITodoService>() ?? throw new InvalidOperationException("ITodoService not registered or ServiceProvider not set");
        public IJielongService JielongService => ServiceProvider?.GetRequiredService<IJielongService>() ?? throw new InvalidOperationException("IJielongService not registered or ServiceProvider not set");
        public IChengyuService ChengyuService => ServiceProvider?.GetRequiredService<IChengyuService>() ?? throw new InvalidOperationException("IChengyuService not registered or ServiceProvider not set");
        public IFishingService FishingService => ServiceProvider?.GetRequiredService<IFishingService>() ?? throw new InvalidOperationException("IFishingService not registered or ServiceProvider not set");
        public IDevWorkflowManager DevWorkflowManager => ServiceProvider?.GetRequiredService<IDevWorkflowManager>() ?? throw new InvalidOperationException("IDevWorkflowManager not registered or ServiceProvider not set");
        public IUserCreditService UserCreditService => ServiceProvider?.GetRequiredService<IUserCreditService>() ?? throw new InvalidOperationException("IUserCreditService not registered or ServiceProvider not set");
        public IIncomeRepository IncomeRepository => ServiceProvider?.GetRequiredService<IIncomeRepository>() ?? throw new InvalidOperationException("IIncomeRepository not registered or ServiceProvider not set");
        public IGroupOfficalRepository GroupOfficalRepository => ServiceProvider?.GetRequiredService<IGroupOfficalRepository>() ?? throw new InvalidOperationException("IGroupOfficalRepository not registered or ServiceProvider not set");
        public IFriendRepository FriendRepository => ServiceProvider?.GetRequiredService<IFriendRepository>() ?? throw new InvalidOperationException("IFriendRepository not registered or ServiceProvider not set");
        public IToolService ToolService => ServiceProvider?.GetRequiredService<IToolService>() ?? throw new InvalidOperationException("IToolService not registered or ServiceProvider not set");
        public ISchemaRepository SchemaRepository => ServiceProvider?.GetRequiredService<ISchemaRepository>() ?? throw new InvalidOperationException("ISchemaRepository not registered or ServiceProvider not set");
        public IWeatherRepository WeatherRepository => ServiceProvider?.GetRequiredService<IWeatherRepository>() ?? throw new InvalidOperationException("IWeatherRepository not registered or ServiceProvider not set");
        public IBotLogRepository BotLogRepository => ServiceProvider?.GetRequiredService<IBotLogRepository>() ?? throw new InvalidOperationException("IBotLogRepository not registered or ServiceProvider not set");
        public IAgentRepository AgentRepository => ServiceProvider?.GetRequiredService<IAgentRepository>() ?? throw new InvalidOperationException("IAgentRepository not registered or ServiceProvider not set");
        public IAgentSubscriptionRepository AgentSubscriptionRepository => ServiceProvider?.GetRequiredService<IAgentSubscriptionRepository>() ?? throw new InvalidOperationException("IAgentSubscriptionRepository not registered or ServiceProvider not set");
        public IAgentLogRepository AgentLogRepository => ServiceProvider?.GetRequiredService<IAgentLogRepository>() ?? throw new InvalidOperationException("IAgentLogRepository not registered or ServiceProvider not set");
        public IAgentTagRepository AgentTagRepository => ServiceProvider?.GetRequiredService<IAgentTagRepository>() ?? throw new InvalidOperationException("IAgentTagRepository not registered or ServiceProvider not set");
        public ILLMRepository LLMRepository => ServiceProvider?.GetRequiredService<ILLMRepository>() ?? throw new InvalidOperationException("ILLMRepository not registered or ServiceProvider not set");
        public IIDCRepository IDCRepository => ServiceProvider?.GetRequiredService<IIDCRepository>() ?? throw new InvalidOperationException("IIDCRepository not registered or ServiceProvider not set");
        public ICityRepository CityRepository => ServiceProvider?.GetRequiredService<ICityRepository>() ?? throw new InvalidOperationException("ICityRepository not registered or ServiceProvider not set");
        public ICidianRepository CidianRepository => ServiceProvider?.GetRequiredService<ICidianRepository>() ?? throw new InvalidOperationException("ICidianRepository not registered or ServiceProvider not set");
        public IGroupMsgCountService GroupMsgCountService => ServiceProvider?.GetRequiredService<IGroupMsgCountService>() ?? throw new InvalidOperationException("IGroupMsgCountService not registered or ServiceProvider not set");
        public IGroupWarnService GroupWarnService => ServiceProvider?.GetRequiredService<IGroupWarnService>() ?? throw new InvalidOperationException("IGroupWarnService not registered or ServiceProvider not set");
        public IQuestionInfoService QuestionInfoService => ServiceProvider?.GetRequiredService<IQuestionInfoService>() ?? throw new InvalidOperationException("IQuestionInfoService not registered or ServiceProvider not set");
        public IBotService BotService => ServiceProvider?.GetRequiredService<IBotService>() ?? throw new InvalidOperationException("IBotService not registered or ServiceProvider not set");
        public IRmbDaxieService RmbDaxieService => ServiceProvider?.GetRequiredService<IRmbDaxieService>() ?? throw new InvalidOperationException("IRmbDaxieService not registered or ServiceProvider not set");
        public IPinyinService PinyinService => ServiceProvider?.GetRequiredService<IPinyinService>() ?? throw new InvalidOperationException("IPinyinService not registered or ServiceProvider not set");
        public IEncryptService EncryptService => ServiceProvider?.GetRequiredService<IEncryptService>() ?? throw new InvalidOperationException("IEncryptService not registered or ServiceProvider not set");

        // 兼容旧代码的辅助方法
        public static async Task<SqlHelper.TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
            => await SqlHelper.BeginTransactionAsync(existingTrans, level);

        public static async Task<int> ExecAsync(string sql, params object?[] args)
        {
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = SqlHelper.ResolveSql(sql, actualArgs);
            
            using var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            if (conn is System.Data.Common.DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            
            var dapperParams = new DynamicParameters();
            foreach (var p in parameters) dapperParams.Add(p.ParameterName, p.Value);
            if (explicitParams != null)
                foreach (var p in explicitParams) dapperParams.Add(p.ParameterName, p.Value);

            return await conn.ExecuteAsync(resolvedSql, dapperParams, trans);
        }

        private static (IDbTransaction? trans, object?[] actualArgs, IDataParameter[]? parameters) ParseArgs(object?[] args)
        {
            IDbTransaction? trans = null;
            object?[] actualArgs = args;
            IDataParameter[]? parameters = null;

            int start = 0;
            if (args.Length > 0 && args[0] is IDbTransaction t)
            {
                trans = t;
                start = 1;
            }

            if (args.Length > start && args[^1] is IDataParameter[] p)
            {
                parameters = p;
                actualArgs = args[start..^1];
            }
            else
            {
                actualArgs = args[start..];
            }

            return (trans, actualArgs, parameters);
        }

        public static string RetryMsg => "⚠️ 操作失败，请稍后再试";
        public static string CreditSystemClosed => "⚠️ 积分系统未开启";
    }
}
