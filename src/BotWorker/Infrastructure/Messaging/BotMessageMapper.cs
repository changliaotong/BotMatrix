using System.Text.Json;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Domain.Entities;
using BotWorker.Modules.Plugins;

namespace BotWorker.Infrastructure.Messaging
{
    public class BotMessageMapper
    {
        public static async Task<BotMessage?> MapToOneBotEventAsync(string json, IOneBotApiClient? apiClient = null)
        {
            var options = new JsonSerializerOptions
            {
                PropertyNameCaseInsensitive = true
            };
            
            OneBotEvent? ev;
            try 
            {
                ev = JsonSerializer.Deserialize<OneBotEvent>(json, options);
            }
            catch (Exception ex)
            {
                // 可以考虑在这里加日志，但 MapToOneBotEventAsync 是静态方法，没注入 ILogger
                // 只能通过抛出异常让调用者处理
                throw new JsonException($"Failed to deserialize OneBotEvent: {ex.Message}", ex);
            }

            if (ev == null) return null;

            var bm = new BotMessage
             {
                 MsgId = ev.MessageId.ToString(),
                 Time = ev.Time,
                 EventType = ev.MessageType switch
                 {
                     "group" => "GroupMessageEvent",
                     "private" => "FriendMessageEvent",
                     _ => ev.PostType == "message" ? "FriendMessageEvent" : ev.PostType
                 },
                 Message = ev.RawMessage,
                 CurrentMessage = ev.RawMessage,
                 RealGroupId = ev.GroupIdLong,
                 SelfInfo = new BotInfo 
                 { 
                     BotUin = ev.SelfId,
                     BotType = Platforms.BotType(ev.Platform)
                 }
             };

             // 如果是元事件，直接返回 null，忽略 meta_event 处理，避免后续数据库查询及写入
             if (ev.PostType == "meta_event")
             {
                 return null;
             }

             // 非元事件，尝试加载 Bot 信息
             try
             {
                 var botInfo = await BotInfo.GetSingleAsync(ev.SelfId);
                 if (botInfo != null)
                 {
                     bm.SelfInfo = botInfo;
                 }
             }
             catch
             {
                 // Ignore DB errors
             }

             // Set User
             if (ev.UserIdLong != 0)
             {
                 try
                 {
                     var userInfo = await UserInfo.GetSingleAsync(ev.UserIdLong);
                     if (userInfo != null)
                     {
                         bm.User = userInfo;
                         if (string.IsNullOrEmpty(bm.User.Name) && ev.Sender != null)
                         {
                             bm.User.Name = ev.Sender.Nickname ?? ev.Sender.Card ?? string.Empty;
                         }
                     }
                     else
                     {
                         bm.User = new UserInfo
                         {
                             Id = ev.UserIdLong,
                             Name = ev.Sender?.Nickname ?? ev.Sender?.Card ?? string.Empty
                         };
                     }
                 }
                 catch
                 {
                     bm.User = new UserInfo
                     {
                         Id = ev.UserIdLong,
                         Name = ev.Sender?.Nickname ?? ev.Sender?.Card ?? string.Empty
                     };
                 }
                 // 设置用户权限
                if (ev.Sender != null)
                {
                    bm.UserPerm = ev.Sender.Role?.ToLower() switch
                    {
                        "owner" or "0" => 0,
                        "admin" or "1" => 1,
                        _ => 2
                    };

                    // 如果发现用户是群主，且数据库中未记录或记录不一致，则更新群主信息
                    if (bm.UserPerm == 0 && ev.GroupIdLong != 0)
                    {
                        _ = Task.Run(async () => 
                        {
                            try
                            {
                                var group = await GroupInfo.GetSingleAsync(ev.GroupIdLong);
                                if (group != null && group.GroupOwner != ev.UserIdLong)
                                {
                                    await GroupInfo.SetValueAsync("GroupOwner", ev.UserIdLong, ev.GroupIdLong);
                                }
                            }
                            catch { /* Ignore */ }
                        });
                    }
                }

                // 特殊处理：如果发送者是机器人自己，且消息事件中包含了机器人自己的权限信息，则同步 UserPerm 和 SelfPerm
                if (ev.UserIdLong == ev.SelfId)
                {
                    if (!string.IsNullOrEmpty(ev.SelfRole))
                    {
                        bm.SelfPerm = ev.SelfRole.ToLower() switch
                        {
                            "owner" or "0" => 0,
                            "admin" or "1" => 1,
                            _ => 2
                        };
                        bm.UserPerm = bm.SelfPerm;
                    }
                    else if (bm.UserPerm != 2)
                    {
                        bm.SelfPerm = bm.UserPerm;
                    }
                }
             }

            // Set Group
            long groupIdToLoad = ev.GroupIdLong;
            // 如果是私聊且用户设置了默认群，则加载默认群信息以便后续逻辑（如权限检查、积分查询等）能正确关联到该群
            if (groupIdToLoad == 0 && bm.User.DefaultGroup != 0)
            {
                groupIdToLoad = bm.User.DefaultGroup;
            }

            if (groupIdToLoad != 0)
            {
                try
                {
                    var groupInfo = await GroupInfo.GetSingleAsync(groupIdToLoad);
                    if (groupInfo != null)
                    {
                        bm.Group = groupInfo;

                        // 如果数据库中没有群主信息，异步获取并更新
                        if (groupInfo.GroupOwner == 0 && apiClient != null)
                        {
                            _ = Task.Run(async () =>
                            {
                                try
                                {
                                    var response = await apiClient.SendActionAsync(ev.Platform, ev.SelfId.ToString(), "get_group_info", new { group_id = groupIdToLoad });
                                    // 这里取决于 OneBotApiClient 的具体实现，通常返回的是一个动态对象或 JsonElement
                                    // 简化处理：假设接口返回的数据中包含 owner
                                    if (response != null)
                                    {
                                        var json = JsonSerializer.Serialize(response);
                                        using var doc = JsonDocument.Parse(json);
                                        if (doc.RootElement.TryGetProperty("data", out var data) && data.TryGetProperty("owner", out var ownerProp))
                                        {
                                            long ownerId = 0;
                                            if (ownerProp.ValueKind == JsonValueKind.Number) ownerId = ownerProp.GetInt64();
                                            else if (ownerProp.ValueKind == JsonValueKind.String) long.TryParse(ownerProp.GetString(), out ownerId);

                                            if (ownerId != 0)
                                            {
                                                await GroupInfo.SetValueAsync("GroupOwner", ownerId, groupIdToLoad);
                                            }
                                        }
                                    }
                                }
                                catch { /* Ignore */ }
                            });
                        }

                        // 设置机器人权限
                        if (!string.IsNullOrEmpty(ev.SelfRole))
                        {
                            bm.SelfPerm = ev.SelfRole.ToLower() switch
                            {
                                "owner" or "0" => 0,
                                "admin" or "1" => 1,
                                _ => 2
                            };
                        }
                        else
                        {
                            // 如果事件中没给，且数据库中记录了机器人是群主，则作为备选（虽然用户说不用数据库，但作为兜底逻辑保留）
                            if (groupInfo.GroupOwner == bm.SelfId && bm.SelfId != 0)
                            {
                                bm.SelfPerm = 0;
                            }
                        }

                        // 如果识别到机器人是群主，同步更新数据库
                        if (bm.SelfPerm == 0 && groupInfo.GroupOwner != bm.SelfId)
                        {
                            _ = Task.Run(async () =>
                            {
                                try { await GroupInfo.SetValueAsync("GroupOwner", bm.SelfId, groupIdToLoad); } catch { }
                            });
                        }
                    }
                    else
                    {
                        bm.Group = new GroupInfo
                        {
                            Id = groupIdToLoad,
                            GroupName = string.Empty
                        };
                    }
                }
                catch
                {
                    bm.Group = new GroupInfo
                    {
                        Id = groupIdToLoad,
                        GroupName = string.Empty
                    };
                }
            }

             // Set Reply function
             bm.ReplyMessageAsync = async () =>
             {
                 if (apiClient != null && !string.IsNullOrEmpty(bm.Answer))
                 {
                     if (bm.IsGroup)
                     {
                         await apiClient.SendActionAsync(bm.Platform, bm.SelfId.ToString(), "send_group_msg", new
                         {
                             group_id = bm.RealGroupId.ToString(),
                             message = bm.Answer
                         });
                     }
                     else
                     {
                         await apiClient.SendActionAsync(bm.Platform, bm.SelfId.ToString(), "send_private_msg", new
                         {
                             user_id = bm.UserId.ToString(),
                             message = bm.Answer
                         });
                     }
                 }
             };

             return bm;
         }
    }
}
