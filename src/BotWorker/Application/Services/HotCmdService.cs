using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using System.Linq;
using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Entities;
using BotWorker.Infrastructure.Tools;
using BotWorker.Application.Messaging.Handlers;
using BotWorker.Infrastructure.Utils;
using BotWorker.Modules.Games;
using BotWorker.Modules.Office;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Communication.Platforms.BotPublic;
using BotWorker.Common;

namespace BotWorker.Application.Services
{
    public class HotCmdService : IHotCmdService
    {
        private readonly IPermissionService _permissionService;
        private readonly IBotApiService _botApiService;
        private readonly IUserService _userService;

        public HotCmdService(IPermissionService permissionService, IBotApiService botApiService, IUserService userService)
        {
            _permissionService = permissionService;
            _botApiService = botApiService;
            _userService = userService;
        }

        public bool IsHot(BotMessage botMsg)
        {
            List<string> listRegex = new List<string>
            {
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
            };

            foreach (var item in listRegex)
            {
                if (botMsg.CurrentMessage.IsMatch(item))
                {
                    if (item == Regexs.BlockCmdMult && botMsg.CurrentMessage.RegexReplace(item, "").Trim() != "")
                        continue;
                    return true;
                }
            }
            return false;
        }

        public async Task<CommandResult> HandleHotCmdAsync(BotMessage botMsg)
        {
            string answer = string.Empty;
            string currentMessage = botMsg.CurrentMessage;

            // 1. 学习功能 (不移除表情)
            string regexCmd = Regexs.Study;
            int c_da = currentMessage.Split('答').Length - 1;
            if (c_da > 1) regexCmd = Regexs.Study2;

            foreach (Match match in currentMessage.Matches(regexCmd))
            {
                answer = AppendAnswer(answer, match.Groups["question"].Value, match.Groups["answer"].Value);
            }

            if (!string.IsNullOrEmpty(answer))
                return CommandResult.Intercepted(answer);

            // 2. 预处理消息 (移除表情和机器人QQ号)
            currentMessage = currentMessage.RemoveQqFace();
            
            // 微信公众号绑定QQ
            regexCmd = Regexs.BindToken;
            if (Regex.IsMatch(currentMessage, regexCmd))
            {
                answer = ClientPublic.GetBindToken(botMsg, botMsg.Message.RegexGetValue(regexCmd, "token_type"), currentMessage.RegexGetValue(regexCmd, "bind_token"));
                return CommandResult.Intercepted(answer);
            }

            // 宠物
            regexCmd = Regexs.BuyBet;
            if (Regex.IsMatch(currentMessage, regexCmd))
            {
                answer = await PetOld.GetBuyPetAsync(botMsg.SelfId, botMsg.GroupId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, currentMessage.RegexGetValue(regexCmd, "UserId"));
                return CommandResult.Intercepted(answer);
            }

            // 移除机器人QQ号
            currentMessage = currentMessage.RemoveUserId(botMsg.SelfId);

            // 3. 核心快捷指令逻辑
            
            // 存取积分
            if (currentMessage.IsMatch(Regexs.SaveCredit))
            {
                answer = await _userService.HandleSaveCreditAsync(botMsg);
            }
            else if (currentMessage.IsMatch(Regexs.RewardCredit))
            {
                answer = await _userService.HandleRewardCreditAsync(botMsg);
            }
            else if (currentMessage.IsMatch(Regexs.CreditList))
            {
                answer = await _userService.GetCreditRankAsync(botMsg);
            }
            // 数学表达式
            else if (currentMessage.IsMatch(Regexs.Formula))
            {
                answer = Calc.GetJsRes(currentMessage.RegexGetValue(Regexs.Formula, "formula"));
            }
            // 敏感词设置
            else if (currentMessage.IsMatch(Regexs.WarnCmd))
            {
                botMsg.GetWarnSetup(Regexs.WarnCmd); // 暂时保留
                answer = botMsg.Answer;
            }
            // Todo 待办
            else if (currentMessage.IsMatch(Regexs.Todo))
            {
                string cmdName = currentMessage.RegexGetValue(Regexs.Todo, "cmd_oper");
                string cmdPara = currentMessage.RegexGetValue(Regexs.Todo, "cmdPara");
                answer = Todo.GetTodoRes(botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, cmdName, cmdPara);
            }
            // 退出(超级管理员权限)
            else if (currentMessage.IsMatch(Regexs.LeaveGroup) && botMsg.IsSuperAdmin)
            {
                long targetUin = currentMessage.RegexGetValue(Regexs.LeaveGroup, "GroupId").AsLong();
                await _botApiService.LeaveGroupAsync(botMsg.SelfId, targetUin);
                answer = "收到，马上退";
            }
            // 答题
            else if (currentMessage.IsMatch(Regexs.Dati))
            {
                botMsg.CurrentMessage = currentMessage.RegexGetValue(Regexs.Dati, "CmdName");
                var datiResult = await botMsg.GetDatiAsync(botMsg);
                answer = datiResult.Answer;
            }
            // 钓鱼
            else if (currentMessage.IsMatch(Regexs.Fishing))
            {
                answer = await Fishing.GetFishing(botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name, currentMessage.RegexGetValue(Regexs.Fishing, "CmdName"), "");
            }
            // 买渔具
            else if (currentMessage.IsMatch(Regexs.FishingBuy))
            {
                answer = Fishing.GetBuyTools(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name,
                    currentMessage.RegexGetValue(Regexs.FishingBuy, "CmdName"),
                    currentMessage.RegexGetValue(Regexs.FishingBuy, "cmdPara"),
                    currentMessage.RegexGetValue(Regexs.FishingBuy, "cmdPara2"));
            }
            // 积分金币充值扣除
            else if (botMsg.Message.IsMatch(Regexs.AddMinus))
            {
                answer = GroupMember.AddCoinsRes(botMsg.SelfId, botMsg.GroupId, botMsg.GroupName, botMsg.UserId, botMsg.Name,
                    currentMessage.RegexGetValue(Regexs.AddMinus, "CmdName"),
                    currentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara"),
                    currentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara2"),
                    currentMessage.RegexGetValue(Regexs.AddMinus, "cmdPara3"));
            }
            // 兑换金币
            else if (currentMessage.IsMatch(Regexs.ExchangeCoins))
            {
                answer = await _userService.ExchangeCoinsAsync(botMsg, currentMessage.RegexGetValue(Regexs.ExchangeCoins, "cmdPara"), currentMessage.RegexGetValue(Regexs.ExchangeCoins, "cmdPara2"));
            }
            // 猜拳
            else if (currentMessage.IsMatch(Regexs.Caiquan))
            {
                string cmdName = Block.GetCmd(currentMessage.RegexGetValue(Regexs.Caiquan, "CmdName"), botMsg.UserId);
                string cmdPara = currentMessage.RegexGetValue(Regexs.Caiquan, "cmdPara");
                botMsg.CurrentMessage = $"{cmdName} {cmdPara}";
                await botMsg.GetCmdResAsync();
                answer = botMsg.Answer;
            }
            // 身价
            else if (currentMessage.IsMatch(Regexs.PetPrice))
            {
                long friend_qq = currentMessage.RegexGetValue(Regexs.PetPrice, "UserId").AsLong();
                answer = $"[@:{friend_qq}]的身价：{PetOld.GetSellPrice(botMsg.GroupId, friend_qq)}";
                if (BlackList.IsSystemBlack(friend_qq)) answer += "\n{BlackListMsg}";
            }
            // 积分查询
            else if (currentMessage.IsMatch(Regexs.CreditUserId) || currentMessage.IsMatch(Regexs.CreditUserId2))
            {
                string matchRegex = currentMessage.IsMatch(Regexs.CreditUserId) ? Regexs.CreditUserId : Regexs.CreditUserId2;
                long credit_qq = currentMessage.RegexGetValue(matchRegex, "UserId").AsLong();
                answer = $"[@:{credit_qq}]的{{积分类型}}：{{积分}}";
                if (BlackList.IsSystemBlack(credit_qq)) answer += "\n{BlackListMsg}";
            }
            // 金币查询
            else if (currentMessage.IsMatch(Regexs.CoinsUserId))
            {
                long coins_qq = currentMessage.RegexGetValue(Regexs.CoinsUserId, "UserId").AsLong();
                answer = $"[@:{coins_qq}]的金币：{GroupMember.GetCoins((int)CoinsLog.CoinsType.goldCoins, botMsg.GroupId, botMsg.UserId):#0.00}";
                if (BlackList.IsSystemBlack(coins_qq)) answer += "\n{BlackListMsg}";
            }
            // 猜大小
            else if (currentMessage.IsMatch(Regexs.BlockCmd))
            {
                botMsg.CmdName = $"{Block.GetCmd(currentMessage.RegexGetValue(Regexs.BlockCmd, "CmdName"), botMsg.UserId)}";
                botMsg.CmdPara = $"{currentMessage.RegexGetValue(Regexs.BlockCmd, "cmdPara")}";
                answer = await botMsg.GetBlockResAsync();
            }
            // 身份证号
            else if (currentMessage.IsMatch(Regexs.Cid))
            {
                botMsg.CurrentMessage = currentMessage.RegexGetValue(Regexs.Cid, "cid");
                answer = CID.GetCidRes(botMsg);
            }

            if (!string.IsNullOrEmpty(answer))
                return CommandResult.Intercepted(answer);

            return CommandResult.Continue();
        }

        private string AppendAnswer(string currentAnswer, string question, string answer)
        {
            // 模拟 BotMessage 中的 AppendAnswer 逻辑
            string res = $"{question} {answer}";
            if (string.IsNullOrEmpty(currentAnswer)) return res;
            return currentAnswer + "\n" + res;
        }
    }
}
