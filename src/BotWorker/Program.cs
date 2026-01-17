using Serilog;
using StackExchange.Redis;
using BotWorker.Application.Messaging.Pipeline;
using BotWorker.Common.Config;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Repositories;
using BotWorker.Modules.AI.Services;
using BotWorker.Modules.AI.Skills;
using BotWorker.Modules.AI.Tools;
using BotWorker.Modules.Tools;
using BotWorker.Infrastructure.Communication.OneBot;
using Npgsql;
using Dapper;
using Microsoft.Extensions.Configuration;
using BotWorker.Modules.AI.Models;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Modules.AI.Providers;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;

var builder = WebApplication.CreateBuilder(args);

// Dapper Snake_Case Mapping
DefaultTypeMap.MatchNamesWithUnderscores = true;

// 初始化静态配置
GlobalConfig.Initialize(builder.Configuration);
AppConfig.Initialize(builder.Configuration);

// 配置 Serilog
Log.Logger = new LoggerConfiguration()
    .MinimumLevel.Information()
    .MinimumLevel.Override("Microsoft", Serilog.Events.LogEventLevel.Warning)
    .MinimumLevel.Override("System", Serilog.Events.LogEventLevel.Warning)
    .WriteTo.Console()
    .CreateLogger();
builder.Host.UseSerilog();

// 添加基础服务
builder.Services.AddControllers();
builder.Services.AddHttpClient();

// 添加 CORS
builder.Services.AddCors(options =>
{
    options.AddPolicy("AllowAll", policy =>
    {
        policy.AllowAnyOrigin()
              .AllowAnyMethod()
              .AllowAnyHeader();
    });
});

// 注册 Redis
var redisHost = builder.Configuration["redis:host"] ?? "localhost";
var redisPort = builder.Configuration["redis:port"] ?? "6379";
var redisPassword = builder.Configuration["redis:password"];
// 增加超时设置，减少 ConnectionAborted 错误
var redisConn = $"{redisHost}:{redisPort},abortConnect=false,allowAdmin=true,connectTimeout=10000,syncTimeout=10000,keepAlive=60";
if (!string.IsNullOrEmpty(redisPassword))
{
    redisConn += $",password={redisPassword}";
}
builder.Services.AddSingleton<IConnectionMultiplexer>(sp => ConnectionMultiplexer.Connect(redisConn));
builder.Services.AddSingleton<ICacheService, RedisCacheService>();

builder.Services.AddSingleton<IKnowledgeBaseService, BotWorker.Modules.AI.Plugins.KnowledgeBaseService>(sp => 
{
    var httpClient = sp.GetRequiredService<IHttpClientFactory>().CreateClient();
    httpClient.BaseAddress = new Uri(AppConfig.KbApiUrl ?? "http://localhost:5000");
    return new BotWorker.Modules.AI.Plugins.KnowledgeBaseService(httpClient);
});

// 注册核心业务服务
builder.Services.AddSingleton<IDbInitializer, PostgresDbInitializer>();
builder.Services.AddSingleton<IEventNexus, EventNexus>();
builder.Services.AddSingleton<IToolAuditService, ToolAuditService>();
builder.Services.AddSingleton<SandboxService>();

// 注册 AI 存储层 (Repository) - 为 Go 迁移做准备
builder.Services.AddSingleton<ILLMRepository, PostgresLLMRepository>();
builder.Services.AddSingleton<IAgentRepository, PostgresAgentRepository>();
builder.Services.AddSingleton<ILLMCallLogRepository, PostgresLLMCallLogRepository>();
builder.Services.AddSingleton<IAgentLogRepository, PostgresAgentLogRepository>();
builder.Services.AddSingleton<IAgentSubscriptionRepository, PostgresAgentSubscriptionRepository>();
builder.Services.AddSingleton<IAgentTagRepository, PostgresAgentTagRepository>();
builder.Services.AddSingleton<IKnowledgeFileRepository, PostgresKnowledgeFileRepository>();
builder.Services.AddSingleton<IToolAuditRepository, PostgresToolAuditRepository>();

