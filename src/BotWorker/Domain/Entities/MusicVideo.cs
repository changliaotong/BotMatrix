using System;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Domain.Repositories;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;
using Mirai.Net.Data.Messages.Concretes;
using BotWorker.Common;
using System.Linq;

namespace BotWorker.Domain.Entities
{
    [Table("music_video")]
    public class MusicVideo
    {
        private static IMusicVideoRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IMusicVideoRepository>() 
            ?? throw new InvalidOperationException("IMusicVideoRepository not registered");

        [Key]
        public long Id { get; set; }
        public string MvVid { get; set; } = string.Empty;
        public string MvContent { get; set; } = string.Empty;
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public DateTime InsertDate { get; set; }

        // 通过MusicVideoID合成App消息
        public static async Task<AppMessage> GetAppMessageAsync(string mvVid)
        {
            return new AppMessage
            {
                Content = await Repository.GetContentByVidAsync(mvVid)
            }; 
        }

        // 收到MV分享消息时添加到MV库
        public static async Task<int> HandleAppAsync(long botUin, long groupId, long userId, AppMessage message)
        {
            // ShowMessage($"content:{message.Content}");
            int i = 0;
            string mv_vid = GetVid(message.Content);
            if (mv_vid != "")
            {
                if (await ExistsMvAsync(message.Content))
                {
                    // ShowMessage("此MV已存在", ConsoleColor.DarkGreen);
                }
                else
                {
                    i = await AppendAsync(mv_vid, message.Content, groupId, userId);
                    if (i == -1)
                        Logger.Error("添加MV失败");
                    else
                    {
                        // ShowMessage($"✅ 添加MV成功！\nVID：{mv_vid}\nMV数量：{CountAsync}", ConsoleColor.DarkRed);
                    }
                }
            }
            return i;
        }

        public static async Task<bool> ExistsMvAsync(string content)
        {
            return await Repository.ExistsByVidAsync(GetVid(content));
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

        public static async Task<int> AppendAsync(string mvVid, string mvContent, long groupId, long userId)
        {
            return await Repository.AddAsync(new MusicVideo
            {
                MvVid = mvVid,
                MvContent = mvContent,
                GroupId = groupId,
                UserId = userId,
                InsertDate = DateTime.Now
            });
        }
    }
}
