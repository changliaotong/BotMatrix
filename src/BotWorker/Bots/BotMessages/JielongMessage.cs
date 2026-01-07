using sz84.Bots.Entries;
using sz84.Bots.Games;
using sz84.Bots.Users;
using BotWorker.Common.Exts;
using sz84.Core.Data;
using sz84.Core.MetaDatas;

namespace sz84.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        // 成语接龙
        public async Task<string> GetJielongRes()
        {
            CmdPara = CmdPara.RemoveBiaodian().Trim();
            if (CmdPara == "结束")
            {
                //只想结束翻译?
                if (UserInGame())
                {
                    return GameOver() == -1
                        ? RetryMsg
                        : $"✅ 成语接龙游戏结束{Jielong.MinusCredit(this)}";
                }
                return "";
            }

            bool inGame = InGame();
            string currCy;
            string res;
            string creditInfo;
            if (!inGame)
            {
                if (CmdPara == "")
                    CmdPara = CurrCy();

                if (CmdPara.IsNull())
                    CmdPara = Chengyu.GetRandom("chengyu").RemoveBiaodian();
                else if (!Chengyu.Exists(CmdPara))
                    return User.IsSuper || User.CreditTotal > 10000 ? $"【{CmdPara}】不是成语" : $"您输入的不是成语";                

                Jielong.Append(GroupId, UserId, Name, CmdPara, 1);
                Start();
                currCy = CmdPara;
                creditInfo = Jielong.AddCredit(this);
                res = $"✅ 成语接龙开始！";
            }
            else
            {
                currCy = CurrCy();
                string pinyin = Chengyu.PinYin(currCy);
                CmdPara = CmdPara.RemoveQqAds();
                if (CmdPara == "")
                    return Message.Contains("接龙") || Message == ""
                        ? $"发【结束】退出游戏\n📌 请接：{currCy}\n🔤 拼音：{pinyin}"
                        : "";

                if (CmdPara == "提示")
                    return Jielong.GetJielong(GroupId, UserId, currCy).MaskIdiom();

                if (!Chengyu.Exists(CmdPara))
                {
                    if (IsGroup && GroupInfo.GetInt("DATEDIFF(MINUTE, LastChengyuDate, GETDATE())", GroupId) > 10)
                    {
                        GroupInfo.SetInGame(0, GroupId);
                        Answer = "✅ 成语接龙超时自动结束";
                        await SendMessageAsync();
                        return "";
                    }
                    return CmdPara.Length == 4 || Message.StartsWith("接龙") || Message.StartsWith("jl")
                        ? $"【{CmdPara}】不是成语\n💡 发【结束】退出游戏\n📌 请接：{currCy}{Jielong.MinusCredit(this)}"
                        : "";
                }

                //是否正确
                if (Chengyu.PinYinFirst(CmdPara) == Chengyu.PinYinLast(currCy))
                {
                    if (Jielong.IsDup(GroupId, UserId, CmdPara))
                        return "已有人接过此成语，请勿重复！";

                    creditInfo = Jielong.AddCredit(this);
                    Jielong.Append(GroupId, UserId, Name, CmdPara, 0);
                    currCy = CmdPara;
                    res = $"✅ 接龙『{CmdPara}』成功！{Jielong.GetGameCount(GroupId, UserId)}";
                }
                else if (CmdPara == currCy)
                    return "被人抢先了，下次出手要快！";
                else
                    return $"接龙『{CmdPara}』不成功！\n📌 请接：{currCy}\n🔤 拼音：{pinyin}{Jielong.MinusCredit(this)}";
            }

            currCy = Jielong.GetJielong(GroupId, UserId, currCy);
            if (currCy != "")
            {
                SetLastChengyu(currCy);
                if (IsGroup)
                    Jielong.Append(GroupId, SelfId, "", currCy, 0);
                else
                    Jielong.Append(GroupId, UserId, Name, currCy, 0);
                res = $"{res}\n📌 请接：{currCy}\n🔤 拼音：{Chengyu.PinYin(currCy)}{creditInfo}";
            }
            else
            {
                GameOver();
                SetLastChengyu("");
                res = $"✅ {res}\n📌 我不会接『{CmdPara}』，你赢了{creditInfo}";
            }
            return res;
        }        

        /// 更新游戏当前要接龙的成语到数据库
        public int SetLastChengyu(string currCy)
        {
            return IsGroup
                ? GroupInfo.StartCyGame(1, currCy, GroupId)
                : UserInfo.SetValue("LastChengyu", currCy, UserId);
        }

        // 开局游戏
        public int Start()
        {
            return IsGroup
                ? GroupInfo.StartCyGame(1, CmdPara, GroupId)
                : UserInfo.SetState(UserInfo.States.GameCy, UserId);
        }

        // 结束游戏
        public int GameOver()
        {
            return IsGroup
                ? GroupInfo.SetInGame(0, GroupId)
                : UserInfo.SetState(UserInfo.States.Chat, UserId);
        }

        // 当前成语
        public string CurrCy()
        {
            return !IsGroup
                ? User.LastChengyu
                : Group.LastChengyu;
        }

        // 用户是否游戏中
        public  bool UserInGame()
        {
            int state = User.State;
            return !IsGroup ? state == (int)UserInfo.States.GameCy : state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
        }

        // 判断群或个人是否在游戏中
        public bool InGame()
        {
            int state = User.State;
            if (!IsGroup)            
                return state == (int)UserInfo.States.GameCy;            
            else
            {
                var isInGame = Group.IsInGame > 0;
                return isInGame && state.In((int)UserInfo.States.Chat, (int)UserInfo.States.GameCy);
            }
        }
    }
}