// 注册模型提供商系统
builder.Services.AddSingleton<IModelProviderFactory, ModelProviderFactory>();
builder.Services.AddSingleton<ModelProviderManager>();
builder.Services.AddSingleton<LLMApp>();

// 注册 Evolution 存储层
builder.Services.AddSingleton<ISkillDefinitionRepository, PostgresSkillDefinitionRepository>();
builder.Services.AddSingleton<IJobDefinitionRepository, PostgresJobDefinitionRepository>();
builder.Services.AddSingleton<IEmployeeInstanceRepository, PostgresEmployeeInstanceRepository>();
builder.Services.AddSingleton<ITaskRecordRepository, PostgresTaskRecordRepository>();
builder.Services.AddSingleton<ITaskStepRepository, PostgresTaskStepRepository>();

// 注册 Billing 存储层
builder.Services.AddSingleton<IWalletRepository, PostgresWalletRepository>();
builder.Services.AddSingleton<ILeaseResourceRepository, PostgresLeaseResourceRepository>();

// 注册 BaseInfo 存储层 (Dapper)
builder.Services.AddSingleton<ILeaseContractRepository, PostgresLeaseContractRepository>();
builder.Services.AddSingleton<IBillingTransactionRepository, PostgresBillingTransactionRepository>();

// 注册 BaseInfo 存储层
builder.Services.AddSingleton<IBotRepository, BotRepository>();
builder.Services.AddSingleton<IChengyuRepository, ChengyuRepository>();
builder.Services.AddSingleton<ICidianRepository, CidianRepository>();
builder.Services.AddSingleton<ICityRepository, CityRepository>();
builder.Services.AddSingleton<IUserRepository, UserRepository>();
builder.Services.AddSingleton<IGroupRepository, GroupRepository>();
builder.Services.AddSingleton<IGroupMemberRepository, GroupMemberRepository>();
builder.Services.AddSingleton<ICoinsLogRepository, CoinsLogRepository>();
builder.Services.AddSingleton<ICreditLogRepository, CreditLogRepository>();
builder.Services.AddSingleton<IBalanceLogRepository, BalanceLogRepository>();
builder.Services.AddSingleton<ITokensLogRepository, TokenLogRepository>();
builder.Services.AddSingleton<IBotLogRepository, BotLogRepository>();
builder.Services.AddSingleton<IBlackListRepository, BlackListRepository>();
builder.Services.AddSingleton<IWhiteListRepository, WhiteListRepository>();
builder.Services.AddSingleton<IGreyListRepository, GreyListRepository>();
builder.Services.AddSingleton<IBugRepository, BugRepository>();
builder.Services.AddSingleton<IBotHintsRepository, BotHintsRepository>();
builder.Services.AddSingleton<ITokenRepository, TokenRepository>();
builder.Services.AddSingleton<IGroupOfficalRepository, GroupOfficalRepository>();
builder.Services.AddSingleton<IGroupEventRepository, GroupEventRepository>();
builder.Services.AddSingleton<IFriendRepository, FriendRepository>();
builder.Services.AddSingleton<IPartnerRepository, PartnerRepository>();
builder.Services.AddSingleton<IPriceRepository, PriceRepository>();
builder.Services.AddSingleton<IWeatherRepository, WeatherRepository>();
builder.Services.AddSingleton<ISchemaRepository, SchemaRepository>();
builder.Services.AddSingleton<IToolService, ToolService>();
builder.Services.AddSingleton<IRmbDaxieService, RmbDaxieService>();
builder.Services.AddSingleton<IPinyinService, PinyinService>();
builder.Services.AddSingleton<IEncryptService, EncryptService>();
builder.Services.AddSingleton<IJielongRepository, JielongRepository>();
builder.Services.AddSingleton<IFishingUserRepository, FishingUserRepository>();
builder.Services.AddSingleton<IFishingBagRepository, FishingBagRepository>();
builder.Services.AddSingleton<IIncomeRepository, IncomeRepository>();
builder.Services.AddSingleton<IGroupVipRepository, GroupVipRepository>();
builder.Services.AddSingleton<IGroupWarnRepository, GroupWarnRepository>();
builder.Services.AddSingleton<IQuestionInfoRepository, QuestionInfoRepository>();
builder.Services.AddSingleton<IBotCmdRepository, BotCmdRepository>();
builder.Services.AddSingleton<ITodoRepository, TodoRepository>();
builder.Services.AddSingleton<IHandleQuestionRepository, HandleQuestionRepository>();
builder.Services.AddSingleton<IGroupPropsRepository, GroupPropsRepository>();
builder.Services.AddSingleton<IPropRepository, PropRepository>();
builder.Services.AddSingleton<IGroupMsgCountRepository, GroupMsgCountRepository>();
builder.Services.AddSingleton<IGreetingRecordsRepository, GreetingRecordsRepository>();
builder.Services.AddSingleton<ISystemSettingRepository, SystemSettingRepository>();
builder.Services.AddSingleton<IGoodsOrderRepository, GoodsOrderRepository>();
builder.Services.AddSingleton<IGoodsTransRepository, GoodsTransRepository>();
builder.Services.AddSingleton<IIDCRepository, IDCRepository>();
builder.Services.AddSingleton<IBotEventLogRepository, BotEventLogRepository>();
builder.Services.AddSingleton<IBotMessageRepository, BotMessageRepository>();

