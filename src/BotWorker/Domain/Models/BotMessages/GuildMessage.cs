namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<(bool, bool)> HandleGuildMessageAsync()
        {
            var isNewGroup = false;

            //官机群及群成员号码处理
            if (IsGuild)
            {
                if (SelfInfo.BotUin == 0)
                    SelfInfo = await BotRepository.GetByIdAsync(await UserService.GetUserIdAsync(SelfId, SelfId.ToString(), GroupOpenid)) ?? new();
                else
                    SelfInfo = await BotRepository.GetByIdAsync(SelfInfo.BotUin) ?? new();

                User = await UserRepository.GetByIdAsync(await UserService.GetUserIdAsync(SelfId, UserOpenId, GroupOpenid)) ?? new();

                if (!GroupOpenid.IsNull())
                {
                    (RealGroupId, isNewGroup) = await GroupService.GetGroupIdAsync(GroupOpenid, GroupName, UserId, SelfId, SelfName);
                    await GroupMemberRepository.AppendAsync(RealGroupId, UserId, Name);
                    Group = await GroupRepository.GetByIdAsync(RealGroupId) ?? new();
                }

                if (GuildId.IsNull()) 
                    Group.GroupName = GroupOpenid.MaskNo();

                if (GroupId < 1000000) // 假设 groupMin 是 1000000
                {
                    var groupName = await GroupRepository.GetValueAsync<string>("GroupName", GroupId);
                    if (groupName != GroupName)
                        Group.GroupName = groupName;
                }
                Message = Message.RegexReplace(@"<faceType=\d+,faceId=""(\d+)"",ext=""[^""]*"">", m => $"[face{m.Groups[1].Value}.gif]");
                Message = Message.RegexReplace(@"<@![^>]*>", "").Trim(); //频道机器人有@数据，群聊没用？                
                Message = Message.ConvertEmojiToFace();
            }

            //自动处理 TargetGroup TargetUserId
            var isbot = false;
            try
            {
                foreach (var bot in OfficalBots)
                {
                    isbot = isbot || UserId == bot || await UpdateTargetGroupAndTargetQQAsync(bot);
                }
            }
            catch (Exception ex)
            {
                Logger.Error($"{ex.Message}");
            }
            return (isNewGroup, isbot);
        }

        public async Task<bool> UpdateTargetGroupAndTargetQQAsync(long botUin)
        {            
            if (Message.IsNull()) return false;
            if (Message.Contains($":{botUin}"))
            {
                var sourceQQ = await UserRepository.GetSourceQQAsync(botUin, UserId);
                if (sourceQQ == 0)
                {
                    _ = Task.Run(async () =>
                    {
                        var sql = "CALL sp_UpdateUserIds()";
                        await ExecAsync(sql);
                    });
                }
                else
                {
                    var sourceGroupId = await GroupRepository.GetSourceGroupIdAsync(botUin, GroupId);
                    if (sourceGroupId == 0)
                    {
                        _ = Task.Run(async () =>
                        {
                            await Task.Delay(20000).ConfigureAwait(false);
                            Message = Message.RemoveUserId(botUin).Trim();
                            
                            // 这里暂时保留 raw SQL，因为涉及多表复杂查询，但改用 ExecAsync/QueryScalarAsync
                            string sql = $@"
                                SELECT GroupId 
                                FROM SendMessage 
                                WHERE ABS(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - InsertDate))) < 40 
                                AND LTRIM(RTRIM(question)) = @message 
                                AND GroupId > 980000000000 
                                AND GroupId != @groupId 
                                AND UserId = @userId 
                                ORDER BY Id DESC 
                                LIMIT 1";

                            using var conn = Persistence.Database.DbProviderFactory.CreateConnection();
                            if (conn is System.Data.Common.DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
                            sourceGroupId = await conn.ExecuteScalarAsync<long>(sql, new { message = Message.Trim(), groupId = GroupId, userId = UserId });

                            if (sourceGroupId != 0)
                            {
                                var sqlUpdate = $"CALL sp_UpdateGroupInfo({sourceGroupId}, {GroupId})";
                                await conn.ExecuteAsync(sqlUpdate);
                            }
                        });
                    }
                }
                if (IsReply && Message.Contains("撤回"))                
                    return false;                
                else
                {
                    Reason += "[艾特官机]";
                    return true;
                }
            }
            return false;
        }
}
