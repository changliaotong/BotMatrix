using sz84.Bots.Entries;
using sz84.Bots.Games;
using sz84.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage :MetaData<BotMessage>
    {
        public const string ErrorFormat = "命令格式：开盲盒 + 数字1-6\n例如：\n开盲盒 3\nKMH 6";

        // 暗恋系统
        public async Task<string> GetSecretLove()
        {
            string strWhyLove = "\n为什么暗恋那么好？因为暗恋从来不会失恋。\n你一笑我高兴很多天，你一句话我记得好多年。";

            long countLove = SecretLove.GetCountLove(UserId);
            long countLoveme = SecretLove.GetCountLoveMe(UserId);

            if (!CmdPara.IsMatchQQ())
                return "📌 游戏格式：暗恋 + QQ 例如：\n暗恋 {客服QQ}";

            long loveQQ = CmdPara.AsLong();
            if (loveQQ == UserId)
                return "暗恋自己？简称自恋！";

            if (BotInfo.IsRobot(loveQQ))
                return "不要疯狂的迷恋我，我只是个传说！";

            if (SecretLove.Exists(UserId, loveQQ))
                return "这个已经暗恋过了，换一个？";

            if (SecretLove.Append(SelfId, UserId, loveQQ, RealGroupId) == -1)
                return RetryMsg;

            countLove++;

            if (SecretLove.IsLoveEachother(UserId, loveQQ))
            {
                Answer = $"✅ 恭喜你：你暗恋的对象[@:{CmdPara}]刚好也暗恋你，你们可以正大光明地恋爱了！";
                await SendMessageAsync();
            }
            else
                Answer = "✅ 登记成功！若TA也暗恋了你，会通知你们";

            Answer += $"\n你已暗恋{countLove}人，有{countLoveme}人暗恋你。\n{SecretLove.GetLoveStatus()}{strWhyLove}";   
            return Answer;
        }

        // 猜拳
        public string GetCaiquan()
        {
            if (!Group.IsCreditSystem) 
                return CreditSystemClosed;

            if (IsTooFast()) return RetryMsgTooFast;

            if (!CmdPara.IsNum() || CmdName == "猜拳")
                return "📌 游戏格式：\n石头 {最低积分}\n剪刀 {最低积分}\n布 {最低积分}";

            long blockCredit = CmdPara.AsLong();            
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < blockCredit)
                return $"您的积分{creditValue}不足{blockCredit}";

            int iRobot = Common.RandomInt(1, 3);
            long bonus = blockCredit;
            string strRobot = iRobot switch
            {
                1 => "剪刀",
                2 => "石头",
                3 => "布",
                _ => "剪刀"
            };
            if (strRobot == CmdName)
                return $"✅ 我出{strRobot}, 打平了！";

            //判输赢
            bool is_win = (CmdName == "石头" && strRobot == "剪刀")
                          || (CmdName == "剪刀" && strRobot == "布")
                          || (CmdName == "布" && strRobot == "石头");

            string strWin = "赢";
            if (is_win)
                bonus += (bonus * 98) / 100;
            else
            {
                bonus = 0;
                strWin = "输";
            }
            (int i, creditValue) = AddCredit(bonus - blockCredit, "猜拳得分");
            return i == -1
                ? RetryMsg
                : $"✅ 我出{strRobot}，你{strWin}了！ \n得分：{bonus}，累计：{creditValue}";
        }   


        public string GetGuessNum()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            string res = "";
            int cszTimes = UserInfo.GetInt("csz_times", UserId);
            int resCsz = UserInfo.GetInt("csz_res", UserId);
            long cszCredit = UserInfo.GetLong("csz_credit", UserId);
            long creditValue;
            if (CmdName == "猜数字")
            {
                //判断上局游戏是否结束
                if (resCsz != -1) return "上局游戏未结束，继续请发 我猜 + 数字";

                creditValue = UserInfo.GetCredit(GroupId, UserId);
                if (!CmdPara.IsNum())
                {
                    if (CmdPara == "梭哈")
                        CmdPara = creditValue.ToString();
                    else
                        return $"请押积分！您的积分{creditValue}";
                }

                long blockCredit = CmdPara.AsLong(); 
                if (blockCredit < Group.BlockMin)
                    return $"至少押{Group.BlockMin}分";

                if (creditValue < blockCredit)
                    return $"您的积分{creditValue}不足{blockCredit}";

                //生成随机数，保存积分以及猜测次数
                resCsz = Common.RandomInt(1, 13);
                cszCredit = blockCredit;

                if (UserInfo.NewGuessNumGame(resCsz, cszCredit, UserId) != -1)
                {
                    //扣分
                    MinusCredit(cszCredit, "猜数字扣分");
                    return $"您有3次机会，请发送：\n" +
                           $"我猜 + 数字\n-{cszCredit}分，累计：{creditValue}";
                }
                else
                    return "系统出错，请稍后重试";
            }
            else if (CmdName == "我猜")
            {
                if (resCsz == -1) return "开始游戏请先发 猜数字 + 积分 ";

                if (!CmdPara.IsNum())
                    return "请猜数字";

                int resGuess = int.Parse(CmdPara);
                if (resGuess < 0 || resGuess > 13)
                    return "请猜 0-13 中的一个数字";

                if (resCsz == resGuess)
                {
                    //猜对了结束游戏 加分
                    UserInfo.UpdateCszGame(-1, 0, 0, UserId);
                    long creditWin = (cszCredit * 19) / 10;
                    (int i, creditValue) = AddCredit(creditWin, "猜数字赢");
                    return i == -1 ? RetryMsg : $"✅ 恭喜：{cszTimes + 1}次猜对！\n得分：{creditWin}，累计：{creditValue}";
                }
                else
                {
                    //没猜对
                    if (cszTimes == 2)
                    {
                        //结束游戏
                        UserInfo.UpdateCszGame(-1, 0, 0, UserId);
                        return $"您猜错了，正确答案是：{resCsz}";
                    }
                    else
                    {
                        //继续猜
                        UserInfo.UpdateCszGame(resCsz, cszCredit, cszTimes + 1, UserId);
                        if (resCsz > resGuess)
                            return $"✅ 比{resGuess}大，还有{2 - cszTimes}次机会";
                        else
                            return $"✅ 比{resGuess}小，还有{2 - cszTimes}次机会";
                    }
                }
            }

            return res;
        }

        public string GetLuckyDraw()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (IsTooFast()) return RetryMsgTooFast;

            long creditValue = GetCredit();
            if (!CmdPara.IsNum())
            {
                if (CmdPara == "梭哈")
                    CmdPara = $"{creditValue}";
                else
                    return "🎁 格式：抽奖 + 数值\n📌 例如：抽奖 {最低积分}";
            }

            long credit = CmdPara.AsLong();
            if (credit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            if (creditValue < credit)
                return $"您只有{creditValue}分";

            long bonus = RandomInt64(credit * 2);
            long creditGet = bonus - credit;
            (int i, creditValue) = AddCredit(creditGet, $"抽奖 押{credit}中{bonus}得{creditGet}");
            return i == -1
                ? RetryMsg
                : $"✅ 得分：{bonus}，累计：{creditValue}";
        }

        public string GetSanggongRes()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (!CmdPara.IsNum())
            {
                if (CmdPara == "梭哈")
                {
                    CmdPara = UserInfo.GetCredit(GroupId, UserId).ToString();
                }
                else
                    return "🎁 格式：SG + 数值\n" +
                           "📌 例如：SG {最低积分}";
            }
            CmdName = "蓝";

            return GetRedBlueRes(false);
        }

        public bool IsTooFast()
        {
            //频率限制1分钟不能超过6次
            return CreditLog.CreditCount(UserId, "得分") > 20;
        }

        public string GetSanggongRes2()
        {
            if (IsTooFast()) return RetryMsgTooFast;

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (!CmdPara.IsNum())
            {
                if (CmdPara == "梭哈")
                    CmdPara = creditValue.ToString();
                else
                    return "格式：SG + 积分数\n例如：SG {最低积分}";
            }

            long blockCredit = CmdPara.AsLong();
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";
            if (creditValue < blockCredit)
                return $"您只有{creditValue}分";

            string typeName = $"押大";
            int typeId = BlockType.GetTypeId(typeName);
            int blockNum = BlockRandom.RandomNum();
            bool isWin = Block.IsWin(typeId, typeName, blockNum);
            long creditGet = 0;
            long creditAdd;
            if (isWin)
            {
                int blockOdds = Block.GetOdds(typeId, typeName, blockNum);
                creditAdd = blockCredit * blockOdds;
                creditGet = blockCredit * (blockOdds + 1);
            }
            else
                creditAdd = -blockCredit;

            (int i, creditValue) = AddCredit(creditAdd, "三公得分");
            return i == -1
                ? RetryMsg
                : $"✅ 得分：{creditGet}，累计：{creditValue}";
        }

        public async Task<string> GetMuteMeAsync()
        {
            if (IsNewAnswer)
                return "";

            if (!IsGroup)
                return "你让我禁言我就禁言？那样我岂不是很没面子";

            await MuteAsync(SelfId, RealGroupId, UserId, 10 * 60);

            return "";           
        }

        public async Task<string> GetKickmeAsync()
        {
            if (IsNewAnswer) return "";

            if (!IsGroup)
                return "你让我踢我就踢？那样我岂不是很没面子！";

            await KickOutAsync(SelfId, RealGroupId, UserId);

            return "";
        }

        public string GetDouniwan()
        {
            string res = SetupPrivate(false);
            if (res != "")
                return res;

            if (IsGroup)
                return "请私聊使用此功能";

            if (CmdPara.Trim() == "结束")
                return UserInfo.SetState(UserInfo.States.Chat, UserId) == -1
                    ? RetryMsg
                    : "✅ 逗你玩结束";

            //切换到逗你玩状态
            if (CmdPara == "")
            {
                UserInfo.SetState(UserInfo.States.Douniwan, UserId);
                res = "发消息逗群【{默认群}】的人玩吧～\n每条-10分，脏话或广告-50分或-100分";
            }
            else
            {
                //扣分
                long credit_minus = 10;
                if (CmdPara.IsMatch(Regexs.AdWords))
                    credit_minus = 50;
                if (CmdPara.IsMatch(Regexs.DirtyWords))
                    credit_minus = 100;
                MinusCreditRes(credit_minus, "逗你玩扣分");

                if ((credit_minus == 10) || IsSuperAdmin)
                {
                    //todo 转发消息到群
                    //this.AddGroupMessage(CurrentGroupId, UserId, CmdPara);
                    res = $"✅ 发送成功\n -{credit_minus}分，累计：{{积分}}";
                }
                else
                    res = $"禁止发脏话或广告\n -{credit_minus}分，累计：{{积分}}";
            }

            return res + GetHintInfo();
        }
    }
}
