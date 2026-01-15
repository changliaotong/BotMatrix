using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Net.Http;
using System.Text.Json;
using System.Threading.Tasks;
using System.Web;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Newtonsoft.Json;

namespace BotWorker.Domain.Entities
{
    [Table("Music")]
    public class Music
    {
        private static IMusicRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IMusicRepository>() 
            ?? throw new InvalidOperationException("IMusicRepository not registered");

        [Key]
        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string Kind { get; set; } = string.Empty;
        public string Title { get; set; } = string.Empty;
        public string Summary { get; set; } = string.Empty;
        public string JumpUrl { get; set; } = string.Empty;
        public string PictureUrl { get; set; } = string.Empty;
        public string MusicUrl { get; set; } = string.Empty;
        public string MusicUrl2 { get; set; } = string.Empty;
        public string Brief { get; set; } = string.Empty;
        public string SongId { get; set; } = string.Empty;
        public string Payload { get; set; } = string.Empty;
        public bool IsPayload { get; set; }
        public DateTime PayloadDate { get; set; }
        public bool IsVIP { get; set; }
        public DateTime InsertDate { get; set; }

        // 忽略 SSL 证书错误（解决你的异常）
        private static readonly HttpClient _http = new(
            new HttpClientHandler
            {
                ServerCertificateCustomValidationCallback = HttpClientHandler.DangerousAcceptAnyServerCertificateValidator
            }
        );

        private const string Api = "https://music-api.gdstudio.xyz/api.php";

        public static async Task<SongResult?> SearchSongAsync(string keyword)
        {
            keyword = keyword.Replace("点歌", "").Trim();
            if (string.IsNullOrWhiteSpace(keyword))
                return null;

            // ① 搜索歌曲（使用稳定源：kuwo）
            var song = await SearchKuwoSongAsync(keyword);
            if (song == null || song.Count == 0)
                return null;

            // ② 获取真实音频 URL（高音质）
            string? audio = await GetSongUrlAsync(song[0].Id);
            if (audio == null)
                audio = "";

            return new SongResult
            {
                Name = song[0].Name,
                Artist = song[0].Artist,
                Cover = await GetCoverAsync(song[0].PicId),
                AudioUrl = audio
            };
        }

        // ------------------------------
        // ① 搜索歌曲
        // ------------------------------
        public static async Task<List<(string Id, string Name, string Artist, string PicId)>> SearchKuwoSongAsync(string keyword)
        {
            string url = $"{Api}?types=search&source=kuwo&name={HttpUtility.UrlEncode(keyword)}&count=3&pages=1";

            string json = await _http.GetStringAsync(url);
            var doc = JsonDocument.Parse(json);

            var arr = doc.RootElement;
            if (arr.ValueKind == JsonValueKind.Array && arr.GetArrayLength() == 0)
                return [];

            var items = new List<(string Id, string Name, string Artist, string PicId)>();
            if (arr.ValueKind == JsonValueKind.Array)
            {
                for (int i = 0; i < arr.GetArrayLength(); i++)
                {
                    var item = arr[i];
                    items.Add(
                        (
                            Id: item.GetProperty("id").GetString()!,
                            Name: item.GetProperty("name").GetString()!,
                            Artist: string.Join("/", item.GetProperty("artist").EnumerateArray().Select(a => a.GetString())),
                            PicId: item.GetProperty("pic_id").GetString()!
                        )
                    );
                }
            }
            
            return items;
        }

        // ------------------------------
        // ② 获取真实音频 URL
        // ------------------------------
        public static async Task<string?> GetSongUrlAsync(string trackId)
        {
            string url = $"{Api}?types=url&source=kuwo&id={trackId}&br=999";

            string json = await _http.GetStringAsync(url);
            var doc = JsonDocument.Parse(json);

            if (doc.RootElement.TryGetProperty("url", out var urlProp))
            {
                return urlProp.GetString();
            }
            return null;
        }

        // ------------------------------
        // ③ 获取专辑封面
        // ------------------------------
        private static async Task<string> GetCoverAsync(string picId)
        {
            string url = $"{Api}?types=pic&source=kuwo&id={picId}&size=300";

            string json = await _http.GetStringAsync(url);
            var doc = JsonDocument.Parse(json);

            if (doc.RootElement.TryGetProperty("url", out var urlProp))
            {
                return urlProp.GetString()!;
            }
            return "";
        }

        public static BotWorker.Models.MusicShareMessage GetMusicShareMessage(long id)
        {
            return Repository.GetMusicShareMessageAsync(id).GetAwaiter().GetResult() ?? new BotWorker.Models.MusicShareMessage();
        }

        public static string GetString(dynamic value)
        {
            try
            {
                return value == null ? "" : (string)value;
            }
            catch
            {
                return value?.ToString() ?? "";
            }
        }

