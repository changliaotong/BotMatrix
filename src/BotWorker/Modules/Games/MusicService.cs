using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Microsoft.Extensions.Logging;
using System.Text;
using System.Web;
using System.Text.Json;
using BotWorker.Domain.Entities;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.music",
        Name = "超级点歌系统",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "独立点歌系统：支持全网搜歌、送歌给好友、点歌记录查询",
        Category = "Games"
    )]
    public class MusicService : IPlugin
    {
        private IRobot? _robot;
        private readonly ILogger<MusicService> _logger;
        private readonly ISongOrderRepository _orderRepo;
        private static readonly HttpClient _http = new(new HttpClientHandler
        {
            ServerCertificateCustomValidationCallback = HttpClientHandler.DangerousAcceptAnyServerCertificateValidator
        });

        private const string Api = "https://music-api.gdstudio.xyz/api.php";

        public List<Intent> Intents => [
            new() { Name = "点歌", Keywords = ["点歌", "music"] },
            new() { Name = "送歌", Keywords = ["送歌", "give"] },
            new() { Name = "点歌历史", Keywords = ["点歌历史", "musiclog"] }
        ];

        public MusicService(ILogger<MusicService> logger, ISongOrderRepository orderRepo)
        {
            _logger = logger;
            _orderRepo = orderRepo;
        }

        public async Task InitAsync(IRobot robot)
        {
            _robot = robot;
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "点歌系统",
                Commands = ["点歌", "送歌", "点歌历史"],
                Description = "【点歌 歌名】直接听歌；【送歌 @某人 歌名 [寄语]】传情达意"
            }, HandleMusicCommandAsync);
        }

        public async Task StopAsync() => await Task.CompletedTask;

        private async Task EnsureTablesCreatedAsync()
        {
            await _orderRepo.EnsureTableCreatedAsync();
        }

        private async Task<string> HandleMusicCommandAsync(IPluginContext ctx, string[] args)
        {
            var cmd = ctx.RawMessage.Trim().Split(' ')[0];
            return cmd switch
            {
                "点歌" or "music" => await OrderSongAsync(ctx, args),
                "送歌" or "give" => await GiveSongAsync(ctx, args),
                "点歌历史" or "musiclog" => await GetMusicLogAsync(ctx),
                _ => "🎵 想要听歌？试试【点歌 歌名】或者【送歌 @某人 歌名】吧！"
            };
        }

        private async Task<string> OrderSongAsync(IPluginContext ctx, string[] args)
        {
            if (args.Length == 0) return "你想听什么歌？请输入歌名，例如：点歌 晴天";
            var keyword = string.Join(" ", args);

            var song = await SearchSongInternalAsync(keyword);
            if (song == null) return "❌ 没找到这首歌，换个关键词试试吧。";

            // 发送音乐卡片 (这里需要 IRobotClient 的支持，通常通过返回特定格式或调用 API)
            // 假设我们可以通过 ctx 直接发送复杂消息
            await ctx.SendMusicAsync(song.Name, song.Artist, song.AudioUrl, song.Cover, song.AudioUrl);

            return $"🎧 正在为你播放：{song.Name} - {song.Artist}";
        }

        private async Task<string> GiveSongAsync(IPluginContext ctx, string[] args)
        {
            // 格式：送歌 @用户 歌名 [寄语]
            if (args.Length < 2) return "使用方法：送歌 @用户 歌名 [寄语]";

            // 解析目标用户 (假设第一个参数是 @提及)
            var targetUserId = ""; 
            var targetNickname = "TA";
            var startIndex = 0;

            if (ctx.MentionedUsers.Count > 0)
            {
                var target = ctx.MentionedUsers[0];
                targetUserId = target.UserId;
                targetNickname = target.Name;
                startIndex = 1; // 跳过 @提及
            }
            else
            {
                // 可能是文字提及或需要解析
                return "请 @ 一个你想送歌的好友！";
            }

            var songArgs = args.Skip(startIndex).ToList();
            if (songArgs.Count == 0) return "你想送什么歌？请输入歌名。";

            var keyword = songArgs[0];
            var message = songArgs.Count > 1 ? string.Join(" ", songArgs.Skip(1)) : "愿这首歌带给你好心情！";

            var song = await SearchSongInternalAsync(keyword);
            if (song == null) return "❌ 没找到这首歌，换个关键词试试吧。";

            // 保存记录
            var order = new SongOrder
            {
                FromUserId = ctx.UserId,
                FromNickname = ctx.UserName,
                ToUserId = targetUserId,
                ToNickname = targetNickname,
                SongName = song.Name,
                Artist = song.Artist,
                Message = message
            };
            await _orderRepo.InsertAsync(order);

            // 发送通知给目标用户 (如果是群聊，可能需要 @TA)
            await ctx.SendMusicAsync(song.Name, song.Artist, song.AudioUrl, song.Cover, song.AudioUrl);

            return $"💌 成功送出心意！\n🎁 送给：{targetNickname}\n🎵 歌曲：{song.Name}\n📝 寄语：{message}";
        }

        private async Task<string> GetMusicLogAsync(IPluginContext ctx)
        {
            var logs = await _orderRepo.GetHistoryAsync(ctx.UserId);
            if (logs.Count == 0) return "📭 你还没有点过歌，或者还没有收到过别人的赠歌。";

            var sb = new StringBuilder();
            sb.AppendLine("📜 【最近点歌/收歌记录】");
            foreach (var log in logs.Take(10))
            {
                var role = log.FromUserId == ctx.UserId ? "送给" : "收到";
                var other = log.FromUserId == ctx.UserId ? log.ToNickname : log.FromNickname;
                sb.AppendLine($"[{log.OrderTime:MM-dd}] {role} {other}: 《{log.SongName}》");
            }

            return sb.ToString();
        }

        private async Task<SongResult?> SearchSongInternalAsync(string keyword)
        {
            try
            {
                var url = $"{Api}?msg={HttpUtility.UrlEncode(keyword)}&type=json";
                var json = await _http.GetStringAsync(url);
                var result = JsonSerializer.Deserialize<SongResult>(json, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                return result;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Search song failed for keyword: {Keyword}", keyword);
                return null;
            }
        }
    }
}
