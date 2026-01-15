using System.Reflection;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.jielong",
        Name = "成语接龙",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "趣味成语接龙游戏，答对奖励积分，答错扣除积分",
        Category = "Games"
    )]
    public class JielongPlugin : IPlugin
    {
        public BotPluginAttribute Metadata => GetType().GetCustomAttribute<BotPluginAttribute>()!;

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(new SkillCapability("成语接龙", ["接龙", "jl"]), HandleJielongAsync);
            // 注册消息处理事件，用于在游戏中直接接龙
            await robot.RegisterEventAsync("message", HandleUserMessageAsync);
        }

        private async Task EnsureTablesCreatedAsync()
        {
            // await Jielong.EnsureTableCreatedAsync();
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleJielongAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var cmdPara = args.Length > 0 ? string.Join(" ", args) : "";
            
            return await Jielong.GetJielongResAsync(ctx, cmdPara);
        }

        private async Task HandleUserMessageAsync(IPluginContext ctx)
        {
            if (ctx.GroupId == null) return;
            
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId);

            // 如果在游戏中，且消息看起来像成语（4个字）或者是强制接龙指令
            if (await Jielong.InGameAsync(groupId, userId))
            {
                var msg = ctx.RawMessage.Trim();
                if (msg.Length == 4 || msg.StartsWith("接龙") || msg.StartsWith("jl"))
                {
                    var cmdPara = msg.StartsWith("接龙") ? msg[2..].Trim() : (msg.StartsWith("jl") ? msg[2..].Trim() : msg);
                    var res = await Jielong.GetJielongResAsync(ctx, cmdPara);
                    if (!string.IsNullOrEmpty(res))
                    {
                        await ctx.ReplyAsync(res);
                    }
                }
            }
        }
    }

    public class Jielong
    {
        private static IJielongRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IJielongRepository>() 
            ?? throw new InvalidOperationException("IJielongRepository not registered");

        private static IUserRepository UserRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IUserRepository>() 
            ?? throw new InvalidOperationException("IUserRepository not registered");

        private static IGroupRepository GroupRepository => 
            BotMessage.ServiceProvider?.GetRequiredService<IGroupRepository>() 
            ?? throw new InvalidOperationException("IGroupRepository not registered");

        public long Id { get; set; }
        public long GroupId { get; set; }
        public long UserId { get; set; }
        public string UserName { get; set; } = "";
        public string chengyu { get; set; } = "";
        public int GameNo { get; set; }
        public int Credit { get; set; }
        public DateTime InsertDate { get; set; } = DateTime.Now;

        public static async Task<string> GetJielongResAsync(IPluginContext ctx, string cmdPara)
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

                if (cmdPara.IsNull())
                    cmdPara = (await Chengyu.GetRandomAsync("chengyu")).RemoveBiaodian();
                else if (!await Chengyu.ExistsAsync(cmdPara))
                {
                    var user = await UserInfo.GetSingleAsync(userId);
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
                string pinyin = await Chengyu.PinYinAsync(currCy);
                cmdPara = cmdPara.RemoveQqAds();
                if (cmdPara == "")
                    return ctx.RawMessage.Contains("接龙") || ctx.RawMessage == ""
                        ? $"发【结束】退出游戏\n📌 请接：{currCy}\n🔤 拼音：{pinyin}"
                        : "";

                if (cmdPara == "提示")
                    return (await GetJielongAsync(groupId, userId, currCy)).MaskIdiom();

                if (!await Chengyu.ExistsAsync(cmdPara))
                {
                    if (isGroup && await GroupInfo.GetChengyuIdleMinutesAsync(groupId) > 10)
                    {
                        await GroupInfo.SetInGameAsync(0, groupId);
                        return "✅ 成语接龙超时自动结束";
                    }
                    return cmdPara.Length == 4 || ctx.RawMessage.StartsWith("接龙") || ctx.RawMessage.StartsWith("jl")
                        ? $"【{cmdPara}】不是成语\n💡 发【结束】退出游戏\n📌 请接：{currCy}{await MinusCreditAsync(ctx)}"
                        : "";
                }

                //是否正确
                if (await Chengyu.PinYinFirstAsync(cmdPara) == await Chengyu.PinYinLastAsync(currCy))
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
                res = $"{res}\n📌 请接：{currCy}\n🔤 拼音：{await Chengyu.PinYinAsync(currCy)}{creditInfo}";
            }
            else
            {
                await GameOverAsync(groupId, userId, isGroup);
                await SetLastChengyuAsync(groupId, userId, isGroup, "");
                res = $"✅ {res}\n📌 我不会接『{cmdPara}』，你赢了{creditInfo}";
            }
            return res;
        }

        public static async Task<int> SetLastChengyuAsync(long groupId, long userId, bool isGroup, string currCy)
        {
            return isGroup
                ? await GroupInfo.StartCyGameAsync(1, currCy, groupId)
                : await UserInfo.SetValueAsync("LastChengyu", currCy, userId);
        }

        public static async Task<int> StartAsync(long groupId, long userId, bool isGroup, string cmdPara)
        {
            return isGroup
                ? await GroupInfo.StartCyGameAsync(1, cmdPara, groupId)
                : await UserInfo.SetStateAsync(UserInfo.States.GameCy, userId);
        }

        public static async Task<int> GameOverAsync(long groupId, long userId, bool isGroup)
        {
            return isGroup
                ? await GroupInfo.SetInGameAsync(0, groupId)
                : await UserInfo.SetStateAsync(UserInfo.States.Chat, userId);
        }

        public static async Task<string> CurrCyAsync(long groupId, long userId, bool isGroup)
        {
            if (!isGroup)
            {
                var user = await UserInfo.GetSingleAsync(userId);
                return user?.LastChengyu ?? "";
            }
            else
            {
                return (await GroupInfo.GetSingleAsync(groupId))?.LastChengyu ?? "";
            }
        }

        public static async Task<bool> UserInGameAsync(long groupId, long userId, bool isGroup)
        {
            var user = await UserInfo.GetSingleAsync(userId);
            if (user == null) return false;
            int state = user.State;
            return !isGroup ? state == (int)UserInfo.States.GameCy : state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
        }

        public static async Task<bool> InGameAsync(long groupId, long userId)
        {
            var user = await UserInfo.GetSingleAsync(userId);
            if (user == null) return false;
            int state = user.State;
            
            var group = await GroupInfo.GetSingleAsync(groupId);
            bool isGroup = group != null;

            if (!isGroup)            
                return state == (int)UserInfo.States.GameCy;            
            else
            {
                var isInGame = group != null && group.IsInGame > 0;
                return isInGame && state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
            }
        }

        // 添加接龙成功的数据到数据库
        public static async Task<int> AppendAsync(long groupId, long qq, string name, string chengYu, int gameNo)
        {
            return await Repository.AppendAsync(groupId, qq, name, chengYu, gameNo);
        }

        // 是否重复成语
        public static async Task<bool> IsDupAsync(long groupId, long qq, string chengYu)
        {
            return await Repository.IsDupAsync(groupId, qq, chengYu);
        }

        // 为机器人获取一个随机接龙结果
        public static async Task<string> GetJielongAsync(long groupId, long UserId, string currCy)
        {
            string pinyin = await Chengyu.PinYinLastAsync(currCy);
            return await Repository.GetChengYuByPinyinAsync(pinyin, groupId) ?? "";
        }

        // 接龙游戏最大ID
        public static async Task<int> GetMaxIdAsync(long groupId)
        {
            return await Repository.GetMaxIdAsync(groupId);
        }

        // 接龙成功数量
        public static async Task<string> GetGameCountStrAsync(long groupId, long userId)
        {
            int count = await GetCountAsync(groupId, userId);
            return count > 0 ? $"(第{count}个)" : "";
        }

        // 接龙成功数量
        public static async Task<int> GetCountAsync(long groupId, long userId)
        {
            return await Repository.GetCountAsync(groupId, userId);
        }

        // 接龙加分总数
        public static async Task<long> GetCreditAddAsync(long userId)
        {
            return await Repository.GetCreditAddAsync(userId);
        }

        // 成语接龙加分
        public static async Task<string> AddCreditAsync(IPluginContext ctx)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var isGroup = ctx.GroupId != null;

            var creditAdd = 10;
            string res = "";
            
            var group = await GroupInfo.GetSingleAsync(groupId);
            if ((!isGroup || await GetCreditAddAsync(userId) < 2000) && group?.IsCreditSystem == true)
            {
                var addRes = await UserInfo.AddCreditAsync(long.Parse(ctx.BotId), groupId, group.GroupName, userId, ctx.UserName, creditAdd, "成语接龙");
                if (addRes.Item1 != -1)
                    res = $"\n💎 积分：+{creditAdd}，累计：{addRes.Item2:N0}";
            }
            return res;
        }

        // 成语接龙扣分
        public static async Task<string> MinusCreditAsync(IPluginContext ctx)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            
            var creditMinus = 10;
            string res = "";
            
            var group = await GroupInfo.GetSingleAsync(groupId);
            int c_chengyu = await GetCountAsync(groupId, userId);
            if (c_chengyu > 0 && group?.IsCreditSystem == true)
            {
                var addRes = await UserInfo.AddCreditAsync(long.Parse(ctx.BotId), groupId, group.GroupName, userId, ctx.UserName, -creditMinus, "成语接龙扣分");
                if (addRes.Item1 != -1)
                    res = $"\n💎 积分：-{creditMinus} 累计：{addRes.Item2:N0}";
            }
            return res;
        }
    }
}
