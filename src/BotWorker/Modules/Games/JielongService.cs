using System;
using System.Threading.Tasks;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Enums;
using Microsoft.Extensions.Logging;
using BotWorker.Common;

namespace BotWorker.Modules.Games
{
    public class JielongService : IJielongService
    {
        private readonly IJielongRepository _repository;
        private readonly IUserRepository _userRepo;
        private readonly IGroupRepository _groupRepo;
        private readonly IChengyuService _chengyuService;
        private readonly ILogger<JielongService> _logger;

        public JielongService(
            IJielongRepository repository,
            IUserRepository userRepo,
            IGroupRepository groupRepo,
            IChengyuService chengyuService,
            ILogger<JielongService> logger)
        {
            _repository = repository;
            _userRepo = userRepo;
            _groupRepo = groupRepo;
            _chengyuService = chengyuService;
            _logger = logger;
        }

        public async Task<string> GetJielongResAsync(IPluginContext ctx, string cmdPara)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var name = ctx.UserName;
            var isGroup = ctx.GroupId != null;

            cmdPara = cmdPara.RemoveBiaodian().Trim();
            if (cmdPara == "结束")
            {
                if (await UserInGameAsync(groupId, userId, isGroup))
                {
                    var gameOverRes = await GameOverAsync(groupId, userId, isGroup);
                    return gameOverRes == -1
                        ? "操作失败，请重试"
                        : $"✅ 成语接龙游戏结束{await MinusCreditAsync(ctx)}";
                }
                return "";
            }

            bool inGame = await InGameAsync(groupId, userId);
            string currCy;
            string res;
            string creditInfo = "";
            if (!inGame)
            {
                if (cmdPara == "")
                    cmdPara = await CurrCyAsync(groupId, userId, isGroup);

                if (string.IsNullOrEmpty(cmdPara))
                {
                    cmdPara = (await _repository.GetRandomChengyuAsync())?.RemoveBiaodian() ?? "";
                }
                else if (await _chengyuService.GetOidAsync(cmdPara) == 0)
                {
                    var user = await _userRepo.GetAsync(userId);
                    return (user?.IsSuper == true || (user?.CreditTotal ?? 0) > 10000) ? $"【{cmdPara}】不是成语" : $"您输入的不是成语";
                }

                await AppendAsync(groupId, userId, name, cmdPara, 1);
                await StartAsync(groupId, userId, isGroup, cmdPara);
                currCy = cmdPara;
                creditInfo = await AddCreditAsync(ctx);
                res = $"✅ 成语接龙开始！";
            }
            else
            {
                currCy = await CurrCyAsync(groupId, userId, isGroup);
                string pinyin = await _chengyuService.PinYinAsync(currCy);
                cmdPara = cmdPara.RemoveQqAds();
                if (cmdPara == "")
                    return ctx.RawMessage.Contains("接龙") || ctx.RawMessage == ""
                        ? $"发【结束】退出游戏\n📌 请接：{currCy}\n🔤 拼音：{pinyin}"
                        : "";

                if (cmdPara == "提示")
                    return (await GetJielongAsync(groupId, userId, currCy)).MaskIdiom();

                if (await _chengyuService.GetOidAsync(cmdPara) == 0)
                {
                    if (isGroup && await _groupRepo.GetChengyuIdleMinutesAsync(groupId) > 10)
                    {
                        await _groupRepo.SetInGameAsync(0, groupId);
                        return "✅ 成语接龙超时自动结束";
                    }
                    return cmdPara.Length == 4 || ctx.RawMessage.StartsWith("接龙") || ctx.RawMessage.StartsWith("jl")
                        ? $"【{cmdPara}】不是成语\n💡 发【结束】退出游戏\n📌 请接：{currCy}{await MinusCreditAsync(ctx)}"
                        : "";
                }

                //是否正确
                if (await _chengyuService.PinYinFirstAsync(cmdPara) == await _chengyuService.PinYinLastAsync(currCy))
                {
                    if (await IsDupAsync(groupId, userId, cmdPara))
                        return "已有人接过此成语，请勿重复！";

                    creditInfo = await AddCreditAsync(ctx);
                    await AppendAsync(groupId, userId, name, cmdPara, 0);
                    currCy = cmdPara;
                    res = $"✅ 接龙『{cmdPara}』成功！{await GetGameCountStrAsync(groupId, userId)}";
                }
                else if (cmdPara == currCy)
                    return "被人抢先了，下次出手要快！";
                else
                    return $"接龙『{cmdPara}』不成功！\n📌 请接：{currCy}\n🔤 拼音：{pinyin}{await MinusCreditAsync(ctx)}";
            }

            currCy = await GetJielongAsync(groupId, userId, currCy);
            if (currCy != "")
            {
                await SetLastChengyuAsync(groupId, userId, isGroup, currCy);
                if (isGroup)
                    await AppendAsync(groupId, long.Parse(ctx.BotId), "", currCy, 0);
                else
                    await AppendAsync(groupId, userId, name, currCy, 0);
                res = $"{res}\n📌 请接：{currCy}\n🔤 拼音：{await _chengyuService.PinYinAsync(currCy)}{creditInfo}";
            }
            else
            {
                await GameOverAsync(groupId, userId, isGroup);
                await SetLastChengyuAsync(groupId, userId, isGroup, "");
                res = $"✅ {res}\n📌 我不会接『{cmdPara}』，你赢了{creditInfo}";
            }
            return res;
        }

