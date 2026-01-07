using Mirai.Net.Data.Messages.Concretes;
using Newtonsoft.Json;
using System.Data;
using System.Text.Json;
using System.Web;
using BotWorker.Common;
using BotWorker.Common.Exts;
using sz84.Core.Data;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class Music : MetaData<Music>
    {
        public override string TableName => "Music";
        public override string KeyField => "Id";

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
            if (song == null)
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
            if (arr.GetArrayLength() == 0)
                return [];

            var items = new List<(string Id, string Name, string Artist, string PicId)>();
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

            return doc.RootElement.GetProperty("url").GetString();
        }

        // ------------------------------
        // ③ 获取专辑封面
        // ------------------------------
        private static async Task<string> GetCoverAsync(string picId)
        {
            string url = $"{Api}?types=pic&source=kuwo&id={picId}&size=300";

            string json = await _http.GetStringAsync(url);
            var doc = JsonDocument.Parse(json);

            return doc.RootElement.GetProperty("url").GetString()!;
        }

        public static MusicShareMessage GetMusicShareMessage(long id)
        {
            var msm = new MusicShareMessage();
            using (var dt = QueryDataset($"select top 1 * from {FullName} where Id = {id}"))
            {
                foreach (DataRow dr in dt.Tables[0].Rows)
                {
                    msm.Title = dr["Title"].ToString();
                    msm.Brief = dr["Brief"].ToString();
                    msm.Summary = dr["Summary"].ToString();
                    msm.PictureUrl = dr["PictureUrl"].ToString();
                    msm.MusicUrl = dr["IsVIP"].AsBool() ? dr["MusicUrl2"].ToString() : dr["MusicUrl"].ToString();
                    msm.JumpUrl = dr["JumpUrl"].ToString();
                    msm.Kind = dr["Kind"].ToString();
                }
            }
            return msm;
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

        public static MusicShareMessage ParseMusicPayload(string payloadJson)
        {
            dynamic? data = JsonConvert.DeserializeObject(payloadJson);
            if (data == null) return new MusicShareMessage();

            return new MusicShareMessage
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
            return GetValue("ISNULL(payload, JumpUrl)", id);
        }

        public static long GetMusicId(Song song)
        {
            return song.Kind == MusicKind.ZMMusic 
                ? song.SongId.AsLong() 
                : GetWhere<long>("Id", $"Kind = {GetMusicKind(song.Kind).Quotes()} and SongId={song.SongId.Quotes()}", "");
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
            return !song.SongId.IsNull() && ExistsAandB("Kind", GetMusicKind(song.Kind), "SongId", song.SongId);
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
                    musicUrl = GetWhere("MusicUrl", $"JumpUrl = {jumpUrl.Quotes()}", "Id desc");
                if (!musicUrl.IsNull())
                {
                    var uri = new Uri(musicUrl);
                    var query = HttpUtility.ParseQueryString(uri.Query);
                    song.SongId = query["album_audio_id"];
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
            string sql = $"select top 1 title, brief, summary, pictureurl, musicurl, kind from {FullName} where Kind = {GetMusicKind(song.Kind).Quotes()} and SongId = {song.SongId}";
            return QueryRes(sql, "<a href=\"{4}\" target=\"_blank\">{2}:{0}<br /><img width=\"75\" src=\"{3}\"></a>");                         
        }

        public static string GetSongUrlPublic(Song song)
        {
            string sql = $"select top 1 title, brief, summary, pictureurl, musicurl, kind from {FullName} where Kind = {GetMusicKind(song.Kind).Quotes()} and SongId = {song.SongId}";
            return QueryRes(sql, "✅ 点歌成功!\n<a href=\"{4}\">{2}:{0}\n点击此链接开始播放</a>");
        }

        public static long Append(string Kind, string Title, string Summary, string JumpUrl, string PictureUrl, string MusicUrl, string Brief, string SongId, long groupId, long userId, string payload)
        {
            var covs = new List<Cov>
            {
                new("GroupId", groupId),
                new("UserId", userId),
                new("Kind", Kind),
                new("Title", Title),
                new("Summary", Summary),
                new("JumpUrl", JumpUrl),
                new("PictureUrl", PictureUrl),
                new("MusicUrl", MusicUrl),
                new("Brief", Brief),
                new("SongId", SongId),
            };

            if (!payload.IsNull())
            {
                covs.Add(new Cov("Payload", payload));
                covs.Add(new Cov("IsPayload", true));
                covs.Add(new Cov("PayloadDate", DateTime.MinValue));
            }

            return Insert(covs) == -1 ? 0 : GetAutoId(FullName);
        }

    }
    public class SongResult
    {
        public string Name { get; set; } = "";
        public string Artist { get; set; } = "";
        public string Cover { get; set; } = "";
        public string AudioUrl { get; set; } = "";
    }
}
