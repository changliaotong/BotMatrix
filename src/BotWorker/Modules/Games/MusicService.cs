using BotWorker.Domain.Interfaces;
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
        private ILogger? _logger;
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

        public MusicService(ILogger<MusicService> logger)
        {
            _logger = logger;
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
            try
            {
                var checkTable = await SongOrder.QueryScalarAsync<int>("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'UserSongOrders'");
                if (checkTable == 0)
                {
                    var sql = BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<SongOrder>();
                    await SongOrder.ExecAsync(sql);
                }
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "MusicService 数据库初始化失败");
            }
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
                startIndex = 1;
            }
            else
            {
                return "请在指令中 @ 你想送歌的好友！";
            }

            var songKeyword = args[startIndex];
            var message = args.Length > startIndex + 1 ? string.Join(" ", args.Skip(startIndex + 1)) : "送你一首歌，祝你开心每一天！";

            var song = await SearchSongInternalAsync(songKeyword);
            if (song == null) return "❌ 没找到这首歌，无法送出。";

            // 保存记录
            var order = new SongOrder
            {
                FromUserId = ctx.UserId,
                FromNickname = ctx.UserName,
                ToUserId = targetUserId,
                ToNickname = targetNickname,
                SongName = song.Name,
                Artist = song.Artist,
                Message = message,
                OrderTime = DateTime.Now
            };
            await order.InsertAsync();

            // 发送通知给目标
            await ctx.SendMusicAsync(song.Name, song.Artist, song.AudioUrl, song.Cover, song.AudioUrl);

            var sb = new StringBuilder();
            sb.AppendLine($"🎁 送歌成功！");
            sb.AppendLine($"来自 {ctx.UserName} 的礼物已送达给 {targetNickname}。");
            sb.AppendLine($"💬 寄语：{message}");
            return sb.ToString();
        }

        private async Task<string> GetMusicLogAsync(IPluginContext ctx)
        {
            var logs = await SongOrder.GetHistoryAsync(ctx.UserId);
            if (logs.Count == 0) return "你还没有点歌或收到歌的历史记录。";

            var sb = new StringBuilder();
            sb.AppendLine("📜 【点歌历史】");
            sb.AppendLine("━━━━━━━━━━━━━━");
            foreach (var log in logs.Take(10))
            {
                var type = log.FromUserId == ctx.UserId ? "📤 送出" : "📥 收到";
                var partner = log.FromUserId == ctx.UserId ? log.ToNickname : log.FromNickname;
                sb.AppendLine($"{log.OrderTime:MM-dd HH:mm} {type} {partner}");
                sb.AppendLine($"   🎵 {log.SongName} - {log.Artist}");
            }
            sb.AppendLine("━━━━━━━━━━━━━━");
            return sb.ToString();
        }

        private async Task<SongResult?> SearchSongInternalAsync(string keyword)
        {
            try
            {
                string searchUrl = $"{Api}?types=search&source=kuwo&name={HttpUtility.UrlEncode(keyword)}&count=1&pages=1";
                string json = await _http.GetStringAsync(searchUrl);
                using var doc = JsonDocument.Parse(json);
                var arr = doc.RootElement;
                if (arr.GetArrayLength() == 0) return null;

                var item = arr[0];
                var id = item.GetProperty("id").GetString()!;
                var name = item.GetProperty("name").GetString()!;
                var artist = string.Join("/", item.GetProperty("artist").EnumerateArray().Select(a => a.GetString()));
                var picId = item.GetProperty("pic_id").GetString()!;

                // 获取 URL
                string urlReq = $"{Api}?types=url&source=kuwo&id={id}&br=320";
                string urlJson = await _http.GetStringAsync(urlReq);
                using var urlDoc = JsonDocument.Parse(urlJson);
                var audioUrl = urlDoc.RootElement.GetProperty("url").GetString();

                // 获取封面
                string picReq = $"{Api}?types=pic&source=kuwo&id={picId}";
                string picJson = await _http.GetStringAsync(picReq);
                using var picDoc = JsonDocument.Parse(picJson);
                var cover = picDoc.RootElement.GetProperty("url").GetString();

                return new SongResult
                {
                    Name = name,
                    Artist = artist,
                    AudioUrl = audioUrl ?? "",
                    Cover = cover ?? "",
                    Source = "kuwo"
                };
            }
            catch (Exception ex)
            {
                _logger?.LogError(ex, "搜歌失败: {Keyword}", keyword);
                return null;
            }
        }
    }
}
