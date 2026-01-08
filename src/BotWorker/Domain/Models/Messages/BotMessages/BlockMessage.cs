using System.Text.RegularExpressions;
using sz84.Bots.Entries;
using sz84.Bots.Games;
using sz84.Bots.Users;
using BotWorker.Common;
using BotWorker.Common.Exts;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    //猜大小
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string GetAllIn()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (!CmdPara.In("大", "小", "单", "双", "围", "d", "x", "w", "s", "j", "z", "红", "蓝", "和", "三公", "剪刀", "石头", "布", "抽奖", "庄", "闲") && !CmdPara.IsNum())
            {
                if (CmdPara.Length <= 3)
                    return $"🎁 梭哈 + 大小单双围4-17\n📌 例如：梭哈 大\n         梭哈 9\n💎 积分:{{积分}}全押 ✨";
                else
                    return "";
            }
            if (CmdPara.IsNum())
            {
                long i = CmdPara.AsLong();
                if ((i >= 4) & (i <= 17))
                    CmdName = "押点" + CmdPara;
                else
                    return "点数只能是4到17";
            }
            else
                CmdName = CmdPara;

            long min = Group.BlockMin;
            long credit = UserInfo.GetCredit(GroupId, UserId);
            if (credit < min)
                return $"您的积分{credit}不足{min}";
            
            CmdPara = credit.AsString();
            return GetBlockRes();
        }

        public string GetBlockRes()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (IsTooFast()) return RetryMsgTooFast;

            CmdName = Block.GetCmd(CmdName, UserId);

            if (CmdName.In("押大", "押小", "押单", "押双", "押围", "押全围") && !CmdPara.IsNum())
                return "请押积分，您的积分：{积分}";

            if (CmdName.In("红", "和", "蓝", "庄", "闲"))
                return GetRedBlueRes(GroupId == 10084);

            if (CmdName.In("剪刀", "石头", "布"))            
                return GetCaiquan();

            if (CmdName.In("三公"))
                return GetSanggongRes();

            if (CmdName.In("抽奖"))
                return GetLuckyDraw();

            int blockNum;
            if (CmdName.In("押对", "押点"))
            {
                blockNum = CmdPara.RegexGetValue(Regexs.BlockPara, "BlockNum").AsInt();
                CmdPara = CmdPara.RegexGetValue(Regexs.BlockPara, "cmdPara");

                if ((CmdName == "押对") & ((blockNum < 1) | (blockNum > 6)))
                    return "对数只能是1到6";

                if ((CmdName == "押点") & ((blockNum < 4) | (blockNum > 17)))
                    return "点数只能是4到17";

                CmdName += blockNum.ToString();
            }

            if (!CmdPara.IsNum())
                return "请押积分，您的积分：{积分}";

            long blockCredit = CmdPara.AsLong();
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < blockCredit)
                return $"您只有{creditValue}分";

            int typeId = BlockType.GetTypeId(CmdName);
            blockNum = Block.GetNum(SelfId, GroupId, GroupName, UserId, Name);
            bool isWin = Block.IsWin(typeId, CmdName, blockNum);
            long creditGet = 0;
            long creditAdd;
            if (isWin)
            {
                int odds = Block.GetOdds(typeId, CmdName, blockNum);
                creditAdd = blockCredit * odds;
                creditGet = blockCredit * (odds + 1);
            }
            else
                creditAdd = -blockCredit;

            creditValue += creditAdd;
            var sql = UserInfo.SqlAddCredit(SelfId, GroupId, UserId, creditAdd);
            var sql2 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "猜大小得分");
            var res = $"{Block.FormatNum(blockNum)} {Block.Sum(blockNum)} {Block.GetBlockRes(blockNum)}\n得分：{creditGet:N0}，累计：{creditValue:N0}";
            var blockRes = Message + "\n" + res;
            if (Block.Append(SelfId, GroupId, GroupName, UserId, Name, blockRes, sql, sql2) == -1)
                return RetryMsg;

            if ((IsGroup && Group.IsBlock) || (!IsGroup && User.IsBlock))                
                res = $"{res}\n{(IsGroup ? "群链" : "私链")}：{Block.GetHash(GroupId, UserId)[7..23]}";

            return res;
        }

        public string GetMult()
        {
            if (IsTooFast()) return RetryMsgTooFast;

            long blockCredit;
            string cmdPara;
            long sumCredit = 0;
            var matches = Message.Matches(Regexs.BlockCmdMult);
            foreach (Match match in matches)
            {
                string cmdName = match.Groups["CmdName"].Value;
                cmdPara = match.Groups["cmdPara"].Value;
                blockCredit = cmdPara.AsInt();
                if (blockCredit < Group.BlockMin)
                    return $"至少押{Group.BlockMin}分";
                sumCredit += blockCredit;
            }
            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (creditValue < sumCredit)
                return $"您只有{creditValue}分";

            //生成结果
            int blockNum = Block.GetNum(SelfId, GroupId, GroupName, UserId, Name);
            sumCredit = 0;
            long creditAdd = 0;
            string res = "";
            foreach (Match match in matches)
            {
                string cmdName = match.Groups["CmdName"].Value;
                cmdPara = match.Groups["cmdPara"].Value;
                cmdName = Block.GetCmd(cmdName, UserId);
                blockCredit = cmdPara.AsInt();
                int typeId = BlockType.GetTypeId(cmdName);
                bool isWin = Block.IsWin(typeId, cmdName, blockNum);
                if (isWin)
                {
                    int betOdds = Block.GetOdds(typeId, cmdName, blockNum);
                    creditAdd += blockCredit * betOdds;
                    sumCredit += blockCredit * (betOdds + 1);
                    res += $"{cmdName.Replace("押", "").Replace("全", "")} 得分：{blockCredit * (betOdds + 1):N0}\n";
                }
                else
                    creditAdd -= blockCredit;
            }
            creditValue += creditAdd;
            var sql = UserInfo.SqlAddCredit(SelfId, GroupId, UserId, creditAdd);
            var sql2 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "猜大小得分");
            res = $"{Block.FormatNum(blockNum)} {Block.Sum(blockNum)} {Block.GetBlockRes(blockNum)}\n{res}总得分：{sumCredit:N0} 累计：{creditValue:N0}";
            string block_res = Message + "\n" + res;
            if (Block.Append(SelfId, GroupId, GroupName, UserId, Name, block_res, sql, sql2) == -1)
                return RetryMsg;

            if ((IsGroup && Group.IsBlock) || (!IsGroup && User.IsBlock))
                res = $"{res}\n{(IsGroup ? "群链" : "私链")}：{Block.GetHash(GroupId, UserId)[7..23]}";

            return res;
        }
    }
}
