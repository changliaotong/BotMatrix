using Serilog;
using StackExchange.Redis;
using BotWorker.Application.Messaging.Pipeline;
using BotWorker.Common.Config;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Tools;
using BotWorker.Infrastructure.Communication.OneBot;

var builder = WebApplication.CreateBuilder(args);

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

// 注册 Redis
var redisHost = builder.Configuration["redis:host"] ?? "localhost";
var redisPort = builder.Configuration["redis:port"] ?? "6379";
var redisPassword = builder.Configuration["redis:password"];
var redisConn = $"{redisHost}:{redisPort},abortConnect=false,allowAdmin=true";
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
builder.Services.AddSingleton<IEventNexus, EventNexus>();
builder.Services.AddSingleton<IToolAuditService, ToolAuditService>();
builder.Services.AddSingleton<SandboxService>();
builder.Services.AddSingleton<LLMApp>();
builder.Services.AddSingleton<IMcpService, MCPManager>();
builder.Services.AddSingleton<IRagService, RagService>();
builder.Services.AddSingleton<IAIService, AIService>();
builder.Services.AddSingleton<IImageGenerationService, ImageGenerationService>();
builder.Services.AddSingleton<IJobService, JobService>();
builder.Services.AddSingleton<IEmployeeService, EmployeeService>();
builder.Services.AddSingleton<IEvaluationService, EvaluationService>();
builder.Services.AddSingleton<IEvolutionService, BotWorker.Modules.AI.Services.EvolutionService>();
builder.Services.AddSingleton<IDevWorkflowManager, DevWorkflowManager>();
builder.Services.AddSingleton<IAgentExecutor, AgentExecutor>();
builder.Services.AddHostedService<McpInitializationService>();
builder.Services.AddHostedService<EvolutionBackgroundService>();
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

// 初始化 MetaData 缓存
MetaData.CacheService = app.Services.GetRequiredService<ICacheService>();

// 注入插件管理器到 BotMessage
BotMessage.LLMApp = app.Services.GetRequiredService<LLMApp>();
BotMessage.Pipeline = app.Services.GetRequiredService<MessagePipeline>();
BotMessage.ServiceProvider = app.Services;
LLMApp.ServiceProvider = app.Services;
BotMessage.PluginManager = app.Services.GetRequiredService<PluginManager>();

if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
}

app.UseRouting();
app.MapControllers();
app.Run();

// 简单的启动加载器
public class StartupLoader(IPluginLoaderService loaderService, LLMApp llmApp) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        Log.Information("[Startup] Starting StartupLoader...");
        // 0. 确保内置指令存在
        await BotCmd.EnsureTableCreatedAsync();
        await BotWorker.Infrastructure.Tools.Todo.EnsureTableCreatedAsync();
        
        // 初始化进化系统相关表
        await BotWorker.Modules.AI.Models.Evolution.JobDefinition.EnsureTableCreatedAsync();
        await BotWorker.Modules.AI.Models.Evolution.EmployeeInstance.EnsureTableCreatedAsync();
        await BotWorker.Modules.AI.Models.Evolution.TaskRecord.EnsureTableCreatedAsync();
        await BotWorker.Modules.AI.Models.Evolution.TaskExecution.EnsureTableCreatedAsync();

        // 注入初始岗位
        var jobService = BotMessage.ServiceProvider.GetRequiredService<IJobService>();
        await jobService.SeedJobsAsync();

        await BotCmd.EnsureCommandExistsAsync("设置Key", "设置Key");
        await BotCmd.EnsureCommandExistsAsync("岗位任务", "岗位任务");
        await BotCmd.EnsureCommandExistsAsync("自动开发", "自动开发");
        await BotCmd.EnsureCommandExistsAsync("开启租赁", "开启租赁");
        await BotCmd.EnsureCommandExistsAsync("关闭租赁", "关闭租赁");
        await BotCmd.EnsureCommandExistsAsync("我的Key", "我的Key");
        await BotCmd.EnsureCommandExistsAsync("积分榜", "积分榜");
        await BotCmd.EnsureCommandExistsAsync("后台", "后台");
        
        Log.Information("[Startup] Initializing AI App...");
        // 1. 初始化 AI 提供商
        await llmApp.InitializeAsync();
        
        Log.Information("[Startup] Loading all plugins...");
        // 2. 加载插件
        await loaderService.LoadAllPluginsAsync();
        Log.Information("[Startup] StartupLoader finished.");
    }
}
