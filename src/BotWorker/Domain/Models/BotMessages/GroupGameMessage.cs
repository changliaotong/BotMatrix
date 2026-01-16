namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        public const string ErrorFormat = "命令格式：开盲盒 + 数字1-6\n例如：\n开盲盒 3\nKMH 6";

        // 暗恋系统
        public async Task<string> GetSecretLove()
        {
            string strWhyLove = "\n为什么暗恋那么好？因为暗恋从来不会失恋。\n你一笑我高兴很多天，你一句话我记得好多年。";

            long countLove = await SecretLove.GetCountLoveAsync(UserId);
            long countLoveme = await SecretLove.GetCountLoveMeAsync(UserId);

            if (!CmdPara.IsMatchQQ())
                return "📌 游戏格式：暗恋 + QQ 例如：\n暗恋 {客服QQ}";

            long loveQQ = CmdPara.AsLong();
            if (loveQQ == UserId)
                return "暗恋自己？简称自恋！";

            if (BotInfo.IsRobot(loveQQ))
                return "不要疯狂的迷恋我，我只是个传说！";

            if (await SecretLove.ExistsAsync(UserId, loveQQ))
                return "这个已经暗恋过了，换一个？";

            if (await SecretLove.AppendAsync(SelfId, UserId, loveQQ, RealGroupId) == -1)
                return RetryMsg;

            countLove++;

            if (await SecretLove.IsLoveEachotherAsync(UserId, loveQQ))
            {
                Answer = $"✅ 恭喜你：你暗恋的对象[@:{CmdPara}]刚好也暗恋你，你们可以正大光明地恋爱了！";
                await SendMessageAsync();
            }
            else
                Answer = "✅ 登记成功！若TA也暗恋了你，会通知你们";

            Answer += $"\n你已暗恋{countLove}人，有{countLoveme}人暗恋你。\n{await SecretLove.GetLoveStatusAsync()}{strWhyLove}";   
            return Answer;
        }

        // 猜拳
        public async Task<string> GetCaiquanAsync()
        {
            if (!Group.IsCreditSystem) 
                return CreditSystemClosed;

            if (await IsTooFastAsync()) return RetryMsgTooFast;

            if (!CmdPara.IsNum() || CmdName == "猜拳")
                return "📌 游戏格式：\n石头 {最低积分}\n剪刀 {最低积分}\n布 {最低积分}";

            long blockCredit = CmdPara.AsLong();            
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            using var wrapper = await BeginTransactionAsync();
            try
            {
                long creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < blockCredit)
                {
                    await wrapper.RollbackAsync();
                    return $"您的{Group.CreditName}{creditValue:N0}不足{blockCredit:N0}";
                }

                int iRobot = RandomInt(1, 3);
                long bonus = blockCredit;
                string strRobot = iRobot switch
                {
                    1 => "剪刀",
                    2 => "石头",
                    3 => "布",
                    _ => "剪刀"
                };
                if (strRobot == CmdName)
                {
                    await wrapper.RollbackAsync();
                    return $"✅ 我出{strRobot}, 打平了！";
                }

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

                var (res, newCreditValue, logId) = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, bonus - blockCredit, "猜拳得分", wrapper.Transaction);
                
                if (res == -1)
                {
                    await wrapper.RollbackAsync();
                    return RetryMsg;
                }

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, newCreditValue);

                return $"✅ 我出{strRobot}，你{strWin}了！ \n得分：{bonus}，累计：{newCreditValue:N0}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[GetCaiquan Error] {ex.Message}");
                return RetryMsg;
            }
        }   

        public async Task<string> GetGuessNumAsync()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            string res = "";
            int cszTimes = await UserInfo.GetIntAsync("csz_times", UserId);
            int resCsz = await UserInfo.GetIntAsync("csz_res", UserId);
            long cszCredit = await UserInfo.GetLongAsync("csz_credit", UserId);
            long creditValue;
            if (CmdName == "猜数字")
            {
                //判断上局游戏是否结束
                if (resCsz != -1) return "上局游戏未结束，继续请发 我猜 + 数字";

                using var wrapper = await BeginTransactionAsync();
                try
                {
                    creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                    if (!CmdPara.IsNum())
                    {
                        if (CmdPara == "梭哈")
                            CmdPara = creditValue.ToString();
                        else
                        {
                            await wrapper.RollbackAsync();
                            return $"请押积分！您的{Group.CreditName}{creditValue:N0}";
                        }
                    }

                    long blockCredit = CmdPara.AsLong(); 
                    if (blockCredit < Group.BlockMin)
                    {
                        await wrapper.RollbackAsync();
                        return $"至少押{Group.BlockMin}分";
                    }

                    if (creditValue < blockCredit)
                    {
                        await wrapper.RollbackAsync();
                        return $"您的{Group.CreditName}{creditValue:N0}不足{blockCredit:N0}";
                    }

                    //生成随机数，保存积分以及猜测次数
                    resCsz = RandomInt(1, 13);
                    cszCredit = blockCredit;

                    if (await UserInfo.NewGuessNumGameAsync(resCsz, cszCredit, UserId, wrapper.Transaction) != -1)
                    {
                        //扣分
                        var minusRes = await MinusCreditAsync(cszCredit, "猜数字扣分", wrapper.Transaction);
                        if (minusRes.Result == -1)
                        {
                            await wrapper.RollbackAsync();
                            return "系统出错，请稍后重试";
                        }
                        
                        await wrapper.CommitAsync();

                        // 同步缓存
                        await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, minusRes.CreditValue);

                        return $"您有3次机会，请发送：\n" +
                               $"我猜 + 数字\n-{cszCredit}分，累计：{minusRes.CreditValue:N0}";
                    }
                    else
                    {
                        await wrapper.RollbackAsync();
                        return "系统出错，请稍后重试";
                    }
                }
                catch (Exception ex)
                {
                    await wrapper.RollbackAsync();
                    Logger.Error($"[GetGuessNum Start Error] {ex.Message}");
                    return RetryMsg;
                }
            }
            else if (CmdName == "我猜")
            {
                if (resCsz == -1) return "开始游戏请先发 猜数字 + 积分 ";

                if (!CmdPara.IsNum())
                    return "请猜数字";

                int resGuess = int.Parse(CmdPara);
                if (resGuess < 0 || resGuess > 13)
                    return "请猜 0-13 中的一个数字";

                using var wrapper = await BeginTransactionAsync();
                try
                {
                    if (resCsz == resGuess)
                    {
                        //猜对了结束游戏 加分
                        await UserInfo.UpdateCszGameAsync(-1, 0, 0, UserId, wrapper.Transaction);
                        long creditWin = (cszCredit * 19) / 10;
                        var addRes = await AddCreditAsync(creditWin, "猜数字赢", wrapper.Transaction);
                        
                        if (addRes.Result == -1)
                        {
                            await wrapper.RollbackAsync();
                            return RetryMsg;
                        }

                        await wrapper.CommitAsync();

                        // 同步缓存
                        await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, addRes.CreditValue);

                        return $"✅ 恭喜：{cszTimes + 1}次猜对！\n得分：{creditWin}，累计：{addRes.CreditValue:N0}";
                    }
                    else
                    {
                        //没猜对
                        if (cszTimes == 2)
                        {
                            //结束游戏
                            await UserInfo.UpdateCszGameAsync(-1, 0, 0, UserId, wrapper.Transaction);
                            await wrapper.CommitAsync();
                            return $"您猜错了，正确答案是：{resCsz}";
                        }
                        else
                        {
                            //继续猜
                            await UserInfo.UpdateCszGameAsync(resCsz, cszCredit, cszTimes + 1, UserId, wrapper.Transaction);
                            await wrapper.CommitAsync();
                            if (resCsz > resGuess)
                                return $"✅ 比{resGuess}大，还有{2 - cszTimes}次机会";
                            else
                                return $"✅ 比{resGuess}小，还有{2 - cszTimes}次机会";
                        }
                    }
                }
                catch (Exception ex)
                {
                    await wrapper.RollbackAsync();
                    Logger.Error($"[GetGuessNum Guess Error] {ex.Message}");
                    return RetryMsg;
                }
            }

            return res;
        }

        public async Task<string> GetLuckyDrawAsync()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (await IsTooFastAsync()) return RetryMsgTooFast;

            using var wrapper = await BeginTransactionAsync();
            try
            {
                long creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
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

                // 使用事务执行加分操作
                var (res, newCreditValue, logId) = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditGet, $"抽奖 押{credit}中{bonus}得{creditGet}", wrapper.Transaction);
                
                if (res == -1)
                {
                    await wrapper.RollbackAsync();
                    return RetryMsg;
                }

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, newCreditValue);

                return $"✅ 得分：{bonus}，累计：{newCreditValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[GetLuckyDraw Error] {ex.Message}");
                return RetryMsg;
            }
        }

        public async Task<bool> IsTooFastAsync()
        {
            //频率限制1分钟不能超过6次
            return await CreditLog.CreditCountAsync(UserId, "得分") > 20;
        }

        public async Task<string> GetSanggongResAsync()
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (!CmdPara.IsNum())
            {
                if (CmdPara == "梭哈")
                {
                    CmdPara = (await UserInfo.GetCreditAsync(GroupId, UserId)).ToString();
                }
                else
                    return "🎁 格式：SG + 数值\n" +
                           "📌 例如：SG {最低积分}";
            }
            CmdName = "蓝";

            return await GetRedBlueResAsync(false);
        }

        public async Task<string> GetSanggongRes2Async()
        {
            if (await IsTooFastAsync()) return RetryMsgTooFast;

            using var wrapper = await BeginTransactionAsync();
            try
            {
                long creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (!CmdPara.IsNum())
                {
                    if (CmdPara == "梭哈")
                        CmdPara = creditValue.ToString();
                    else
                    {
                        await wrapper.RollbackAsync();
                        return "格式：SG + 积分数\n例如：SG {最低积分}";
                    }
                }

                long blockCredit = CmdPara.AsLong();
                if (blockCredit < Group.BlockMin)
                {
                    await wrapper.RollbackAsync();
                    return $"至少押{Group.BlockMin}分";
                }
                if (creditValue < blockCredit)
                {
                    await wrapper.RollbackAsync();
                    return $"您只有{creditValue}分";
                }

                string typeName = $"押大";
                int typeId = await BlockType.GetTypeIdAsync(typeName, wrapper.Transaction);
                int blockNum = await BlockRandom.RandomNumAsync(wrapper.Transaction);
                bool isWin = await Block.IsWinAsync(typeId, typeName, blockNum, wrapper.Transaction);
                long creditGet = 0;
                long creditAdd;
                if (isWin)
                {
                    decimal blockOdds = await Block.GetOddsAsync(typeId, typeName, blockNum, wrapper.Transaction);
                    creditAdd = (long)(blockCredit * blockOdds);
                    creditGet = (long)(blockCredit * (blockOdds + 1));
                }
                else
                    creditAdd = -blockCredit;

                var (res, newValue, logId) = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "三公得分", wrapper.Transaction);
                
                if (res == -1)
                {
                    await wrapper.RollbackAsync();
                    return RetryMsg;
                }

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, newValue);

                return $"✅ 得分：{creditGet}，累计：{newValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[GetSanggongRes2 Error] {ex.Message}");
                return RetryMsg;
            }
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

        public async Task<string> GetDouniwanAsync()
        {
            string res = await SetupPrivateAsync(false);
            if (res != "")
                return res;

            if (IsGroup)
                return "请私聊使用此功能";

            if (CmdPara.Trim() == "结束")
                return await UserInfo.SetStateAsync(UserInfo.States.Chat, UserId) == -1
                    ? RetryMsg
                    : "✅ 逗你玩结束";

            //切换到逗你玩状态
            if (CmdPara == "")
            {
                await UserInfo.SetStateAsync(UserInfo.States.Douniwan, UserId);
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
                await MinusCreditResAsync(credit_minus, "逗你玩扣分");

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
