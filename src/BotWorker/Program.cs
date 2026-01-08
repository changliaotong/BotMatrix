using Serilog;
using BotWorker.Application.Messaging.Pipeline;
using BotWorker.Common.Config;
using BotWorker.Application.Messaging.Handlers;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Modules.Plugins;
using BotWorker.Application.Services;
using BotWorker.Infrastructure.Persistence.Database;
using BotWorker.Infrastructure.Utils;

var builder = WebApplication.CreateBuilder(args);

// 初始化静态配置
AppConfig.Initialize(builder.Configuration);

// 配置 Serilog
Log.Logger = new LoggerConfiguration()
    .WriteTo.Console()
    .CreateLogger();
builder.Host.UseSerilog();

// 添加基础服务
builder.Services.AddControllers();
builder.Services.AddSignalR();
builder.Services.AddSingleton<Microsoft.AspNetCore.SignalR.IHubFilter, BotWorker.Infrastructure.SignalR.HubLoggingFilter>();
builder.Services.AddHttpClient();

// 注册核心业务服务
builder.Services.AddSingleton<IEventNexus, EventNexus>();
builder.Services.AddSingleton<IMcpService, MCPManager>();
builder.Services.AddSingleton<IAIService, AIService>();
builder.Services.AddSingleton<II18nService, I18nService>();
builder.Services.AddSingleton<PluginManager>();
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
builder.Services.AddSingleton<AiMiddleware>();
builder.Services.AddTransient<AutoSignInMiddleware>();

  // 注册指令处理器
   builder.Services.AddSingleton<AdminCommandHandler>();
   builder.Services.AddSingleton<SetupCommandHandler>();
   builder.Services.AddSingleton<HotCommandHandler>();
   builder.Services.AddSingleton<GameCommandHandler>();

  // 注册并配置 Pipeline
builder.Services.AddSingleton<MessagePipeline>(sp => 
{
    var pipeline = new MessagePipeline(sp);
    // 1. 全局异常处理
    pipeline.Use(sp.GetRequiredService<ExceptionMiddleware>());
    // 2. 最终消息加工
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
    // 13. AI/智能体
    pipeline.Use(sp.GetRequiredService<AiMiddleware>());
    // 14. 自动化业务 (自动签到)
    pipeline.Use(sp.GetRequiredService<AutoSignInMiddleware>());
    // 15. 业务插件级分发 (普通插件)
    pipeline.Use(sp.GetRequiredService<PluginMiddleware>());
    return pipeline;
});

// 注册启动时加载插件的任务
builder.Services.AddHostedService<StartupPluginLoader>();

var app = builder.Build();

// 注入插件管理器到 BotMessage
BotMessage.PluginManager = app.Services.GetRequiredService<PluginManager>();
BotMessage.Pipeline = app.Services.GetRequiredService<MessagePipeline>();

// 检查是否为测试模式
if (args.Contains("--test"))
{
    await BotWorker.TestConsole.RunAsync(builder.Configuration);
    return;
}

if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
}

app.UseRouting();
app.MapControllers();
app.MapHub<ChatHub>("/chatHub");

app.Run();

// 简单的启动插件加载器
public class StartupPluginLoader(IPluginLoaderService loaderService) : BackgroundService
{
    protected override async Task ExecuteAsync(CancellationToken stoppingToken)
    {
        await loaderService.LoadAllPluginsAsync();
    }
}
