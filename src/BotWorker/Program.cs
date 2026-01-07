using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using System.Linq;
using BotWorker.Services;
using BotWorker.Core.Plugin;
using BotWorker.Core.Services;
using BotWorker.Infrastructure.Caching;
using Serilog;
using Microsoft.EntityFrameworkCore;
using BotWorker.Infrastructure.Background;
using BotWorker.Core.Pipeline;

using BotWorker.Core.Configurations;

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
builder.Services.AddHttpClient();

// 注册数据库
builder.Services.AddDbContext<BotDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("DefaultConnection")));

// 注册 Redis
builder.Services.AddSingleton<EntityCacheHelper>(sp => 
    new EntityCacheHelper(builder.Configuration.GetConnectionString("Redis") ?? "localhost"));

// 注册核心业务服务
builder.Services.AddSingleton<IMcpService, MCPManager>();
builder.Services.AddSingleton<IAIService, AIService>();
builder.Services.AddSingleton<II18nService, I18nService>();
builder.Services.AddSingleton<PluginManager>();
builder.Services.AddSingleton<MessagePipeline>();
builder.Services.AddSingleton<IPluginLoaderService, PluginLoaderService>();
builder.Services.AddSingleton<IMCPHost, PluginMcpHost>();

// 注册启动时加载插件的任务
builder.Services.AddHostedService<StartupPluginLoader>();

var app = builder.Build();

// 注入插件管理器到 BotMessage
BotWorker.Bots.BotMessages.BotMessage.PluginManager = app.Services.GetRequiredService<PluginManager>();
BotWorker.Bots.BotMessages.BotMessage.Pipeline = app.Services.GetRequiredService<MessagePipeline>();

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
