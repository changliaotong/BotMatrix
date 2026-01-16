using OneBotSharp.Objs.Message;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{


        //收到音乐分享消息时添加到音乐库
        public void HandleMusic(BotWorker.Models.MusicShareMessage message, string payload, bool isForce = false)
        {
            // ShowMessage($"kind:{message.Kind}\n Title:{message.Title}\n Summary:{message.Summary}\n JumpUrl:{message.JumpUrl}\n PictureUrl:{message.PictureUrl}\n MusicUrl:{message.MusicUrl}\n Brief:{message.Brief}");
            Song song = Music.GetSong(message.JumpUrl, message.MusicUrl);
            long musicId = 0;
            string songId = song.SongId.AsString();
            if (!songId.IsNull() || isForce)
            {
                string info;
                //处理音乐库
                if (!isForce && Music.ExistsSong(message.JumpUrl, message.MusicUrl))
                    info = "此音乐已存在";
                else
                {
                    musicId = Music.Append(message.Kind, message.Title, message.Summary, message.JumpUrl, message.PictureUrl, message.MusicUrl, message.Brief, song?.SongId ?? "", GroupId, UserId, payload);
                    if (musicId == 0)
                        info = "添加音乐失败";
                    else
                        info = $"✅ 添加成功！ \nMusicId: {musicId} SongId：{song?.SongId ?? ""} Music数量：{Count()}\n{message.Title} {message.Summary}";
                }
                // ShowMessage(info);
                if (!songId.IsNull() || (isForce && musicId != 0))
                {
                    //处理问答库
                    bool IsDj = message.Title.Contains("DJ", StringComparison.CurrentCultureIgnoreCase);
                    string dj = IsDj ? "dj" : "";
                    string title = message.Title.RegexReplace(@"[\(（][\s|\S]*?[\)）]?$", "");
                    string summary = message.Summary.RegexReplace(@"《[\s|\S]*?》?$", "");
                    var jumpUrl = isForce ? $"https://sz84.com/music/{musicId}" : message.JumpUrl;
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{dj}", songId, jumpUrl);
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{summary}{dj}", songId, jumpUrl);
                    MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{summary}{title}{dj}", songId, jumpUrl);
                    var singers = message.Summary.Split("/");
                    if (singers.Length > 1)
                    {
                        foreach (var singer in singers)
                        {
                            MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{title}{singer}{dj}", songId, jumpUrl);
                            MusicAnswer.Append(SelfId, GroupId, UserId, $"点歌{singer}{title}{dj}", songId, jumpUrl);
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
            if (!Platform.In(Platforms.Mirai, Platforms.QQ, Platforms.QQGuild))
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
                    HandleMusic(SongResult.ToMusicShareMessage(), "", true);
                }
            }

            if (Answer.IsMatch($"{Regexs.SongId}|{Regexs.SongIdNetease}|{Regexs.SongIdNetease2}|{Regexs.SongIdKugou}"))
            {
                var msm = Music.GetMusicShareMessage(Music.GetSong(Answer).MusicId); ;
                if (msm == null) return;
                Song song = Music.GetSong(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "");
                if (song == null) return;
                Answer = MsgMusic.BuildCustom(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "", msm?.Title ?? "", msm?.Summary, User.IsMusicLogo ? UserInfo.GetHead(UserId) : msm?.PictureUrl).BuildSendCq();
            }
            else if (Answer.IsMatch(Regexs.MusicIdZaomiao))
            {
                var msm = Music.GetMusicShareMessage(Music.GetSong(Answer).MusicId);
                Answer = MsgMusic.BuildCustom(msm?.JumpUrl ?? "", msm?.MusicUrl ?? "", msm?.Title ?? "", msm?.Summary, User.IsMusicLogo ? UserInfo.GetHead(UserId) : msm?.PictureUrl).BuildSendCq();
            }
            else if (IsMusic && Answer.StartsWith("[Music]"))
            {
                Answer = MsgMusic.BuildCustom(SongResult?.AudioUrl ?? "", SongResult?.AudioUrl ?? "", SongResult?.Name ?? "", SongResult?.Artist, User.IsMusicLogo ? UserInfo.GetHead(UserId) : SongResult?.Cover).BuildSendCq();
            }
        }
    }

}