// 注册 Game 存储层
builder.Services.AddSingleton<IUserMarriageRepository, UserMarriageRepository>();
builder.Services.AddSingleton<IMarriageProposalRepository, MarriageProposalRepository>();
builder.Services.AddSingleton<IWeddingItemRepository, WeddingItemRepository>();
builder.Services.AddSingleton<ISweetHeartRepository, SweetHeartRepository>();
builder.Services.AddSingleton<IBabyRepository, BabyRepository>();
builder.Services.AddSingleton<IBabyEventRepository, BabyEventRepository>();
builder.Services.AddSingleton<IBabyConfigRepository, BabyConfigRepository>();
builder.Services.AddSingleton<IRobberyRecordRepository, RobberyRecordRepository>();
builder.Services.AddSingleton<IUserPairingProfileRepository, UserPairingProfileRepository>();
builder.Services.AddSingleton<IPairingRecordRepository, PairingRecordRepository>();
builder.Services.AddSingleton<IBrickRecordRepository, BrickRecordRepository>();
builder.Services.AddSingleton<IGiftRepository, GiftRepository>();
builder.Services.AddSingleton<IGiftStoreItemRepository, GiftStoreItemRepository>();
builder.Services.AddSingleton<IGiftBackpackRepository, GiftBackpackRepository>();
builder.Services.AddSingleton<IGiftLogRepository, GiftLogRepository>();
builder.Services.AddSingleton<IGroupGiftRepository, GroupGiftRepository>();
builder.Services.AddSingleton<IVehicleRepository, VehicleRepository>();
builder.Services.AddSingleton<IDigitalStaffRepository, DigitalStaffRepository>();
builder.Services.AddSingleton<ICognitiveMemoryRepository, CognitiveMemoryRepository>();
builder.Services.AddSingleton<IStaffKpiRepository, StaffKpiRepository>();
builder.Services.AddSingleton<IStaffTaskRepository, StaffTaskRepository>();
builder.Services.AddSingleton<IUserModuleAccessRepository, UserModuleAccessRepository>();
builder.Services.AddSingleton<IUserLevelRepository, UserLevelRepository>();
builder.Services.AddSingleton<IPetRepository, PetRepository>();
builder.Services.AddSingleton<IPetInventoryRepository, PetInventoryRepository>();
builder.Services.AddSingleton<IMountRepository, MountRepository>();
builder.Services.AddSingleton<IBuyFriendsRepository, BuyFriendsRepository>();
builder.Services.AddSingleton<ICultivationProfileRepository, CultivationProfileRepository>();
builder.Services.AddSingleton<ICultivationRecordRepository, CultivationRecordRepository>();
builder.Services.AddSingleton<ISecretLoveRepository, SecretLoveRepository>();
builder.Services.AddSingleton<IShuffledDeckRepository, ShuffledDeckRepository>();
builder.Services.AddSingleton<IBlockRepository, BlockRepository>();
builder.Services.AddSingleton<IBlockTypeRepository, BlockTypeRepository>();
builder.Services.AddSingleton<IBlockWinRepository, BlockWinRepository>();
builder.Services.AddSingleton<IBlockRandomRepository, BlockRandomRepository>();
builder.Services.AddSingleton<IMusicRepository, MusicRepository>();
builder.Services.AddSingleton<ISongOrderRepository, SongOrderRepository>();
builder.Services.AddSingleton<IUserMetricRepository, UserMetricRepository>();
builder.Services.AddSingleton<IUserAchievementRepository, UserAchievementRepository>();