        public static BotWorker.Models.MusicShareMessage ParseMusicPayload(string payloadJson)
        {
            dynamic? data = JsonConvert.DeserializeObject(payloadJson);
            if (data == null) return new BotWorker.Models.MusicShareMessage();

            return new BotWorker.Models.MusicShareMessage
            {
                Kind = GetString(data.meta.music.tag).Replace("QQ音乐", "QQMusic").Replace("网易云音乐", "NeteaseCloudMusic").Replace("酷狗音乐", "KugouMusic"),
                Title = data.meta.music.title,
                Summary = data.meta.music.desc,
                JumpUrl = data.meta.music.jumpUrl,
                PictureUrl = data.meta.music.preview,
                MusicUrl = data.meta.music.musicUrl,
                Brief = data.prompt,
            };
        }

        public static string GetMusicSharePayload(long id)
        {
            return Repository.GetPayloadAsync(id).GetAwaiter().GetResult();
        }

        public static long GetMusicId(Song song)
        {
            return song.Kind == MusicKind.ZMMusic 
                ? song.SongId.AsLong() 
                : Repository.GetMusicIdAsync(GetMusicKind(song.Kind), song.SongId).GetAwaiter().GetResult();
        }

        public static string GetMusicKind(MusicKind mk)
        {
            return mk switch
            {
                MusicKind.QQMusic => "QQMusic",
                MusicKind.NeteaseCloudMusic => "NeteaseCloudMusic",
                MusicKind.KugouMusic => "KugouMusic",
                _ => "",
            };
        }

        public static string GetMusicType(string mk)
        {
            return mk switch
            {
                "QQMusic" => "qq",
                "NeteaseCloudMusic" => "n163",
                "KugouMusic" => "kugou",
                _ => "qq",
            };
        }

        public static bool ExistsSong(string jumpUrl, string musicUrl = "")
        {
            Song song = GetSong(jumpUrl, musicUrl);
            if (song.SongId.IsNull()) return false;
            
            // Check existence via ID
            long id = Repository.GetMusicIdAsync(GetMusicKind(song.Kind), song.SongId).GetAwaiter().GetResult();
            return id > 0;
        }

        public static Song GetSong(string jumpUrl, string musicUrl = "")
        {
            Song song = new();

            if (jumpUrl.IsMatch(Regexs.SongId))
            {
                song.Kind = MusicKind.QQMusic;
                song.SongId = jumpUrl.RegexReplace(Regexs.SongId, "$1").Trim();
            }
            else if (jumpUrl.IsMatch(Regexs.SongIdNetease))
            {
                song.Kind = MusicKind.NeteaseCloudMusic;
                song.SongId = jumpUrl.RegexReplace(Regexs.SongIdNetease, "$1").Trim();
            }
            else if (jumpUrl.IsMatch(Regexs.SongIdNetease2))
            {
                song.Kind = MusicKind.NeteaseCloudMusic;
                song.SongId = jumpUrl.RegexReplace(Regexs.SongIdNetease2, "$1").Trim();
            }
            else if (jumpUrl.IsMatch(Regexs.SongIdKugou))
            {
                song.Kind = MusicKind.KugouMusic;
                if (musicUrl.IsNull())
                    musicUrl = Repository.GetMusicUrlByJumpUrlAsync(jumpUrl).GetAwaiter().GetResult();
                
                if (!musicUrl.IsNull())
                {
                    var uri = new Uri(musicUrl);
                    var query = HttpUtility.ParseQueryString(uri.Query);
                    song.SongId = query["album_audio_id"] ?? "";
                }
            }
            else if (jumpUrl.IsMatch(Regexs.MusicIdZaomiao))
            {
                song.Kind = MusicKind.ZMMusic;
                song.SongId = jumpUrl.RegexReplace(Regexs.MusicIdZaomiao, "$1").Trim();
                song.MusicId = song.SongId.AsLong();
            }

            song.MusicId = GetMusicId(song);

            return song;
        }

        public static string GetSongUrl(Song song)
        {
            return Repository.GetMusicUrlAsync(GetMusicKind(song.Kind), song.SongId).GetAwaiter().GetResult();
        }

        public static string GetSongUrlPublic(Song song)
        {
            return Repository.GetMusicUrlPublicAsync(GetMusicKind(song.Kind), song.SongId).GetAwaiter().GetResult();
        }

        public static long Append(string Kind, string Title, string Summary, string JumpUrl, string PictureUrl, string MusicUrl, string Brief, string SongId, long groupId, long userId, string payload)
        {
            var music = new Music
            {
                GroupId = groupId,
                UserId = userId,
                Kind = Kind,
                Title = Title,
                Summary = Summary,
                JumpUrl = JumpUrl,
                PictureUrl = PictureUrl,
                MusicUrl = MusicUrl,
                Brief = Brief,
                SongId = SongId,
                InsertDate = DateTime.Now
            };

            if (!payload.IsNull())
            {
                music.Payload = payload;
                music.IsPayload = true;
                music.PayloadDate = DateTime.MinValue; // Default value logic?
            }

            return Repository.AddAsync(music).GetAwaiter().GetResult();
        }
    }
}
