using System.Diagnostics;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage
    {
        // 处理公众号消息
        public async Task<string> HandlePublicMessage(string robotKey, string clientKey, bool isVoice = false)
        {  
            var botPublic = await BotPublicRepository.GetByPublicKeyAsync(robotKey);  
            if (botPublic != null)
            {
                SelfInfo = await BotRepository.GetByBotUinAsync(botPublic.BotUin) ?? new();                
                Group = await GroupRepository.GetByIdAsync(botPublic.GroupId) ?? new();
                User = await UserRepository.GetByIdAsync(await PublicUserRepository.GetUserIdAsync(robotKey, clientKey)) ?? new();
                RealGroupId = Group.Id;
                var lastMsgId = await GroupMemberRepository.GetValueAsync<string>("LastMsgId", GroupId, UserId);
                if (MsgId == lastMsgId)
                {
                    var token = await ServiceProvider.GetRequiredService<ITokenRepository>().GetTokenByUserIdAsync(UserId);
                    var url = $"{_url}/ai?t={token}&gid={GroupId}msgid={MsgId}";
                    Answer = $"已超时请前往\n<a href=\"{url}\">网站后台</a>查看结果\n你的TOKEN：{token}";
                }
                else
                {
                    if (Message.Contains("领积分") && !await PublicUserRepository.IsBindAsync(UserId))
                    {
                        Answer = $"TOKEN:MP{await PublicUserRepository.GetBindTokenAsync(robotKey, clientKey)}\n复制此消息发给QQ机器人即可得分";
                        await GroupSendMessageRepository.AppendAsync(this);
                    }
                    else if (Message == "邀请码")
                    {
                        Answer = $"邀请码：{await PublicUserRepository.GetInviteCodeAsync(robotKey, clientKey)}\n公众号留言此邀请码您与邀请人均可获得5000积分";
                        await GroupSendMessageRepository.AppendAsync(this);
                    }
                    else if (Message.IsMatch(@"^推荐[人]*(qq|q|号码)*[ :：　+＋十]*(?<recQQ>[1-9]\d{4,10})?"))
                    {
                        Answer = await GetRecResAsync(SelfId, GroupId, GroupName, UserId, Name, robotKey, clientKey, Message);
                        await GroupSendMessageRepository.AppendAsync(this);
                    }
                    else
                    {
                        CurrentStopwatch = Stopwatch.StartNew();
                        await HandleEventAsync();
                        CurrentStopwatch.Stop();
                        CostTime = CurrentStopwatch.Elapsed.TotalSeconds;                        
                        await GroupSendMessageRepository.AppendAsync(this);
                    }
                }

                if (await AddGroupMemberAsync() != -1)
                {
                    await GroupMemberRepository.UpdateAsync($"last_msg_id={MsgId.Quotes()}, last_time={SqlDateTime}", GroupId, UserId);
                }

                //音乐消息处理
                if (Music.ExistsSong(Answer))
                {
                    var song = Music.GetSong(Answer);
                    Answer = await MusicRepository.GetMusicUrlPublicAsync(Music.GetMusicKind(song.Kind), song.SongId);
                }

                //转为微信表情
                Answer = FaceRepository.ConvertFacesBack(Answer);

                //公众号最多返回 2047 字节数
                if (Answer.Length > 681)
                    Answer = Answer[..681];
            }

            return IsSend ? Answer : "";
        // 公众号推荐人处理逻辑
        public async Task<string> GetRecResAsync(long botUin, long groupId, string groupName, long userId, string name, string botKey, string clientKey, string message)
        {
            if (!await PublicUserRepository.IsBindAsync(userId))
                return "请先发送【领积分】完成积分任务";

            var clientOid = await PublicUserRepository.GetUserIdAsync(botKey, clientKey);
            string recUserIdStr = await PublicUserRepository.GetValueAsync<string>("RecUserId", clientOid);
            
            const string format = "格式：推荐人+对方QQ\n例如：\n推荐人：{客服QQ}";
            const string regexRec = @"^推荐[人]*(qq|q|号码)*[ :：　+＋十]*(?<recQQ>[1-9]\d{4,10})?";

            if (!string.IsNullOrEmpty(recUserIdStr))
                return $"推荐人已登记为：{recUserIdStr}\n{format}";

            string recUserIdInput = message.RegexGetValue(regexRec, "RecUserId");

            if (recUserIdInput.IsNull())
                return format;

            if (recUserIdInput == userId.ToString())
                return "推荐人不能是自己";

            long recUserId = recUserIdInput.AsLong();
            
            // Check if recUserId exists in the same bot context
            var recUserExists = await PublicUserRepository.CountAsync("WHERE BotKey = @botKey AND UserId = @recUserId", new { botKey, recUserId }) > 0;
            if (!recUserExists)
                return "此号码未登记，请确认号码正确";

            if (await BlackListRepository.IsSystemBlackAsync(recUserId))
                return "此号码已被列入官方黑名单";

            var robotOwner = await GroupRepository.GetRobotOwnerAsync(groupId);
            if (await UserRepository.AppendAsync(botUin, groupId, userId, name, robotOwner) == -1)
                return RetryMsg;

            long creditAdd = 5000;

            using var transWrapper = await BeginTransactionAsync();
            var trans = transWrapper.Transaction;
            try
            {
                // 1. 更新推荐人
                await PublicUserRepository.UpdateAsync($"RecUserId = {recUserId}, BindCredit = BindCredit + 5000", clientOid, trans);

                // 2. 自己加分
                var addResSelf = await UserService.AddCreditAsync(botUin, groupId, groupName, userId, name, creditAdd, "推荐关注", trans);
                if (addResSelf.Result == -1) throw new Exception("自己加分失败");

                // 3. 推荐人加分
                var addResRec = await UserService.AddCreditAsync(botUin, groupId, groupName, recUserId, "", creditAdd, $"推荐关注:{userId}", trans);
                if (addResRec.Result == -1) throw new Exception("推荐人加分失败");

                await transWrapper.CommitAsync();

                // 同步缓存 (If needed, though UserInfo.SyncCacheField was static)
                // UserInfo.SyncCacheField(userId, groupId, "Credit", addResSelf.CreditValue);
                // UserInfo.SyncCacheField(recUserId, groupId, "Credit", addResRec.CreditValue);

                return $"推荐人登记为：\n{recUserId} +{creditAdd}分，累计：{addResRec.CreditValue}\n您的积分：{creditAdd}，累计：{addResSelf.CreditValue}";
            }
            catch (Exception ex)
            {
                await transWrapper.RollbackAsync();
                Console.WriteLine($"[GetRecRes Error] {ex.Message}");
                return RetryMsg;
            }
        }
        // 公众号绑定 Token 处理逻辑
        public async Task<string> GetBindTokenAsync(string tokenType, string bindToken)
        {
            if (string.IsNullOrWhiteSpace(bindToken)) return "";
            
            if (tokenType == "MP")
            {
                if (IsPublic)
                    return "请用QQ发消息领取，体验群：6433316";

                // Check if already bound
                var alreadyBound = await PublicUserRepository.CountAsync("WHERE BotKey = (SELECT BotKey FROM PublicUser WHERE BindToken = @bindToken) AND UserId = @UserId", new { bindToken, UserId }) > 0;
                if (alreadyBound)
                    return "您已领过积分，不能再次领取";

                var res = await PublicUserRepository.GetValueAsync<string>("UserId", "WHERE BindToken = @bindToken", new { bindToken });
                if (string.IsNullOrEmpty(res))
                    return "此TOKEN无效，请确认后再试";

                long bindUserId = res.AsLong();
                if (await PublicUserRepository.IsBindAsync(bindUserId))
                    return "此TOKEN已失效，请重新获取";

                await AddClientAsync();

                long creditAdd = 5000;
                
                using var transWrapper = await BeginTransactionAsync();
                var trans = transWrapper.Transaction;
                try
                {
                    // 1. 更新绑定信息
                    await PublicUserRepository.UpdateAsync($"UserId = {UserId}, IsBind = 1, BindDate = CURRENT_TIMESTAMP, BindCredit = {creditAdd}", $"BindToken = {bindToken.Quotes()}", null, trans);

                    // 2. 更新积分记录关联
                    await ExecAsync(trans, $"UPDATE CreditLog SET UserId = {UserId} WHERE UserId = {bindUserId}");

                    // 3. 更新发送消息记录关联
                    await ExecAsync(trans, $"UPDATE GroupSendMessage SET UserId = {UserId} WHERE UserId = {bindUserId}");

                    // 4. 处理 Token 表
                    var tokenRepo = ServiceProvider.GetRequiredService<ITokenRepository>();
                    if (await tokenRepo.ExistsTokenAsync(UserId, "")) // Need better ExistsToken
                    {
                         await ExecAsync(trans, $"DELETE FROM Token WHERE UserId = {bindUserId}");
                    }
                    else
                    {
                         await ExecAsync(trans, $"UPDATE Token SET UserId = {UserId} WHERE UserId = {bindUserId}");
                    }

                    // 5. 增加积分
                    var addRes = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "关注公众号领积分", trans);
                    if (addRes.Result == -1) throw new Exception("增加积分失败");

                    await transWrapper.CommitAsync();

                    return $"✅ 领取成功!\n获得积分：{creditAdd}\n当前积分：{addRes.CreditValue}";
                }
                catch (Exception ex)
                {
                    await transWrapper.RollbackAsync();
                    Console.WriteLine($"[GetBindToken Error] {ex.Message}");
                    return RetryMsg;
                }
            }
            return "";
        }
    }
}
