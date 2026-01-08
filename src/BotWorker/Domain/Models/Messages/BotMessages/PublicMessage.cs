using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Communication.Platforms.BotPublic;
using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;
using System.Diagnostics;
using BotWorker.Domain.Constants;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 处理公众号消息
        public async Task<string> HandlePublicMessage(string robotKey, string clientKey, bool isVoice = false)
        {  
            var botPublic = await BotPublic.LoadAsync(robotKey);  
            if (botPublic != null)
            {
                SelfInfo = await BotInfo.LoadAsync(botPublic.BotUin) ?? new();                
                Group = await GroupInfo.LoadAsync(botPublic.GroupId) ?? new();
                User = await UserInfo.LoadAsync(ClientPublic.GetUserId(robotKey, clientKey)) ?? new();
                RealGroupId = Group.Id;
                var lastMsgId = GroupMember.Get<string>("LastMsgId", GroupId, UserId);
                if (MsgId == lastMsgId)
                {
                    var token = Token.GetToken(UserId);
                    var url = $"{C.url}/ai?t={token}&gid={GroupId}msgid={MsgId}";
                    Answer = $"已超时请前往\n<a href=\"{url}\">网站后台</a>查看结果\n你的TOKEN：{token}";
                }
                else
                {
                    if (Message.Contains("领积分") && !ClientPublic.IsBind(UserId))
                    {
                        Answer = $"TOKEN:MP{ClientPublic.GetBindToken(robotKey, clientKey)}\n复制此消息发给QQ机器人即可得分";
                        GroupSendMessage.Append(this);
                    }
                    else if (Message == "邀请码")
                    {
                        Answer = $"邀请码：{ClientPublic.InviteCode(robotKey, clientKey)}\n公众号留言此邀请码您与邀请人均可获得5000积分";
                        GroupSendMessage.Append(this);
                    }
                    else if (Message.IsMatch(ClientPublic.regexRec))
                    {
                        Answer = ClientPublic.GetRecRes(SelfId, GroupId, GroupName, UserId, Name, robotKey, clientKey, Message);
                        GroupSendMessage.Append(this);
                    }
                    else
                    {
                        CurrentStopwatch = Stopwatch.StartNew();
                        await HandleEventAsync();
                        CurrentStopwatch.Stop();
                        CostTime = CurrentStopwatch.Elapsed.TotalSeconds;                        
                        GroupSendMessage.Append(this);
                    }
                }

                if (AddGroupMember() != -1)
                {
                    GroupMember.UpdateWhere($"LastMsgId={MsgId.Quotes()}, LastTime=GETDATE()", $"GroupId={GroupId} AND UserId = {UserId}");
                }

                //音乐消息处理
                if (Music.ExistsSong(Answer))
                    Answer = Music.GetSongUrlPublic(Music.GetSong(Answer));

                //转为微信表情
                Answer = FacePublic.ConvertFacesBack(Answer);

                //公众号最多返回 2047 字节数
                if (Answer.Length > 681)
                    Answer = Answer[..681];
            }

            return IsSend ? Answer : "";
        }
    }
}
