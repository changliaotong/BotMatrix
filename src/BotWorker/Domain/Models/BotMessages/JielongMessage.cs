namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        // 成语接龙
        public async Task<string> GetJielongRes()
        {
            CmdPara = CmdPara.RemoveBiaodian().Trim();
            if (CmdPara == "结束")
            {
                //只想结束翻译?
                if (await UserInGameAsync())
                {
                    return await GameOverAsync() == -1
                        ? RetryMsg
                        : $"✅ 成语接龙游戏结束{await Jielong.MinusCreditAsync(this)}";
                }
                return "";
            }

            bool inGame = await InGameAsync();
            string currCy;
            string res;
            string creditInfo;
            if (!inGame)
            {
                if (CmdPara == "")
                    CmdPara = await CurrCyAsync();

                if (CmdPara.IsNull())
                    CmdPara = (await Chengyu.GetRandomAsync("chengyu")).RemoveBiaodian();
                else if (!await Chengyu.ExistsAsync(CmdPara))
                    return User.IsSuper || User.CreditTotal > 10000 ? $"【{CmdPara}】不是成语" : $"您输入的不是成语";                

                await Jielong.AppendAsync(GroupId, UserId, Name, CmdPara, 1);
                await StartAsync();
                currCy = CmdPara;
                creditInfo = await Jielong.AddCreditAsync(this);
                res = $"✅ 成语接龙开始！";
            }
            else
            {
                currCy = await CurrCyAsync();
                string pinyin = await Chengyu.PinYinAsync(currCy);
                CmdPara = CmdPara.RemoveQqAds();
                if (CmdPara == "")
                    return Message.Contains("接龙") || Message == ""
                        ? $"发【结束】退出游戏\n📌 请接：{currCy}\n🔤 拼音：{pinyin}"
                        : "";

                if (CmdPara == "提示")
                    return (await Jielong.GetJielongAsync(GroupId, UserId, currCy)).MaskIdiom();

                if (!await Chengyu.ExistsAsync(CmdPara))
                {
                    if (IsGroup && await GroupInfo.GetChengyuIdleMinutesAsync(GroupId) > 10)
                    {
                        await GroupInfo.SetInGameAsync(0, GroupId);
                        Answer = "✅ 成语接龙超时自动结束";
                        await SendMessageAsync();
                        return "";
                    }
                    return CmdPara.Length == 4 || Message.StartsWith("接龙") || Message.StartsWith("jl")
                        ? $"【{CmdPara}】不是成语\n💡 发【结束】退出游戏\n📌 请接：{currCy}{await Jielong.MinusCreditAsync(this)}"
                        : "";
                }

                //是否正确
                if (await Chengyu.PinYinFirstAsync(CmdPara) == await Chengyu.PinYinLastAsync(currCy))
                {
                    if (await Jielong.IsDupAsync(GroupId, UserId, CmdPara))
                        return "已有人接过此成语，请勿重复！";

                    creditInfo = await Jielong.AddCreditAsync(this);
                    await Jielong.AppendAsync(GroupId, UserId, Name, CmdPara, 0);
                    currCy = CmdPara;
                    res = $"✅ 接龙『{CmdPara}』成功！{await Jielong.GetGameCountAsync(GroupId, UserId)}";
                }
                else if (CmdPara == currCy)
                    return "被人抢先了，下次出手要快！";
                else
                    return $"接龙『{CmdPara}』不成功！\n📌 请接：{currCy}\n🔤 拼音：{pinyin}{await Jielong.MinusCreditAsync(this)}";
            }

            currCy = await Jielong.GetJielongAsync(GroupId, UserId, currCy);
            if (currCy != "")
            {
                await SetLastChengyuAsync(currCy);
                if (IsGroup)
                    await Jielong.AppendAsync(GroupId, SelfId, "", currCy, 0);
                else
                    await Jielong.AppendAsync(GroupId, UserId, Name, currCy, 0);
                res = $"{res}\n📌 请接：{currCy}\n🔤 拼音：{await Chengyu.PinYinAsync(currCy)}{creditInfo}";
            }
            else
            {
                await GameOverAsync();
                await SetLastChengyuAsync("");
                res = $"✅ {res}\n📌 我不会接『{CmdPara}』，你赢了{creditInfo}";
            }
            return res;
        }        

        /// 更新游戏当前要接龙的成语到数据库
        public async Task<int> SetLastChengyuAsync(string currCy)
        {
            return IsGroup
                ? await GroupInfo.StartCyGameAsync(1, currCy, GroupId)
                : await UserInfo.SetValueAsync("LastChengyu", currCy, UserId);
        }

        // 开局游戏
        public async Task<int> StartAsync()
        {
            return IsGroup
                ? await GroupInfo.StartCyGameAsync(1, CmdPara, GroupId)
                : await UserInfo.SetStateAsync(UserInfo.States.GameCy, UserId);
        }

        // 结束游戏
        public async Task<int> GameOverAsync()
        {
            return IsGroup
                ? await GroupInfo.SetInGameAsync(0, GroupId)
                : await UserInfo.SetStateAsync(UserInfo.States.Chat, UserId);
        }

        // 当前成语
        public async Task<string> CurrCyAsync()
        {
            return !IsGroup
                ? User.LastChengyu
                : (await GroupInfo.GetSingleAsync(GroupId))?.LastChengyu ?? "";
        }

        // 用户是否游戏中
        public async Task<bool> UserInGameAsync()
        {
            int state = User.State;
            return !IsGroup ? state == (int)UserInfo.States.GameCy : state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
        }

        // 判断群或个人是否在游戏中
        public async Task<bool> InGameAsync()
        {
            int state = User.State;
            if (!IsGroup)            
                return state == (int)UserInfo.States.GameCy;            
            else
            {
                var group = await GroupInfo.GetSingleAsync(GroupId);
                var isInGame = group != null && group.IsInGame > 0;
                return isInGame && state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
            }
        }
    }
}