        public async Task<int> SetLastChengyuAsync(long groupId, long userId, bool isGroup, string currCy)
        {
            return isGroup
                ? await _groupRepo.StartCyGameAsync(1, currCy, groupId)
                : await _userRepo.SetValueAsync("LastChengyu", currCy, userId);
        }

        public async Task<int> StartAsync(long groupId, long userId, bool isGroup, string cmdPara)
        {
            return isGroup
                ? await _groupRepo.StartCyGameAsync(1, cmdPara, groupId)
                : await _userRepo.SetStateAsync((int)UserStates.GameCy, userId);
        }

        public async Task<int> GameOverAsync(long groupId, long userId, bool isGroup)
        {
            return isGroup
                ? await _groupRepo.SetInGameAsync(0, groupId)
                : await _userRepo.SetStateAsync((int)UserStates.Chat, userId);
        }

        public async Task<string> CurrCyAsync(long groupId, long userId, bool isGroup)
        {
            if (!isGroup)
            {
                var user = await _userRepo.GetAsync(userId);
                return user?.LastChengyu ?? "";
            }
            else
            {
                return (await _groupRepo.GetAsync(groupId))?.LastChengyu ?? "";
            }
        }

        public async Task<bool> UserInGameAsync(long groupId, long userId, bool isGroup)
        {
            var user = await _userRepo.GetAsync(userId);
            if (user == null) return false;
            int state = user.State;
            return !isGroup ? state == (int)UserStates.GameCy : (state == (int)UserStates.Chat || state == (int)UserStates.GameCy);
        }

        public async Task<bool> InGameAsync(long groupId, long userId)
        {
            var user = await _userRepo.GetAsync(userId);
            if (user == null) return false;
            int state = user.State;
            
            var group = await _groupRepo.GetAsync(groupId);
            bool isGroup = group != null;

            if (!isGroup)            
                return state == (int)UserStates.GameCy;            
            else
            {
                var isInGame = group != null && group.IsInGame > 0;
                return isInGame && (state == (int)UserStates.Chat || state == (int)UserStates.GameCy);
            }
        }

        public async Task<int> AppendAsync(long groupId, long qq, string name, string chengYu, int gameNo)
        {
            return await _repository.AppendAsync(groupId, qq, name, chengYu, gameNo);
        }

        public async Task<bool> IsDupAsync(long groupId, long qq, string chengYu)
        {
            return await _repository.IsDupAsync(groupId, qq, chengYu);
        }

        public async Task<string> GetJielongAsync(long groupId, long UserId, string currCy)
        {
            string pinyin = await _chengyuService.PinYinLastAsync(currCy);
            return await _repository.GetChengYuByPinyinAsync(pinyin, groupId) ?? "";
        }

        public async Task<int> GetMaxIdAsync(long groupId)
        {
            return await _repository.GetMaxIdAsync(groupId);
        }

        public async Task<string> GetGameCountStrAsync(long groupId, long userId)
        {
            int count = await GetCountAsync(groupId, userId);
            return count > 0 ? $"(第{count}个)" : "";
        }

        public async Task<int> GetCountAsync(long groupId, long userId)
        {
            return await _repository.GetCountAsync(groupId, userId);
        }

        public async Task<long> GetCreditAddAsync(long userId)
        {
            return await _repository.GetCreditAddAsync(userId);
        }

        public async Task<string> AddCreditAsync(IPluginContext ctx)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var isGroup = ctx.GroupId != null;

            var creditAdd = 10;
            string res = "";
            
            var group = await _groupRepo.GetAsync(groupId);
            if ((!isGroup || await GetCreditAddAsync(userId) < 2000) && group?.IsCreditSystem == true)
            {
                var addRes = await _userRepo.AddCreditAsync(long.Parse(ctx.BotId), groupId, group.GroupName, userId, ctx.UserName, creditAdd, "成语接龙");
                if (addRes.Success)
                    res = $"\n💎 积分：+{creditAdd}，累计：{addRes.CreditValue:N0}";
            }
            return res;
        }

        public async Task<string> MinusCreditAsync(IPluginContext ctx)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            
            var creditMinus = 10;
            string res = "";
            
            var group = await _groupRepo.GetAsync(groupId);
            int c_chengyu = await GetCountAsync(groupId, userId);
            if (c_chengyu > 0 && group?.IsCreditSystem == true)
            {
                var addRes = await _userRepo.AddCreditAsync(long.Parse(ctx.BotId), groupId, group.GroupName, userId, ctx.UserName, -creditMinus, "成语接龙扣分");
                if (addRes.Success)
                    res = $"\n💎 积分：-{creditMinus} 累计：{addRes.CreditValue:N0}";
            }
            return res;
        }
    }
}
