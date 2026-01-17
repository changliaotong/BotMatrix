namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        // 兑换本群积分/金币/紫币等
        public async Task<string> ExchangeCoinsAsync(string cmdPara, string cmdPara2)
        {
            if (string.IsNullOrEmpty(cmdPara2) || !cmdPara2.IsNum())
                return "数量不正确";

            long coinsValue = cmdPara2.AsLong();
            if (coinsValue < 10)
                return "数量最少为10";

            if ((cmdPara == "积分") | (cmdPara == "群积分"))
                cmdPara = "本群积分";

            int coinsType = CoinsLog.conisNames.IndexOf(cmdPara);
            long minusCredit = coinsValue * 120 / 100;

            long creditGroup = GroupId;

            var groupRepo = ServiceProvider!.GetRequiredService<BotWorker.Domain.Repositories.IGroupRepository>();
            var userCreditService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IUserCreditService>();
            var userRepository = ServiceProvider!.GetRequiredService<BotWorker.Domain.Repositories.IUserRepository>();
            var groupMemberService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IGroupMemberService>();

            if (coinsType == (int)CoinsLog.CoinsType.groupCredit)
            {
                if (!await groupRepo.GetIsCreditAsync(GroupId))
                    return "未开启本群积分，无法兑换";
                creditGroup = 0;
            }

            long creditValue = await userCreditService.GetCreditAsync(SelfId, creditGroup, UserId);

            if (await userRepository.GetIsSuperAsync(UserId))
                minusCredit = coinsValue;

            string saveRes = "";

            if (creditValue < minusCredit)
            {
                //兑换本群积分时，可直接扣已存积分
                long creditSave = await userRepository.GetSaveCreditAsync(UserId);
                if ((cmdPara == "本群积分") & (creditSave >= minusCredit - creditValue))
                {
                    var withdrawRes = await DoSaveCreditAsync(creditValue - minusCredit);
                    if (withdrawRes.Result == -1)
                        return withdrawRes.Res;
                    else
                    {
                        creditValue = withdrawRes.CreditValue;
                        creditSave = withdrawRes.CreditSave;
                        saveRes = $"\n取分：{minusCredit - creditValue}，累计：{creditSave}";
                    }
                }
                else
                    return $"您的积分{creditValue}不足{minusCredit}";
            }

            // 使用事务确保原子性
            var exchangeRes = await groupMemberService.ExchangeCoinsAsync(SelfId, GroupId, GroupName, UserId, Name, coinsType, "兑换", cmdPara, minusCredit, coinsValue, UserId);
            if (exchangeRes == RetryMsg) return RetryMsg;
            if (exchangeRes.StartsWith("兑换"))
            {
                // 如果成功了，拼接上取分的消息
                return exchangeRes + saveRes;
            }
            return exchangeRes;
        }

        public async Task<string> GetGiftRes(long userGift, string giftName, int giftCount = 1)
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            var groupGiftService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IGroupGiftService>();
            var giftRepo = ServiceProvider!.GetRequiredService<BotWorker.Domain.Repositories.IGiftRepository>();

            if (CmdPara == "")
                return $"{GroupGift.GiftFormat}\n\n{await giftRepo.GetGiftListAsync(SelfId, GroupId, UserId)}";

            List<string> users = CmdPara.GetValueList(Regexs.Users);
            CmdPara = CmdPara.RegexReplace(Regexs.Users, "");
            List<string> NumList = CmdPara.GetValueList(@"\d{1,4}");
            CmdPara = CmdPara.RegexReplace(@"\d{1,4}", "");
            giftCount = NumList.Count == 0 ? 1 : NumList.First().AsInt();
            giftName = CmdPara;
            string res = "";

            foreach (string user in users)
            {
                userGift = user.AsLong();
                res += await groupGiftService.GetGiftResAsync(SelfId, GroupId, GroupName, UserId, Name, userGift, giftName, giftCount);
            }

            return res;
        }

        // 爱群主
        public async Task<string> GetLampRes()
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            var groupGiftService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IGroupGiftService>();
            var userCreditService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IUserCreditService>();
            var groupRepo = ServiceProvider!.GetRequiredService<BotWorker.Domain.Repositories.IGroupRepository>();
            var userRepo = ServiceProvider!.GetRequiredService<BotWorker.Domain.Repositories.IUserRepository>();

            var fansValue = await groupGiftService.GetFansValueAsync(GroupId, UserId);
            var fansRanking = await groupGiftService.GetFansRankingAsync(GroupId, UserId);
            var fansLevel = await groupGiftService.GetFansLevelAsync(GroupId, UserId);

            var lampTime = groupGiftService.LampMinutes(GroupId, UserId);
            if (lampTime < 10)
                return $"📌 粉丝灯牌已点亮！\n" +
                       $"🧊 冷却时间：{10 - lampTime}分钟\n" +
                       $"💖 亲密度值：{fansValue}\n" +
                       $"🎖️ 粉丝排名：第{fansRanking}名 LV{fansLevel}\n";

            long creditMinus = IsGuild ? RandomInt(1, 1200) : 100;
            long creditAdd = creditMinus / 2;
            long groupOwner = await groupRepo.GetGroupOwnerAsync(GroupId);

            long creditOwner = await userCreditService.GetCreditAsync(SelfId, GroupId, groupOwner);
            creditOwner += creditAdd;
            
            //送灯牌过程：更新灯牌时间、亲密值、积分记录、更新积分、主人积分更新
            if (UserId == creditOwner)
                creditOwner -= creditMinus;

            using var trans = await BeginTransactionAsync();
            try
            {
                var (sql, paras) = groupGiftService.SqlLightLamp(GroupId, UserId);
                await ExecAsync(sql, trans, paras);

                // 1. 给自己加积分 (包含日志记录)
                var res1 = await userCreditService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditMinus, "爱群主", trans);
                if (res1.Result == -1) throw new Exception("更新积分失败");

                // 2. 给群主加积分 (包含日志记录)
                var res2 = await userCreditService.AddCreditAsync(SelfId, GroupId, GroupName, groupOwner, await userRepo.GetRobotOwnerNameAsync(GroupId), creditAdd, "爱群主", trans);
                if (res2.Result == -1) throw new Exception("更新积分失败");

                await trans.CommitAsync();

                return $"🚀 成功点亮粉丝灯牌！\n" +
                  $"💖 亲密指数：+100→{fansValue + 100}\n" +
                  $"💎 群主积分：+{creditAdd}→{res2.CreditValue:N0}\n" +
                  $"🎖️ 粉丝排名：第{fansRanking}名 LV{fansLevel}\n" +
                  $"🧊 冷却时间：10分钟\n" +
                  $"💎 积分：+{creditMinus}，累计：{res1.CreditValue:N0}";
            }
            catch (Exception ex)
            {
                await trans.RollbackAsync();
                Console.WriteLine($"[GetLamp Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 加入粉丝团
        public async Task<string> GetBingFansAsync(string cmdName)
        {
            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            var groupGiftService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IGroupGiftService>();
            var userCreditService = ServiceProvider!.GetRequiredService<BotWorker.Domain.Interfaces.IUserCreditService>();

            if (cmdName == "加团")
            {
                if (await groupGiftService.IsFansAsync(GroupId, UserId))
                    return "您已是粉丝团成员，无需再次加入";

                long creditMinus = 100;
                long creditValue = await userCreditService.GetCreditAsync(SelfId, GroupId, UserId);
                if (creditValue < creditMinus)
                    return $"您的积分{creditValue}不足{creditMinus}加入粉丝团";

                // 使用事务确保原子性
                using var trans = await BeginTransactionAsync();
                try
                {
                    // 1. 更新粉丝团状态
                    var (sql1, paras1) = groupGiftService.SqlBingFans(GroupId, UserId);
                    await ExecAsync(sql1, trans, paras1);

                    // 2. 扣分并记录日志
                    var addRes = await userCreditService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -creditMinus, "加团扣分", trans);
                    if (addRes.Result == -1) throw new Exception("更新积分失败");

                    await trans.CommitAsync();

                    var fansValue = await groupGiftService.GetFansValueAsync(GroupId, UserId);
                    return $"✅ 恭喜您成为第{groupGiftService.GetFansCount(GroupId)}名粉丝团成员\n亲密度值：+100，累计：{fansValue}\n积分：-{creditMinus}，累计：{addRes.CreditValue:N0}";
                }
                catch (Exception ex)
                {
                    await trans.RollbackAsync();
                    Console.WriteLine($"[GetBingFans Error] {ex.Message}");
                    return RetryMsg;
                }
            }
            if (cmdName == "退灯牌")
            {
                if (!await groupGiftService.IsFansAsync(GroupId, UserId))
                    return "您尚未加入粉丝团";

                //退粉丝团
                if (await ExecAsync($"UPDATE {FullName} SET IsFans = 0, FansValue = 0, FansLevel = 0 WHERE GroupId = {GroupId} AND UserId = {UserId}") == -1)
                    return RetryMsg;
                return "✅ 成功退出粉丝团";
            }
            return "";
        }

        // 爱早喵
        public static async Task<string> GetLoveZaomiaoRes()
        {
            //todo 完善爱早喵功能
            return $"早喵也爱你，么么哒";
        }
    }
