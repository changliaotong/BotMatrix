using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using BotWorker.Services;
using BotWorker.Core.Plugin;
using sz84.Core.Services;
using sz84.Infrastructure.Caching;
using Serilog;
using Microsoft.EntityFrameworkCore;
using sz84.Infrastructure.Background;

var builder = WebApplication.CreateBuilder(args);

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
builder.Services.AddSingleton<PluginManager>();
builder.Services.AddSingleton<IPluginLoaderService, PluginLoaderService>();
builder.Services.AddSingleton<IMCPHost, PluginMcpHost>();

// 注册启动时加载插件的任务
builder.Services.AddHostedService<StartupPluginLoader>();

var app = builder.Build();

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
