using System;
using System.Collections.Generic;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence.Repositories;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class GroupRepository : BaseRepository<GroupInfo>, IGroupRepository
    {
        public GroupRepository(string? connectionString = null) 
            : base("group_info", connectionString ?? GlobalConfig.BaseInfoConnection)
        {
        }

        public async Task<GroupInfo?> GetByOpenIdAsync(string openId, long botUin)
        {
            string sql = $"SELECT * FROM {_tableName} WHERE group_open_id = @openId AND bot_uin = @botUin";
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<GroupInfo>(sql, new { openId, botUin });
        }

        public async Task<long> AddAsync(GroupInfo group)
        {
            return await InsertAsync(group);
        }

        public async Task<bool> UpdateAsync(GroupInfo group)
        {
            group.UpdatedAt = DateTime.Now;
            return await UpdateEntityAsync(group);
        }

        public async Task<long> GetGroupOwnerAsync(long groupId, long def = 0, System.Data.IDbTransaction? trans = null)
        {
            var result = await GetValueAsync<long?>("group_owner", groupId, trans);
            return result ?? def;
        }

        public async Task<bool> GetIsCreditAsync(long groupId)
        {
            return await GetValueAsync<bool>("is_credit", groupId);
        }

        public async Task<bool> GetIsPetAsync(long groupId)
        {
            return await GetValueAsync<bool>("is_pet", groupId);
        }

        public async Task<int> SetPowerOnAsync(long groupId, System.Data.IDbTransaction? trans = null)
        {
            return await SetValueAsync("IsPowerOn", true, groupId, trans);
        }

        public async Task<int> SetPowerOffAsync(long groupId, System.Data.IDbTransaction? trans = null)
        {
            return await SetValueAsync("IsPowerOn", false, groupId, trans);
        }

        public async Task<int> StartCyGameAsync(int state, string lastChengyu, long groupId)
        {
            string sql = $"UPDATE {_tableName} SET is_in_game = @state, last_chengyu = @lastChengyu WHERE id = @groupId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { state, lastChengyu, groupId });
        }

        public async Task<int> GetChengyuIdleMinutesAsync(long groupId)
        {
            // Calculate in memory to avoid DB-specific SQL for date diff
            var lastDate = await GetValueAsync<DateTime?>("last_chat_date", groupId);
            if (lastDate == null || lastDate == DateTime.MinValue) return 999999;
            return (int)(DateTime.Now - lastDate.Value).TotalMinutes;
        }

        public async Task<bool> GetPowerOnAsync(long groupId, System.Data.IDbTransaction? trans = null)
        {
            return await GetValueAsync<bool>("IsPowerOn", groupId, trans);
        }

        public async Task<int> SetRobotOwnerAsync(long groupId, long ownerId, System.Data.IDbTransaction? trans = null)
        {
            return await SetValueAsync("RobotOwner", ownerId, groupId, trans);
        }

        public async Task<long> GetRobotOwnerAsync(long groupId, long def = 0, System.Data.IDbTransaction? trans = null)
        {
            var res = await GetValueAsync<long?>("RobotOwner", groupId, trans);
            return res ?? def;
        }

        public async Task<bool> IsOwnerAsync(long groupId, long userId, System.Data.IDbTransaction? trans = null)
        {
            return userId == await GetRobotOwnerAsync(groupId, 0, trans);
        }

        public async Task<bool> IsPowerOffAsync(long groupId, System.Data.IDbTransaction? trans = null)
        {
            return !await GetPowerOnAsync(groupId, trans);
        }

        public async Task<bool> GetIsValidAsync(long groupId, System.Data.IDbTransaction? trans = null)
        {
            return await GetValueAsync<bool>("IsValid", groupId, trans);
        }

        public async Task<string> GetRobotOwnerNameAsync(long groupId, string botName = "")
        {
            string res = await GetValueAsync<string>("RobotOwnerName", groupId) ?? "";
            if (string.IsNullOrEmpty(res))
            {
                res = (await GetRobotOwnerAsync(groupId)).ToString();
                res = $"[@:{res}]";
            }
            return res;
        }

        public async Task<bool> IsCanTrialAsync(long groupId)
        {
            if (await GroupVip.IsVipOnceAsync(groupId))
                return false;

            string sql = $"SELECT ABS(DATE_PART('day', NOW() - trial_start_date)) FROM {_tableName} WHERE id = @id";
            using var conn = CreateConnection();
            int days = await conn.ExecuteScalarAsync<int>(sql, new { id = groupId });
            
            if (days >= 180)
            {
                string updateSql = $"UPDATE {_tableName} SET is_valid = true, trial_start_date = NOW(), trial_end_date = NOW() + INTERVAL '7 days' WHERE id = @id";
                await conn.ExecuteAsync(updateSql, new { id = groupId });
                return true;
            }
            return await GetIsValidAsync(groupId);
        }

        public async Task<int> SetInvalidAsync(long groupId, string groupName = "", long groupOwner = 0, long robotOwner = 0)
        {
            await AppendAsync(groupId, groupName, BotInfo.BotUinDef, BotInfo.BotNameDef, groupOwner, robotOwner);
            if (await GroupVip.IsVipAsync(groupId))
                return -1;
            else
                return await SetValueAsync("IsValid", false, groupId);
        }

        public async Task<int> SetHintDateAsync(long groupId)
        {
             return await SetValueAsync("LastExitHintDate", DateTime.Now, groupId);
        }

        public async Task<bool> GetIsWhiteAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsWhite", groupId);
        }

        public async Task<string> GetIsBlockResAsync(long groupId)
        {
            return await GetIsBlockAsync(groupId) ? "已开启" : "已关闭";
        }

        public async Task<bool> GetIsBlockAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsBlock", groupId);
        }

        public async Task<int> GetIsOpenAsync(long groupId)
        {
            return await GetValueAsync<int>("IsOpen", groupId);
        }

        public async Task<int> GetLastHintTimeAsync(long groupId)
        {
            string sql = $"SELECT ABS(EXTRACT(EPOCH FROM (last_exit_hint_date - NOW()))) FROM {_tableName} WHERE id = @id";
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<int>(sql, new { id = groupId });
        }

        public async Task<int> CloudAnswerAsync(long groupId)
        {
            return await GetValueAsync<int>("IsCloudAnswer", groupId);
        }

        public async Task<string> CloudAnswerResAsync(long groupId)
        {
            List<string> answers = ["闭嘴", "本群", "官方", "话痨", "终极", "AI"];
            int index = await CloudAnswerAsync(groupId);
            if (index >= 0 && index < answers.Count)
                return answers[index];
            else
                return string.Empty;
        }

        public async Task<bool> GetIsBlackExitAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsBlackExit", groupId);
        }

        public async Task<bool> GetIsBlackKickAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsBlackKick", groupId);
        }

        public async Task<string> GetClosedFuncAsync(long groupId)
        {
            string cmdSql = "SELECT cmd_name FROM bot_cmd"; 
            using var conn = CreateConnection();
            var cmdNames = await conn.QueryAsync<string>(cmdSql);
            
            string closeRegex = await GetValueAsync<string>("CloseRegex", groupId) ?? "";
            if (string.IsNullOrEmpty(closeRegex)) return "";

            StringBuilder sb = new("\n已关闭：");
            string pattern = @"(?<CmdName>" + closeRegex.Replace(" ", "|").Trim() + ")";
             foreach (var cmdName in cmdNames)
            {
                if (Regex.IsMatch(cmdName, pattern))
                {
                    sb.Append($"{cmdName} ");
                }
            }
            return sb.ToString();
        }

        public async Task<string> GetClosedRegexAsync(long groupId)
        {
            string res = await GetValueAsync<string>("CloseRegex", groupId) ?? "";
            if (res != "")
                res = @"^[#＃﹟]{0,1}(?<cmd>(" + res.Trim().Replace(" ", "|") + @"))[+]*(?<cmdPara>[\s\S]*)";
            return res;
        }

        public async Task<bool> GetIsExitHintAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsExitHint", groupId);
        }

        public async Task<bool> GetIsKickHintAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsKickHint", groupId);
        }

        public async Task<bool> GetIsRequirePrefixAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsRequirePrefix", groupId);
        }

        public async Task<string> GetJoinResAsync(long groupId)
        {
            int joinRes = await GetValueAsync<int>("IsAcceptNewmember", groupId);
            return joinRes switch
            {
                0 => "拒绝",
                1 => "同意",
                2 => "忽略",
                _ => "未设置",
            };
        }

        public async Task<string> GetSystemPromptAsync(long groupId)
        {
            return await GetValueAsync<string>("SystemPrompt", groupId) ?? "";
        }

        public async Task<string> GetAdminRightResAsync(long groupId)
        {
            int adminRight = await GetValueAsync<int>("AdminRight", groupId);
            return adminRight switch
            {
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "主人",
            };
        }

        public async Task<string> GetRightResAsync(long groupId)
        {
            return (await GetIsOpenAsync(groupId)) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "已关闭",
            };
        }

        public async Task<string> GetTeachRightResAsync(long groupId)
        {
             return (await GetValueAsync<int>("TeachRight", groupId)) switch
            {
                1 => "所有人",
                2 => "管理员",
                3 => "白名单",
                4 => "主人",
                _ => "",
            };
        }

        public async Task<int> SetInGameAsync(int isInGame, long groupId)
        {
            return await SetValueAsync("IsInGame", isInGame, groupId);
        }

        public async Task<string> GetWelcomeResAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsWelcomeHint", groupId) ? "发送" : "不发送";
        }

        public async Task<string> GetGroupNameAsync(long groupId)
        {
            return await GetValueAsync<string>("GroupName", groupId) ?? "";
        }

        public async Task<string> GetGroupOwnerNicknameAsync(long groupId)
        {
            return await GetValueAsync<string>("GroupOwnerNickname", groupId) ?? "";
        }

        public async Task<bool> GetIsAIAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsAI", groupId);
        }

        public async Task<bool> GetIsOwnerPayAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsOwnerPay", groupId);
        }

        public async Task<int> GetContextCountAsync(long groupId)
        {
            return await GetValueAsync<int>("ContextCount", groupId);
        }

        public async Task<bool> GetIsMultAIAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsMultAI", groupId);
        }

        public async Task<bool> GetIsUseKnowledgebaseAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsUseKnowledgebase", groupId);
        }

        public async Task<int> AppendAsync(long groupId, string name, long selfId, string selfName, long groupOwner = 0, long robotOwner = 0, string openid = "")
        {
            var group = await GetByIdAsync(groupId);
            if (group != null)
            {
                if (!string.IsNullOrEmpty(name)) 
                {
                    group.GroupName = name;
                }
                
                if (groupOwner != 0 && await GetGroupOwnerAsync(groupId) == 0 && !await GroupVip.IsVipAsync(groupId))
                {
                    group.GroupOwner = groupOwner;
                }

                if (robotOwner != 0 && await GetRobotOwnerAsync(groupId) == 0 && !await GroupVip.IsVipAsync(groupId))
                {
                    group.RobotOwner = robotOwner;
                }

                group.BotUin = selfId;
                group.LastDate = DateTime.Now; 
                
                await UpdateEntityAsync(group);
                return 1;
            }
            else
            {
                var newGroup = new GroupInfo
                {
                    Id = groupId,
                    GroupOpenId = openid,
                    GroupName = name,
                    GroupOwner = groupOwner,
                    RobotOwner = robotOwner,
                    BotUin = selfId,
                    BotName = selfName,
                    InsertDate = DateTime.Now,
                    LastDate = DateTime.Now
                };
                await InsertAsync(newGroup);
                return 1;
            }
        }

        public async Task<bool> GetIsNoLogAsync(long groupId) => await GetValueAsync<bool>("IsNoLog", groupId);
        public async Task<bool> GetIsNoCheckAsync(long groupId) => await GetValueAsync<bool>("IsNoCheck", groupId);
        public async Task<bool> GetIsHintCloseAsync(long groupId) => await GetValueAsync<bool>("IsHintClose", groupId);
        public async Task<long> GetSourceGroupIdAsync(long groupId) => await GetValueAsync<long>("SourceGroupId", groupId);
        
        public async Task<int> UpdateGroupAsync(long group, string name, long selfId, long groupOwner = 0, long robotOwner = 0)
        {
            var entity = await GetByIdAsync(group);
            if (entity == null) return 0;
            
            if (!string.IsNullOrEmpty(name)) entity.GroupName = name;
            
            if (groupOwner != 0 && await GetGroupOwnerAsync(group) == 0 && !await GroupVip.IsVipAsync(group))
                entity.GroupOwner = groupOwner;
                
            if (robotOwner != 0 && await GetRobotOwnerAsync(group) == 0 && !await GroupVip.IsVipAsync(group))
                entity.RobotOwner = robotOwner;
                
            entity.BotUin = selfId;
            entity.LastDate = DateTime.Now;
            
            return await UpdateEntityAsync(entity) ? 1 : 0;
        }

        public async Task<long> GetSourceGroupIdAsync(long botUin, long groupId) 
        {
             return await GetSourceGroupIdAsync(groupId);
        }

        public async Task<int> SetIsOpenAsync(bool isOpen, long groupId)
        {
            return await SetValueAsync("IsOpen", isOpen, groupId);
        }

        public async Task<int> SetPowerOnAsync(bool isOpen, long groupId)
        {
            return await SetValueAsync("IsPowerOn", isOpen, groupId);
        }

        public async Task<bool> GetPowerOnAsync(long groupId)
        {
            return await GetValueAsync<bool>("IsPowerOn", groupId);
        }

        public async Task<string> GetSystemPromptStatusAsync(long groupId)
        {
            string prompt = await GetSystemPromptAsync(groupId);
            if (string.IsNullOrEmpty(prompt)) prompt = "未设置";
            return $"📌 设置系统提示词\n内容：\n{prompt}";
        }

        public async Task<string> GetVipResAsync(long groupId)
        {
             string version;
             string res;

             if (await GroupVip.ExistsAsync(groupId))
             {
                 if (await GroupVip.IsYearVIPAsync(groupId))
                     version = "年费版";
                 else
                     version = "VIP版";
                 int valid_days = await GroupVip.RestDaysAsync(groupId);
                 if (valid_days >= 1850)
                     res = "『永久版』";
                 else
                     res = $"『{version}』有效期：{valid_days}天";
             }
             else
             {
                 if (await GroupVip.IsVipOnceAsync(groupId))
                     return "已过期，请及时续费";
                 else
                     version = "体验版";
                 res = $"『{version}』";
             }
             return res;
        }

        public async Task<int> StartCyGameAsync(long groupId)
        {
            return await SetValueAsync("IsCyGame", true, groupId);
        }

        public async Task<int> UpdateIsPowerOnAsync(long groupId, bool isPowerOn, System.Data.IDbTransaction? trans = null)
        {
            return await SetValueAsync("IsPowerOn", isPowerOn, groupId, trans);
        }

        public async Task<int> UpdateAdminRightAsync(long groupId, int adminRight)
        {
            return await SetValueAsync("AdminRight", adminRight, groupId);
        }

        public async Task<int> UpdateUseRightAsync(long groupId, int useRight)
        {
            return await SetValueAsync("UseRight", useRight, groupId);
        }

        public async Task<int> UpdateTeachRightAsync(long groupId, int teachRight)
        {
            return await SetValueAsync("TeachRight", teachRight, groupId);
        }

        public async Task<int> UpdateBlockMinAsync(long groupId, int blockMin)
        {
            return await SetValueAsync("BlockMin", blockMin, groupId);
        }

        public async Task<int> UpdateJoinGroupSettingsAsync(long groupId, int isAccept, string rejectMessage, string regexRequestJoin)
        {
            string sql = $"UPDATE {_tableName} SET is_accept_newmember = @isAccept, reject_message = @rejectMessage, regex_request_join = @regexRequestJoin WHERE id = @groupId";
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { isAccept, rejectMessage, regexRequestJoin, groupId });
        }

        public async Task<int> UpdateIsChangeHintAsync(long groupId, bool isChangeHint)
        {
            return await SetValueAsync("IsChangeHint", isChangeHint, groupId);
        }

        public async Task<int> UpdateWelcomeMessageAsync(long groupId, string message)
        {
            return await SetValueAsync("WelcomeMessage", message, groupId);
        }

        public async Task<int> UpdateIsWelcomeHintAsync(long groupId, bool isWelcomeHint)
        {
            return await SetValueAsync("IsWelcomeHint", isWelcomeHint, groupId);
        }

        public async Task<int> UpdateSystemPromptAsync(long groupId, string systemPrompt)
        {
            return await SetValueAsync("SystemPrompt", systemPrompt, groupId);
        }

        public async Task<int> UpdateReplyModeAsync(long groupId, int modeReply)
        {
            return await SetValueAsync("ReplyMode", modeReply, groupId);
        }

        public async Task<int> UpdateCloseRegexAsync(long groupId, string closeRegex)
        {
            return await SetValueAsync("CloseRegex", closeRegex, groupId);
        }
    }
}