using Microsoft.CodeAnalysis.CSharp.Syntax;
using Newtonsoft.Json;
using OneBotSharp.Objs.Message;
using System.Text.RegularExpressions;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Platform;
using BotWorker.Bots.Services;
using BotWorker.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Models;
using BotWorker.Core.Data;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {


        //收到音乐分享消息时添加到音乐库 (使用 OneBot 格式)
        public void HandleMusic(OneBotSharp.Objs.Message.MsgMusic.MsgData message, string payload, bool isForce = false)
        {            
            Logger.Show($"kind:{message.Type}\n Title:{message.Title}\n Summary:{message.Content}\n JumpUrl:{message.Url}\n PictureUrl:{message.Image}\n MusicUrl:{message.Audio}\n Brief:{message.Content}");
            Song song = Music.GetSong(message.Url ?? "", message.Audio ?? "");
            long musicId = 0;
            string songId = song.SongId.AsString();
            if (!songId.IsNull() || isForce)
            {                
                string info;
                //处理音乐库
                if (!isForce && Music.ExistsSong(message.Url ?? "", message.Audio ?? ""))
                    info = "此音乐已存在";
                else
                {
                    musicId = Music.Append(message.Type ?? "", message.Title ?? "", message.Content ?? "", message.Url ?? "", message.Image ?? "", message.Audio ?? "", message.Content ?? "", song?.SongId ?? "", GroupId, UserId, payload);
                    if (musicId == 0)
                        info = "添加音乐失败";
                    else
                        info = $"✅ 添加成功！ \nMusicId: {musicId} SongId：{song?.SongId ?? ""} Music数量：{Count()}\n{message.Title} {message.Content}";
                }
                Logger.Show(info);
                if (!songId.IsNull() || (isForce && musicId != 0))
                {                    
                    //处理问答库
                    bool IsDj = (message.Title ?? "").Contains("DJ", StringComparison.CurrentCultureIgnoreCase);
                    string dj = IsDj ? "dj" : "";
                    string title = (message.Title ?? "").RegexReplace(@"[\(（][\s|\S]*?[\)）]?$", "");
                    string summary = (message.Content ?? "").RegexReplace(@"《[\s|\S]*?》?$", "");
                    var jumpUrl = isForce ? $"https://sz84.com/music/{musicId}" : message.Url;
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{dj}", songId, jumpUrl ?? "");
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{summary}{dj}", songId, jumpUrl ?? "");
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{summary}{title}{dj}", songId, jumpUrl ?? "");
                    var singers = (message.Content ?? "").Split("/");
                    if (singers.Length > 1)
                    {
                        foreach (var singer in singers)
                        {
                            MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{singer}{dj}", songId, jumpUrl ?? "");
                            MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{singer}{title}{dj}", songId, jumpUrl ?? "");
                        }
                    }
                }
            }
        }

        //送歌
        public static (long recipientId, string songName) ParseSendMusicFlexible(string cmdPara)
        {
            if (cmdPara.IsNull()) return (0, "");

            cmdPara = cmdPara.Trim();
            string recipient = "";
            string song = cmdPara;

            // 1. 尝试匹配 [@123456] 或 [@:123456]
            var matchAt = Regex.Match(cmdPara, @"\[@:?(\d+)\]");
            if (matchAt.Success)
            {
                recipient = matchAt.Groups[1].Value;
                song = cmdPara.Replace(matchAt.Value, "").Trim();
                return (recipient.AsLong(), song);
            }

            // 2. 尝试匹配开头的数字
            var matchStartNumber = Regex.Match(cmdPara, @"^(\d+)\s+(.+)$");
            if (matchStartNumber.Success)
            {
                recipient = matchStartNumber.Groups[1].Value;
                song = matchStartNumber.Groups[2].Value.Trim();
                return (recipient.AsLong(), song);
            }

            // 3. 尝试匹配结尾的数字
            var matchEndNumber = Regex.Match(cmdPara, @"^(.+?)\s+(\d+)$");
            if (matchEndNumber.Success)
            {
                song = matchEndNumber.Groups[1].Value.Trim();
                recipient = matchEndNumber.Groups[2].Value;
                return (recipient.AsLong(), song);
            }

            // 4. 没找到收件人
            return (0, cmdPara);
        }


        public async Task GetMusicResAsync(string cmdPara2 = "")
        {
            if (!Platform.In(Platforms.Mirai, Platforms.NapCat, Platforms.QQGuild))
            {
                Answer = $"当前版本机器人暂无{CmdName}功能";
                return;
            }

            // 识别 "送歌" 功能
            if (CmdName == "送歌")
            {
                (TargetUin, var realPara) = ParseSendMusicFlexible(CmdPara);
                CmdPara = realPara;
                IsAtOthers = false;
            }

            var cmdPara = CmdPara;
            if (cmdPara.Trim() == "随机")
            {
                (AnswerId, Answer) = MusicAnswer.RandomMusic();
                IsMusic = true;
                return;
            }
            CmdPara = $"点歌{cmdPara}{cmdPara2}";
            IsMusic = true;
            await GetAnswerAsync();
            if (Answer.IsNull())
            {
                cmdPara = cmdPara.Replace("点歌", "");
                cmdPara2 += "%";
                cmdPara = QuestionInfo.GetWhere("Question", $"Question like '点歌%{cmdPara.DoubleQuotes()}%{cmdPara2.DoubleQuotes()}' and CAnswer > 0 ", "newid()");
                if (!cmdPara.IsNull())
                {
                    CmdPara = cmdPara;
                    await GetAnswerAsync();
                }
            }

            if (Answer.IsNull())
            {
                var service = new Music();
                var result = await Music.SearchSongAsync(CmdPara);
                if (result == null || result.Name.IsNull())
                    Answer = "没找到这首歌，换一首吧";
                else
                {                    
                    SongResult = result;
                    Answer = $"[Music] 歌名：{result.Name} 演唱者：{result.Artist} 封面：{result.Cover} 音源：{result.AudioUrl}";
                    HandleMusic(SongResult.ToMusicData(), "", true);
                }
            }

            if (Answer.IsMatch($"{Regexs.SongId}|{Regexs.SongIdNetease}|{Regexs.SongIdNetease2}|{Regexs.SongIdKugou}"))
            {
                var msm = Music.GetMusicShareMessage(Music.GetSong(Answer).MusicId); ;
                if (msm == null) return;
                Song song = Music.GetSong(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "");
                if (song == null) return;
                Answer = OneBotSharp.Objs.Message.MsgMusic.BuildCustom(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "", msm?.Title ?? "", msm?.Summary, User.IsMusicLogo ? BotWorker.Bots.Users.UserInfo.GetHead(UserId) : msm?.PictureUrl).BuildSendCq();
            }
            else if (Answer.IsMatch(Regexs.MusicIdZaomiao))
            {
                var msm = Music.GetMusicShareMessage(Music.GetSong(Answer).MusicId);
                Answer = OneBotSharp.Objs.Message.MsgMusic.BuildCustom(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "", msm?.Title ?? "", msm?.Summary, User.IsMusicLogo ? BotWorker.Bots.Users.UserInfo.GetHead(UserId) : msm?.PictureUrl).BuildSendCq();
            }
            else if (IsMusic && Answer.StartsWith("[Music]"))
            {
                Answer = OneBotSharp.Objs.Message.MsgMusic.BuildCustom(SongResult?.AudioUrl ?? "", SongResult?.AudioUrl ?? "", SongResult?.Name ?? "", SongResult?.Artist, User.IsMusicLogo ? BotWorker.Bots.Users.UserInfo.GetHead(UserId) : SongResult?.Cover).BuildSendCq();
            }
        }
    }

}
