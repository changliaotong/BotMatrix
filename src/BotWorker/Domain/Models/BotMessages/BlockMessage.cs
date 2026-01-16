using System.Text.RegularExpressions;
using System.Threading.Tasks;

namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public async Task<string> GetAllInAsync()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (!CmdPara.In("大", "小", "单", "双", "围", "d", "x", "w", "s", "j", "z", "红", "蓝", "和", "三公", "剪刀", "石头", "布", "抽奖", "庄", "闲") && !CmdPara.IsNum())
            {
                if (CmdPara.Length <= 3)
                    return $"🎁 梭哈 + 大小单双围4-17\n📌 例如：梭哈 大\n         梭哈 9\n💎 {{积分类型}}:{{积分}}全押 ✨";
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
            long credit = await UserInfo.GetCreditAsync(GroupId, UserId);
            if (credit < min)
                return $"您的积分{credit}不足{min}";
            
            CmdPara = credit.AsString();
            return await GetBlockResAsync();
        }

        public async Task<string> GetBlockResAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (IsTooFast()) return RetryMsgTooFast;

            CmdName = await Block.GetCmdAsync(CmdName, UserId);

            if (CmdName.In("押大", "押小", "押单", "押双", "押围", "押全围") && !CmdPara.IsNum())
                return "请押积分，您的{{积分类型}}：{{积分}}";

            if (CmdName.In("红", "和", "蓝", "庄", "闲"))
                return await GetRedBlueResAsync(GroupId == 10084);

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
                return "请押积分，您的{{积分类型}}：{{积分}}";

            long blockCredit = CmdPara.AsLong();
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            long creditValue = await UserInfo.GetCreditAsync(GroupId, UserId);
            if (creditValue < blockCredit)
                return $"您只有{{积分}}分";

            int typeId = await BlockType.GetTypeIdAsync(CmdName);
            blockNum = await Block.GetNumAsync(SelfId, GroupId, GroupName, UserId, Name);
            bool isWin = await Block.IsWinAsync(typeId, CmdName, blockNum);
            long creditGet = 0;
            long creditAdd;
            if (isWin)
            {
                decimal odds = await Block.GetOddsAsync(typeId, CmdName, blockNum);
                creditAdd = (long)(blockCredit * odds);
                creditGet = (long)(blockCredit * (odds + 1));
            }
            else
                creditAdd = -blockCredit;

            creditValue += creditAdd;

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取并锁定积分
                creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < blockCredit)
                {
                    wrapper.Rollback();
                    return $"您只有{creditValue}分";
                }

                // 2. 通用加积分函数（含日志记录）
                var addRes = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "猜大小得分", wrapper.Transaction);
                if (addRes.Result == -1) throw new Exception("更新积分失败");
                creditValue = addRes.CreditValue;

                // 3. 记录游戏记录
                var resStr = $"{Block.FormatNum(blockNum)} {Block.Sum(blockNum)} {Block.GetBlockRes(blockNum)}\n得分：{creditGet:N0}，累计：{creditValue:N0}";
                var blockRes = Message + "\n" + resStr;
                var (sql3, sql4) = await Block.SqlAppendAsync(SelfId, GroupId, GroupName, UserId, Name, blockRes, wrapper.Transaction);
                await ExecAsync(sql3.sql, wrapper.Transaction, sql3.paras);
                await ExecAsync(sql4.sql, wrapper.Transaction, sql4.paras);

                wrapper.Commit();

                // 4. 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, creditValue);

                if ((IsGroup && Group.IsBlock) || (!IsGroup && User.IsBlock))
                    resStr = $"{resStr}\n{(IsGroup ? "群链" : "私链")}：{(await Block.GetHashAsync(GroupId, UserId, wrapper.Transaction))[7..23]}";

                return resStr;
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[GetBlockRes Error] {ex.Message}");
                return RetryMsg;
            }
    }

    public async Task<string> GetMultAsync()
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
            long creditValue = await UserInfo.GetCreditAsync(GroupId, UserId);
            if (creditValue < sumCredit)
                return $"您只有{creditValue}分";

            //生成结果
            int blockNum = await Block.GetNumAsync(SelfId, GroupId, GroupName, UserId, Name);
            sumCredit = 0;
            long creditAdd = 0;
            string res = "";
            foreach (Match match in matches)
            {
                string cmdName = match.Groups["CmdName"].Value;
                cmdPara = match.Groups["cmdPara"].Value;
                cmdName = await Block.GetCmdAsync(cmdName, UserId);
                blockCredit = cmdPara.AsInt();
                int typeId = await BlockType.GetTypeIdAsync(cmdName);
                bool isWin = await Block.IsWinAsync(typeId, cmdName, blockNum);
                if (isWin)
                {
                    decimal betOdds = await Block.GetOddsAsync(typeId, cmdName, blockNum);
                    creditAdd += (long)(blockCredit * betOdds);
                    sumCredit += (long)(blockCredit * (betOdds + 1));
                    res += $"{cmdName.Replace("押", "").Replace("全", "")} 得分：{blockCredit * (betOdds + 1):N0}\n";
                }
                else
                    creditAdd -= blockCredit;
            }
            creditValue += creditAdd;

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取并锁定积分
                creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < sumCredit)
                {
                    wrapper.Rollback();
                    return $"您只有{creditValue}分";
                }

                // 2. 通用加积分函数（含日志记录）
                var addRes = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "猜大小得分", wrapper.Transaction);
                if (addRes.Result == -1) throw new Exception("更新积分失败");
                creditValue = addRes.CreditValue;

                // 3. 记录游戏记录
                res = $"{Block.FormatNum(blockNum)} {Block.Sum(blockNum)} {Block.GetBlockRes(blockNum)}\n{res}总得分：{sumCredit:N0} 累计：{creditValue:N0}";
                string block_res = Message + "\n" + res;
                var (sql3, sql4) = await Block.SqlAppendAsync(SelfId, GroupId, GroupName, UserId, Name, block_res, wrapper.Transaction);
                await ExecAsync(sql3.sql, wrapper.Transaction, sql3.paras);
                await ExecAsync(sql4.sql, wrapper.Transaction, sql4.paras);

                wrapper.Commit();

                // 4. 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, creditValue);

                if ((IsGroup && Group.IsBlock) || (!IsGroup && User.IsBlock))
                    res = $"{res}\n{(IsGroup ? "群链" : "私链")}：{(await Block.GetHashAsync(GroupId, UserId, wrapper.Transaction))[7..23]}";

                return res;
            }
            catch (Exception ex)
            {
                wrapper.Rollback();
                Logger.Error($"[GetMult Error] {ex.Message}");
                return RetryMsg;
            }
        }
}
