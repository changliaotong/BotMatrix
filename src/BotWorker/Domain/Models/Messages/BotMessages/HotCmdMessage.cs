using System.Text.RegularExpressions;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Modules.Games;
using BotWorker.Domain.Constants;
using BotWorker.Infrastructure.Communication.Platforms.BotPublic;
using BotWorker.Common;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Tools;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {

        // 是否快捷命令
        public bool IsHot()
        {
            List<string> ListRegex =
            [
                Regexs.Study,
                Regexs.CreditUserId,
                Regexs.CreditUserId2,
                Regexs.SaveCredit,
                Regexs.CoinsUserId,
                Regexs.BlockCmd,
                Regexs.BlockCmdMult,
                Regexs.BuyCredit,
                Regexs.BlockHash16,
                Regexs.BlockHash,
                Regexs.BuyBet,
                Regexs.PetPrice,
                Regexs.Formula,
                Regexs.Caiquan,
                Regexs.WarnCmd,
                Regexs.Dati,
                Regexs.Fishing,
                Regexs.FishingBuy,
                Regexs.ExchangeCoins,
                Regexs.AddMinus,
                Regexs.BindToken,
                Regexs.LeaveGroup,
                Regexs.Todo,
            ];
            foreach (var item in ListRegex)
            {
                if (CurrentMessage.IsMatch(item))
                {
                    if (item == Regexs.BlockCmdMult && CurrentMessage.RegexReplace(item, "").Trim() != "")
                        continue;
                    return true;
                }
            }
            return false;
        }

        // 快捷指令结果
        public async Task GetHotCmdAsync()
        {
            string regexCmd = Regexs.Study;
            int c_da = CurrentMessage.Split('答').Length - 1;
            if (c_da > 1)
                regexCmd = Regexs.Study2;

            foreach (Match match in CurrentMessage.Matches(regexCmd))
                Answer = await AppendAnswerAsync(match.Groups["question"].Value, match.Groups["answer"].Value);

            if (Answer != "")
                return;

            CurrentMessage = CurrentMessage.RemoveQqFace();
            //学习功能不能去掉表情，这个位置不能换； ================================================================================================================

            //微信公众号绑定QQ
            regexCmd = Regexs.BindToken;
            if (Regex.IsMatch(CurrentMessage, regexCmd))
                Answer += ClientPublic.GetBindToken(this, Message.RegexGetValue(regexCmd, "token_type"), CurrentMessage.RegexGetValue(regexCmd, "bind_token"));

            //宠物
            regexCmd = Regexs.BuyBet;
            if (Regex.IsMatch(CurrentMessage, regexCmd))
                Answer += await PetOld.GetBuyPetAsync(SelfId, GroupId, GroupId, GroupName, UserId, Name, CurrentMessage.RegexGetValue(regexCmd, "UserId"));

            //需要@机器人参数的放在前面，不需要的放后面 ===============================================================================================================
            CurrentMessage = CurrentMessage.RemoveUserId(SelfId);
            //需要@机器人参数的放在前面，不需要的放后面 ===============================================================================================================

            // 存取积分
            if (CurrentMessage.IsMatch(Regexs.SaveCredit))
            {
                CmdName = CurrentMessage.RegexGetValue(Regexs.SaveCredit, "CmdName");
                CmdPara = CurrentMessage.RegexGetValue(Regexs.SaveCredit, "credit_value");
                Answer = await GetSaveCreditResAsync();
            }

            // 数学表达式
            else if (CurrentMessage.IsMatch(Regexs.Formula))
                Answer = Calc.GetJsRes(CurrentMessage.RegexGetValue(Regexs.Formula, "formula"));
            else if (CurrentMessage.IsMatch(Regexs.WarnCmd))//开启关闭 内置敏感词            
                GetWarnSetup(Regexs.WarnCmd);

            regexCmd = Regexs.Todo;
            if (CurrentMessage.IsMatch(regexCmd))
            {
                CmdName = CurrentMessage.RegexGetValue(regexCmd, "cmd_oper");
                CmdPara = CurrentMessage.RegexGetValue(regexCmd, "cmdPara");
                Answer = Todo.GetTodoRes(GroupId, GroupName, UserId, Name, CmdName, CmdPara);
            }
            else if (CurrentMessage.IsMatch(Regexs.LeaveGroup) && IsSuperAdmin) //退群
            {
                TargetUin = CurrentMessage.RegexGetValue(Regexs.LeaveGroup, "GroupId").AsLong();                
                await LeaveAsync(SelfId, TargetUin);
                Answer = $"收到，马上退";
            }
            else if (CurrentMessage.IsMatch(Regexs.Dati)) //答题
            {
                CurrentMessage = CurrentMessage.RegexGetValue(Regexs.Dati, "CmdName");
                Answer = (await GetDatiAsync(this)).Answer;
            }

            else if (CurrentMessage.IsMatch(Regexs.Fishing))//钓鱼
                Answer = Fishing.GetFishing(GroupId, GroupName, UserId, Name, CurrentMessage.RegexGetValue(Regexs.Fishing, "CmdName"), "");

            else if (CurrentMessage.IsMatch(Regexs.FishingBuy)) //买渔具
                Answer = Fishing.GetBuyTools(SelfId, GroupId, GroupName, UserId, Name,
                    CurrentMessage.RegexGetValue(Regexs.FishingBuy, "CmdName"),
                    CurrentMessage.RegexGetValue(Regexs.FishingBuy, "cmdPara"),
                    CurrentMessage.RegexGetValue(Regexs.FishingBuy, "cmdPara2"));

            else if (Message.IsMatch(Regexs.AddMinus)) //充值 扣除 积分、金币/紫币/游戏币等
                Answer = await GroupMember.AddCoinsResAsync(SelfId, GroupId, GroupName, UserId, Name,
                    CurrentMessage.RegexGetValue(Regexs.AddMinus, "CmdName"),
                    CurrentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara"),
                    CurrentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara2"),
                    CurrentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara3"));

            else if (CurrentMessage.IsMatch(Regexs.ExchangeCoins)) //兑换/购买 金币/紫币/游戏币等
                Answer = await ExchangeCoinsAsync(CurrentMessage.RegexGetValue(Regexs.ExchangeCoins, "cmdPara"), CurrentMessage.RegexGetValue(Regexs.ExchangeCoins, "cmdPara2"));
            
            else if (CurrentMessage.IsMatch(Regexs.Caiquan))//猜拳
            {
                string cmdName = Block.GetCmd(CurrentMessage.RegexGetValue(Regexs.Caiquan, "CmdName"), UserId);
                string cmdPara = CurrentMessage.RegexGetValue(Regexs.Caiquan, "cmdPara");
                CurrentMessage = $"{cmdName} {cmdPara}";
                await GetCmdResAsync();
            }
            else if (CurrentMessage.IsMatch(regexCmd))//身价
            {
                long friend_qq = CurrentMessage.RegexGetValue(Regexs.PetPrice, "UserId").AsLong();
                Answer = $"[@:{friend_qq}]的身价：{PetOld.GetSellPrice(GroupId, friend_qq)}";
                if (BlackList.IsSystemBlack(friend_qq))
                    Answer += "\n{BlackListMsg}";
            }
            else if (CurrentMessage.IsMatch(Regexs.CreditUserId) || CurrentMessage.IsMatch(Regexs.CreditUserId2))//积分
            {
                regexCmd = CurrentMessage.IsMatch(Regexs.CreditUserId) ? Regexs.CreditUserId : Regexs.CreditUserId2;
                long credit_qq = CurrentMessage.RegexGetValue(regexCmd, "UserId").AsLong();
                Answer = $"[@:{credit_qq}]的{UserInfo.GetCreditType(GroupId, credit_qq)}：{UserInfo.GetCredit(GroupId, credit_qq)}";
                if (BlackList.IsSystemBlack(credit_qq))
                    Answer += "\n{BlackListMsg}";
            }
            else if (CurrentMessage.IsMatch(regexCmd)) //金币
            {
                regexCmd = Regexs.CoinsUserId;
                long coins_qq = CurrentMessage.RegexGetValue(regexCmd, "UserId").AsLong();
                Answer = $"[@:{coins_qq}]的金币：{GroupMember.GetCoins((int)CoinsLog.CoinsType.goldCoins, GroupId, UserId):#0.00}";
                if (BlackList.IsSystemBlack(coins_qq))
                    Answer += "\n{BlackListMsg}";
            }

            // 查询block_info16
            else if (CurrentMessage.IsMatch(Regexs.BlockHash16))
                Answer = Block.GetBlockInfo16(CurrentMessage.RegexGetValue(Regexs.BlockHash16, "block_hash"));

            // 查询block_info
            else if (CurrentMessage.IsMatch(Regexs.BlockHash))
                Answer = Block.GetBlockInfo16(CurrentMessage.RegexGetValue(Regexs.BlockHash, "block_hash"));

            // 猜大小
            else if (CurrentMessage.IsMatch(Regexs.BlockCmd))
            {
                CmdName = $"{Block.GetCmd(CurrentMessage.RegexGetValue(Regexs.BlockCmd, "CmdName"), UserId)}";
                CmdPara = $"{CurrentMessage.RegexGetValue(Regexs.BlockCmd, "cmdPara")}";
                Answer = await GetBlockResAsync();
            }

            else if (CurrentMessage.IsMatch(Regexs.BlockCmdMult))
            {
                if (CurrentMessage.RegexReplace(Regexs.BlockCmdMult, "") == "")
                    Answer = await GetMultAsync();
            }

            //购买积分
            else if (CurrentMessage.IsMatch(Regexs.BuyCredit))
            {
                if (GroupId != 0 && !IsPublic)
                {
                    Answer = "请私聊使用此功能";
                    return;
                }
                Answer = await UserInfo.GetBuyCreditAsync(SelfId, GroupId, GroupName, UserId, Name,
                    CurrentMessage.RegexGetValue(Regexs.BuyCredit, "buy_qq").AsLong(),
                    CurrentMessage.RegexGetValue(Regexs.BuyCredit, "income_money").AsDecimal(),
                    CurrentMessage.RegexGetValue(Regexs.BuyCredit, "pay_method"));
            }

            //官方机器人审核通过的功能放前面，未上线的功能放后面===============================================================================================================
            if (Platform == Platforms.QQGuild)
                return;
            //官方机器人审核通过的功能放前面，未上线的功能放后面===============================================================================================================

            if (CurrentMessage.IsMatch(Regexs.Cid))//身份证号
            {
                CurrentMessage = CurrentMessage.RegexGetValue(Regexs.Cid, "cid");
                Answer = CID.GetCidRes(this);
            }

            return;
        }
    }
}
