namespace BotWorker.Domain.Models.Messages.BotMessages;

public partial class BotMessage : MetaData<BotMessage>
{
        public async Task<(bool, bool)> HandleGuildMessageAsync()
        {
            var isNewGroup = false;

            //官机群及群成员号码处理
            if (IsGuild)
            {
                if (SelfInfo.BotUin == 0)
                    SelfInfo = await BotInfo.LoadAsync(UserGuild.GetUserId(SelfId, SelfId.ToString(), GroupOpenid)) ?? new();
                else
                    SelfInfo = await BotInfo.LoadAsync(SelfInfo.BotUin) ?? new();
                User = await UserInfo.LoadAsync(UserGuild.GetUserId(SelfId, UserOpenId, GroupOpenid)) ?? new();
                if (!GroupOpenid.IsNull())
                {
                    (RealGroupId, isNewGroup) = GroupOffical.GetGroupId(GroupOpenid, GroupName, UserId, SelfId, SelfName);
                    GroupMember.Append(RealGroupId, UserId, Name);
                    Group = await GroupInfo.LoadAsync(RealGroupId) ?? new();
                }
                if (GuildId.IsNull()) 
                    Group.GroupName = GroupOpenid.MaskNo();
                if (GroupId < GroupInfo.groupMin)
                {
                    var groupName = GroupInfo.GetValue("GroupName", GroupId);
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
                    isbot = isbot || UserId == bot || UpdateTargetGroupAndTargetQQ(bot);
                }
            }
            catch (Exception ex)
            {
                InfoMessage($"{ex.Message}");
            }
            return (isNewGroup, isbot);
        }

        public bool UpdateTargetGroupAndTargetQQ(long botUin)
        {            
            if (Message.IsNull()) return false;
            if (Message.Contains($":{botUin}"))
            {
                var sourceQQ = UserInfo.GetSourceQQ(botUin, UserId);
                if (sourceQQ == 0)
                {
                    _ = Task.Run(() =>
                    {
                        var sql = IsPostgreSql ? "CALL sp_UpdateUserIds()" : "EXEC sz84_robot.DBO.sp_UpdateUserIds";
                        Exec(sql);
                    });
                }
                else
                {
                    var sourceGroupId = GroupInfo.GetSourceGroupId(botUin, GroupId);
                    if (sourceGroupId == 0)
                    {
                        _ = Task.Run(async () =>
                        {
                            await Task.Delay(20000).ConfigureAwait(false);
                            Message = Message.RemoveUserId(botUin).Trim();
                            var sql = $"SELECT {SqlTop(1)} GroupId FROM {GroupSendMessage.FullName} WHERE ABS({SqlDateDiff("SECOND", "InsertDate", SqlDateTime)}) < 40 " +
                                      $"AND LTRIM(RTRIM(question)) = {Message.Quotes()} AND GroupId > {UserGuild.MIN_USER_ID} AND GroupId != {GroupId} AND UserId = {UserId} ORDER BY Id DESC {SqlLimit(1)}";
                            sourceGroupId = QueryScalar<long>(sql);
                            if (sourceGroupId != 0)
                            {
                                sql = IsPostgreSql ? $"CALL sp_UpdateGroupInfo({sourceGroupId}, {GroupId})" : $"EXEC sz84_robot.DBO.sp_UpdateGroupInfo {sourceGroupId}, {GroupId} ";
                                Exec(sql);
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