builder.Services.AddSingleton<IMcpService, MCPManager>();
builder.Services.AddSingleton<IRagService, RagService>();
builder.Services.AddSingleton<IAIService, AIService>();
builder.Services.AddSingleton<ICodeRunnerService, CodeRunnerService>();
builder.Services.AddSingleton<IImageGenerationService, ImageGenerationService>();
builder.Services.AddSingleton<IJobService, JobService>();
builder.Services.AddSingleton<IUserService, BotWorker.Application.Services.UserService>();
builder.Services.AddSingleton<BotWorker.Application.Services.IPartnerService, BotWorker.Modules.Office.PartnerService>();
builder.Services.AddSingleton<ISimpleGameService, SimpleGameService>();
builder.Services.AddSingleton<IGame2048Service, Game2048Service>();
builder.Services.AddSingleton<IBlockService, BlockService>();
builder.Services.AddSingleton<IGroupGiftService, GroupGiftService>();
builder.Services.AddSingleton<IGroupMemberService, GroupMemberService>();
builder.Services.AddSingleton<IGoodsTransService, GoodsTransService>();
builder.Services.AddSingleton<IGroupPropsService, GroupPropsService>();
builder.Services.AddSingleton<IGroupWarnService, GroupWarnService>();
builder.Services.AddSingleton<IQuestionInfoService, QuestionInfoService>();
builder.Services.AddSingleton<IGroupMsgCountService, GroupMsgCountService>();
builder.Services.AddSingleton<IGroupService, GroupService>();
builder.Services.AddSingleton<IBotCmdService, BotCmdService>();
builder.Services.AddSingleton<IJielongService, JielongService>();
builder.Services.AddSingleton<IFishingService, FishingService>();
builder.Services.AddSingleton<IRedBlueService, RedBlueService>();
builder.Services.AddSingleton<IMenuService, MenuService>();
builder.Services.AddSingleton<IChengyuService, ChengyuService>();
builder.Services.AddSingleton<IAchievementService, AchievementService>();
builder.Services.AddSingleton<IAgentService, AgentService>();
builder.Services.AddSingleton<PetService>();

// 注册工具/技能系统
builder.Services.AddSingleton<ISkill, FileSkills>();
builder.Services.AddSingleton<ISkill, ShellSkills>();
builder.Services.AddSingleton<ISkill, PlanSkills>();
builder.Services.AddSingleton<ISkill, ReviewSkills>();
builder.Services.AddSingleton<ISkillService, SkillService>();

