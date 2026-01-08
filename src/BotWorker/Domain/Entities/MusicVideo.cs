using Mirai.Net.Data.Messages.Concretes;
using System.Text.RegularExpressions;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class MusicVideo : MetaData<MusicVideo>
    {
        public override string TableName => "MusicVideo";
        public override string KeyField => "Id";

        // 通过MusicVideoID合成App消息
        public static AppMessage GetAppMessage(string mvVid)
        {
            return new AppMessage
            {
                Content = GetWhere("MvContent", $"MvVID='{mvVid}'")
            }; 
        }

        // 收到MV分享消息时添加到MV库
        public static int HandleApp(long botUin, long groupId, long userId, AppMessage message)
        {
            ShowMessage($"content:{message.Content}");
            int i = 0;
            string mv_vid = GetVid(message.Content);
            if (mv_vid != "")
            {
                if (ExistsMv(message.Content))
                    ShowMessage("此MV已存在", ConsoleColor.DarkGreen);
                else
                {
                    i = Append(mv_vid, message.Content, groupId, userId);
                    if (i == -1)
                        ErrorMessage("添加MV失败");
                    else
                        ShowMessage($"✅ 添加MV成功！\nVID：{mv_vid}\nMV数量：{CountAsync}", ConsoleColor.DarkRed);
                }
                //if (Common.IsNum(mv_vid))
                //{
                //    //处理问答库
                //    bool IsDj = message.Title.ToUpper().Contains("DJ");
                //    string dj = IsDj ? "dj" : "";
                //    string title = Regex.Replace(message.Title, @"[\(（][\s|\S]*?[\)）]?$", "");
                //    AppendAnswer(group_id, userId, $"点歌{title}{dj}", song.SongId, message.JumpUrl);
                //    AppendAnswer(group_id, userId, $"点歌{title}{message.Summary}{dj}", song.SongId, message.JumpUrl);
                //    AppendAnswer(group_id, userId, $"点歌{message.Summary}{title}{dj}", song.SongId, message.JumpUrl);
                //    var singers = message.Summary.Split("/");
                //    if (singers.Length > 1)
                //    {
                //        foreach (var singer in singers)
                //        {
                //            AppendAnswer(group_id, userId, $"点歌{title}{singer}{dj}", song.SongId, message.JumpUrl);
                //            AppendAnswer(group_id, userId, $"点歌{singer}{title}{dj}", song.SongId, message.JumpUrl);
                //        }
                //    }
                //}
            }
            return i;
        }

        public static bool ExistsMv(string content)
        {
            return ExistsField("MvVid", GetVid(content));
        }

        public static string GetVid(string content)
        {
            string res = "";
            if (content.IsMatch(Regexs.MusicVideo))
            {                
                MatchCollection matches = content.Matches(Regexs.MusicVideo);
                foreach (Match match in matches.Cast<Match>())
                {
                    res = match.Groups["vid"].Value;
                }
            }
            else
                res = "";
            return res;
        }

        public static int Append(string mvVid, string mvContent, long groupId, long userId)
        {
            return Insert([
                            new Cov("MvVid", mvVid),
                            new Cov("MvContent", mvContent),
                            new Cov("GroupId", groupId),
                            new Cov("UserId", userId)
                        ]);
        }
    }
}