builder.Services.AddSingleton<IEmployeeService, EmployeeService>();
builder.Services.AddSingleton<IBillingService, BillingService>();
builder.Services.AddSingleton<IEvaluationService, EvaluationService>();
builder.Services.AddSingleton<IEvolutionService, BotWorker.Modules.AI.Services.EvolutionService>();
builder.Services.AddSingleton<IDevWorkflowManager, DevWorkflowManager>();
builder.Services.AddSingleton<ITaskDecompositionService, TaskDecompositionService>();
builder.Services.AddSingleton<IUniversalAgentManager, UniversalAgentManager>();
builder.Services.AddSingleton<IAgentExecutor, AgentExecutor>();
// builder.Services.AddHostedService<BotWorker.Infrastructure.Messaging.RedisStreamConsumer>();
builder.Services.AddHostedService<McpInitializationService>();
builder.Services.AddHostedService<BotWorker.Modules.AI.Services.EvolutionBackgroundService>();
builder.Services.AddSingleton<II18nService, I18nService>();
builder.Services.AddSingleton<IOneBotApiClient, OneBotApiClient>();
builder.Services.AddSingleton<IRobot, PluginManager>();
builder.Services.AddSingleton<PluginManager>(sp => (PluginManager)sp.GetRequiredService<IRobot>());
builder.Services.AddSingleton<IPluginLoaderService, PluginLoaderService>();
builder.Services.AddSingleton<IMCPHost, PluginMcpHost>();

// 注册中间件
builder.Services.AddSingleton<ExceptionMiddleware>();
builder.Services.AddSingleton<FriendlyMessageMiddleware>();
builder.Services.AddTransient<PluginMiddleware>();
builder.Services.AddTransient<BlacklistMiddleware>();
builder.Services.AddTransient<PreProcessMiddleware>();
builder.Services.AddTransient<MaintenanceMiddleware>();
builder.Services.AddTransient<SecretSignalMiddleware>();
builder.Services.AddTransient<StatisticsMiddleware>();
builder.Services.AddTransient<VipMiddleware>();
builder.Services.AddTransient<SetupMiddleware>();
builder.Services.AddTransient<PowerStatusMiddleware>();
builder.Services.AddTransient<MediaTypeMiddleware>();
builder.Services.AddTransient<BuiltinCommandMiddleware>();
builder.Services.AddTransient<QaMiddleware>();
builder.Services.AddSingleton<AiMiddleware>();
builder.Services.AddTransient<AutoSignInMiddleware>();

// 注册并配置 Pipeline
builder.Services.AddSingleton<MessagePipeline>(sp => 
{
    var pipeline = new MessagePipeline(sp);
    // 1. 全局异常处理
    pipeline.Use(sp.GetRequiredService<ExceptionMiddleware>());
    // 2. 最终消息加工 (占位符替换等，作为一个包裹整个管道的后置处理器)
    pipeline.Use(sp.GetRequiredService<FriendlyMessageMiddleware>());
    // 3. 消息清洗与预处理
    pipeline.Use(sp.GetRequiredService<PreProcessMiddleware>());
    // 4. 系统级拦截 (维护中)
    pipeline.Use(sp.GetRequiredService<MaintenanceMiddleware>());
    // 5. 全局暗语 (状态查询)
    pipeline.Use(sp.GetRequiredService<SecretSignalMiddleware>());
    // 6. 安全级拦截 (黑名单)
    pipeline.Use(sp.GetRequiredService<BlacklistMiddleware>());
    // 7. 数据统计 (副作用操作)
    pipeline.Use(sp.GetRequiredService<StatisticsMiddleware>());
    // 8. 权限与VIP检查
    pipeline.Use(sp.GetRequiredService<VipMiddleware>());
    // 9. 管理级指令 (不受开关机限制)
    pipeline.Use(sp.GetRequiredService<SetupMiddleware>());
    // 10. 状态级拦截 (开关机状态检查)
    pipeline.Use(sp.GetRequiredService<PowerStatusMiddleware>());
    // 11. 媒体类型处理 (图片/文件等)
    pipeline.Use(sp.GetRequiredService<MediaTypeMiddleware>());
    // 12. 核心内置指令 (踢人/禁言/普通指令)
    pipeline.Use(sp.GetRequiredService<BuiltinCommandMiddleware>());
    // 13. 问答系统
    pipeline.Use(sp.GetRequiredService<QaMiddleware>());
    // 14. AI/智能体
    pipeline.Use(sp.GetRequiredService<AiMiddleware>());
    // 15. 自动化业务 (自动签到)
    pipeline.Use(sp.GetRequiredService<AutoSignInMiddleware>());
    // 16. 业务插件级分发 (普通插件)
    pipeline.Use(sp.GetRequiredService<PluginMiddleware>());
    return pipeline;
});

// 注册启动时加载插件的任务
builder.Services.AddHostedService<StartupLoader>();
builder.Services.AddHostedService<BotWorker.Infrastructure.Messaging.RedisStreamConsumer>();
//builder.Services.AddHostedService<BotNexusClient>();

var app = builder.Build();

Log.Information("[Startup] Application built. Starting host...");

// 初始化数据库
using (var scope = app.Services.CreateScope())
{
    try {
        var initializer = scope.ServiceProvider.GetRequiredService<IDbInitializer>();
        await initializer.InitializeAsync();
    } catch (Exception ex) {
        Log.Warning("[Startup] Database initialization failed: {Message}. Continuing in offline mode...", ex.Message);
    }
}

// 初始化 SqlHelper 缓存
SqlHelper.CacheService = app.Services.GetRequiredService<ICacheService>();

// 全局静态配置注入 (保留必要的)
GlobalConfig.ServiceProvider = app.Services;
LLMApp.ServiceProvider = app.Services;

// 配置 HTTP 请求管道
if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
}

app.UseCors("AllowAll");
app.UseRouting();
app.UseStaticFiles();
app.MapControllers();
app.Run();

// 简单的启动加载器
public class StartupLoader(IServiceProvider serviceProvider, IPluginLoaderService loaderService, LLMApp llmApp, ILLMRepository llmRepository) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        Log.Information("[Startup] Starting StartupLoader...");
        // 0. 确保内置指令存在
        await BotWorker.Infrastructure.Tools.Todo.EnsureTableCreatedAsync();
        
        var jobService = serviceProvider.GetRequiredService<IJobService>();
        var botCmdService = serviceProvider.GetRequiredService<IBotCmdService>();
        
        // 注入初始岗位
        await jobService.SeedJobsAsync();

        // 注入初始 AI 模型
        await SeedAiModelsAsync();

        // [TEST] 验证动态技能
        try {
            var skillService = serviceProvider.GetRequiredService<ISkillService>();
            var testResult = await skillService.ExecuteSkillAsync("PYTEST", "HelloTarget", "Testing dynamic python skill", new Dictionary<string, string>());
            Log.Information("[TEST] Dynamic Skill Output: \n{Result}", testResult);
        } catch (Exception ex) {
            Log.Error(ex, "[TEST] Dynamic Skill Execution Failed");
        }

        await botCmdService.EnsureCommandExistsAsync("设置Key", "设置Key");
        await botCmdService.EnsureCommandExistsAsync("岗位任务", "岗位任务");
        await botCmdService.EnsureCommandExistsAsync("自动开发", "自动开发");
        await botCmdService.EnsureCommandExistsAsync("开启租赁", "开启租赁");
        await botCmdService.EnsureCommandExistsAsync("关闭租赁", "关闭租赁");
        await botCmdService.EnsureCommandExistsAsync("我的Key", "我的Key");
        await botCmdService.EnsureCommandExistsAsync("积分榜", "积分榜");
        await botCmdService.EnsureCommandExistsAsync("后台", "后台");
        
        Log.Information("[Startup] Initializing AI App...");
        // 1. 初始化 AI 提供商
        await llmApp.InitializeAsync();
        
        Log.Information("[Startup] Loading all plugins...");
        // 2. 加载插件
        await loaderService.LoadAllPluginsAsync();
        Log.Information("[Startup] StartupLoader finished.");
    }

    private async Task SeedAiModelsAsync()
    {
        Log.Information("[Startup] Entering SeedAiModelsAsync...");
        try
        {
            // 尝试修复数据库架构：确保 ai_models 有 type 列
            try
            {
                var connString = GlobalConfig.KnowledgeBaseConnection;
                Log.Information("[Startup] Checking database schema for ai_models using connection string: {ConnString}", 
                    string.IsNullOrEmpty(connString) ? "EMPTY" : "Provided");
                
                if (string.IsNullOrEmpty(connString))
                {
                    Log.Warning("[Startup] KnowledgeBaseConnection is empty, skipping schema fix");
                }
                else
                {
                    using var conn = new NpgsqlConnection(connString);
                    await conn.OpenAsync();
                    var checkColumnSql = "SELECT count(*) FROM information_schema.columns WHERE table_name='ai_models' AND column_name='type'";
                    var count = await conn.ExecuteScalarAsync<long>(checkColumnSql);
                    if (count == 0)
                    {
                        Log.Information("[Startup] Adding missing 'type' column to 'ai_models' table...");
                        await conn.ExecuteAsync("ALTER TABLE ai_models ADD COLUMN type VARCHAR(20) DEFAULT 'chat' NOT NULL");
                    }
                    else
                    {
                        Log.Information("[Startup] 'type' column already exists in 'ai_models' table.");
                    }
                }
            }
            catch (Exception ex)
            {
                Log.Warning(ex, "[Startup] Failed to check/fix database schema for ai_models");
            }

            var providers = (await llmRepository.GetActiveProvidersAsync()).ToList();
            Log.Information("[Startup] Found {Count} active providers: {Names}", providers.Count, string.Join(", ", providers.Select(p => p.Name)));
            
            var doubao = providers.FirstOrDefault(p => p.Name.Equals("Doubao", StringComparison.OrdinalIgnoreCase));
            if (doubao != null)
            {
                var models = (await llmRepository.GetModelsByProviderIdAsync(doubao.Id)).ToList();
                Log.Information("[Startup] Found {Count} models for Doubao provider. Active count: {ActiveCount}", 
                    models.Count, models.Count(m => m.IsActive));
                
                // 无论是否有激活的模型，都尝试激活或添加我们需要的核心模型
                var defaultModels = new[] 
                { 
                    "doubao-1-5-pro-32k-250115", 
                    "doubao-embedding-v2", 
                    "doubao-seed-1-8-251228" 
                };
                foreach (var modelName in defaultModels)
                {
                    // 尝试匹配，优先全名匹配
                    var existing = models.FirstOrDefault(m => m.Name.Equals(modelName, StringComparison.OrdinalIgnoreCase))
                                ?? models.FirstOrDefault(m => m.Name.StartsWith(modelName.Split('-')[0], StringComparison.OrdinalIgnoreCase));
                    
                    if (existing != null)
                    {
                        if (!existing.IsActive)
                        {
                            existing.IsActive = true;
                            await llmRepository.UpdateModelAsync(existing);
                            Log.Information("[Startup] Activated existing Doubao model: {ModelName}", existing.Name);
                        }
                    }
                    else
                    {
                        await llmRepository.AddModelAsync(new LLMModel 
                        { 
                            ProviderId = doubao.Id, 
                            Name = modelName, 
                            Type = modelName.Contains("embedding") ? "embedding" : "chat", 
                            IsActive = true 
                        });
                        Log.Information("[Startup] Added and activated new default Doubao model: {ModelName}", modelName);
                    }
                }
            }
            else
            {
                Log.Warning("[Startup] Doubao provider not found in active providers");
            }
        }
        catch (Exception ex)
        {
            Log.Error(ex, "[Startup] Failed to seed AI models");
        }
    }
}
