namespace BotWorker.Domain.Models.BotMessages;

public partial class BotMessage
{
        // 换群
        public string GetChangeGroup() => GetChangeGroupAsync().GetAwaiter().GetResult();
        public async Task<string> GetChangeGroupAsync()
        {
            IsCancelProxy = true;

            if (!CmdPara.IsMatchQQ())
                return "群号不正确，请发命令\n换群 + 新群号";

            if (!await GroupVipRepository.IsVipAsync(GroupId))
                return "体验版无需换群";

            if (!IsRobotOwner())
                return $"你无权换群，你不是群【{GroupId}】机器人主人，";

            long new_groupId = long.Parse(CmdPara);
            if (await GroupVipRepository.IsVipAsync(new_groupId))
                return $"不能换到群【{new_groupId}】，该群已有机器人";

            if (!User.IsSuper)
                return $"非超级分用户不能自己换群，请联系客服QQ处理";

            if (!IsConfirm)
                return await ConfirmMessage("换群将扣除12000分");

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前准确分值（加锁）
                long creditValue = await UserService.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < 12000)
                    return $"您的积分{creditValue}不足12000，换群需扣除12000积分";

                // 2. 扣分
                var (resCode, newValue, _) = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -12000, "换群扣分", wrapper.Transaction);
                if (resCode == -1)
                    throw new Exception("扣分失败");

                // 3. 换群逻辑
                int i = await GroupVipRepository.ChangeGroupAsync(GroupId, new_groupId, UserId, wrapper.Transaction);
                if (i == -1)
                    throw new Exception("换群操作失败");

                await wrapper.CommitAsync();
                return $"✅ 换群成功！将机器人加入新群即可使用\n积分：-12000，累计：{newValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[ChangeGroup Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 换主人
        public string GetChangeOwner() => GetChangeOwnerAsync().GetAwaiter().GetResult();
        public async Task<string> GetChangeOwnerAsync()
        {
            IsCancelProxy = true;

            if (!IsRobotOwner())
                return $"您不是群【{GroupId}】机器人主人，无权换主人";

            if (!CmdPara.IsMatchQQ())
                return $"参数不正确，请发命令 #换主人 + QQ";

            if (!User.IsSuper)
                return $"非超级分用户不能自己换主人，请联系客服QQ处理";

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前准确分值（加锁）
                long creditValue = await UserService.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < 12000)
                    return $"换主人需扣除12000分，您的积分{creditValue}不足";

                // 2. 扣分
                var (resCode, newValue, _) = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -12000, "换主人扣分", wrapper.Transaction);
                if (resCode == -1)
                    throw new Exception("扣分失败");

                // 3. 换主人逻辑
                long newUserId = long.Parse(CmdPara);
                int i = await GroupRepository.SetValueAsync("RobotOwner", newUserId, GroupId, wrapper.Transaction);
                if (i == -1)
                    throw new Exception("修改群机器人主人失败");

                await GroupVipRepository.SetValueAsync("UserId", newUserId, GroupId, wrapper.Transaction);

                await wrapper.CommitAsync();
                return $"✅ 换主人成功！\n积分：-12000，累计：{newValue}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[ChangeOwner Error] {ex.Message}");
                return RetryMsg;
            }
        }

        public async Task<string> GetBuyRobotAsync()
        {
            IsCancelProxy = true;

            string res = SetupPrivate();
            if (res != "")
                return res;

            if (!IsVip)
                return "本群没有开通VIP，余额仅可用于续费";

            if (!CmdPara.IsNum())
                return "📄 格式：续费 + 月数\n📌 例如：续费12\n🔹【续费1】1个月20元\n🔹【续费2】2个月35元\n🔹【续费3】3个月50元\n🔹【续费6】半年80元\n🔹【续费12】一年120元\n🔹【续费24】两年200元\n🔹【续费999】永久498元\n💳 您的余额：{余额}";

            int month = CmdPara.AsInt();
            decimal robotPrice = await PriceRepository.GetRobotPriceAsync(month);
            decimal balance = await UserRepository.GetBalanceAsync(UserId);
            if (balance < robotPrice)
                return $"您的余额{balance:N}不足{robotPrice:N}";

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 余额操作 (自动记录日志)
                var balRes = await UserService.AddBalanceAsync(SelfId, GroupId, GroupName, UserId, Name, -robotPrice, $"群{GroupId}续费{month}个月", wrapper.Transaction);
                if (balRes.Result == -1) throw new Exception("扣除余额失败");

                // 2. 购买机器人 (包含收入记录和 VIP 记录)
                int buyRes = await GroupVipRepository.BuyRobotAsync(SelfId, GroupId, GroupName, UserId, Name, month, robotPrice, "余额", "", "使用余额续费", (int)BotInfo.SystemUid, wrapper.Transaction);
                if (buyRes == -1) throw new Exception("VIP购买失败");

                await wrapper.CommitAsync();

                return $"✅ 群{GroupId}续费{month}个月\n💳 余额：-{robotPrice:N}，累计：{balRes.BalanceValue:N}\n{await GetVipResAsync()}";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Console.WriteLine($"[BuyRobot Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 购买 买入命令分类 买分 买道具 购买一切 根据不同参数调用不同的函数
        public async Task<string> GetBuyResAsync()
        {
            if (CmdPara.Contains("积分"))
            {
                CmdPara = CmdPara.Replace("积分", "").Replace("jf", "").Trim();
                return await UserService.GetBuyCreditResAsync(this, SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            }
            else if (CmdPara == "禁言卡" || CmdPara == "飞机票" || CmdPara == "道具")
                return await GroupPropsRepository.GetBuyResAsync(SelfId, GroupId, GroupName, UserId, Name, CmdPara);
            else
                return await PetService.GetBuyPetAsync(SelfId, GroupId, GroupId, GroupName, UserId, Name, CmdPara);
        }

        // 兑换礼品
        public async Task<string> GetGoodsCreditAsync()
        {
            if (!User.IsSuper)
                return $"仅超级积分可兑换礼品，你的积分类型：{{积分类型}}";

            if (CmdPara == "")
                return "红富士苹果包邮12斤：\n 24个装（中果）：119,520分\n换中果发送【兑换礼品 119520】\n您的{积分类型}：{积分}";

            if (CmdPara != "119520")
                return "参数不正确";

            if (!IsConfirm)
                return await ConfirmMessage("119520分换苹果一箱24个装");

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 获取当前准确分值（加锁）
                long creditValue = await UserService.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);
                if (creditValue < 119520)
                    return $"您的积分{creditValue}不足{119520:N0}";

                // 2. 扣分
                var (resCode, newValue, _) = await UserService.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, -119520, "兑换礼品 苹果一箱24个装（中果）", wrapper.Transaction);
                if (resCode == -1)
                    throw new Exception("扣分失败");

                await wrapper.CommitAsync();
                return "✅ 兑换苹果一箱24个装（中果）成功，请联系客服QQ为您安排发货";
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[GoodsCredit Error] {ex.Message}");
                return RetryMsg;
            }
        }

        // 升级为超级分 
        public async Task<string> GetUpgradeAsync()
        {
            if (!CmdPara.IsMatchQQ())
                return "命令格式：\n升级 + QQ\n例如：\n升级 {客服QQ}";

            if (await PartnerRepository.IsNotPartnerAsync(UserId))
                return "非合伙人无权使用此命令";

            long upgradeQQ = CmdPara.GetAtUserId();
            if (await UserService.GetIsSuperAsync(upgradeQQ))
                return "已为超级积分，无需升级";

            long creditValue = await UserService.GetTotalCreditAsync(SelfId, GroupId, upgradeQQ);
            if (creditValue > 1000)
                return $"该用户有{creditValue}分，升级前请先将原有积分清零";

            // 直接通过 UserRepository 修改
            int res = await UserRepository.UpdateAsync($"is_super=1, super_date=CURRENT_TIMESTAMP, ref_qq={UserId}", upgradeQQ);
            if (res == -1)
                return RetryMsg;

            return $"✅ {upgradeQQ}升级超级积分成功！";
        }

        // 降级为普通分
        public async Task<string> GetCancelSuperAsync()
        {
            if (CmdPara != "")
                return "";

            if (!User.IsSuper)
                return "普通积分无需降级";

            if (IsConfirm && await UserService.GetCreditAsync(SelfId, GroupId, UserId) <= 1000)
            {
                int i = await UserRepository.SetValueAsync("IsSuper", false, UserId);
                return i == -1 ? RetryMsg : "降级成功";
            }
            else
                return await ConfirmMessage("确认降级为普通积分");
        }

        // 版本及有效期
        public async Task<string> GetVipResAsync()
        {
            IsCancelProxy = true;

            if (GroupId == 0 || IsPublic)
            {
                return await GroupVipRepository.GetVipListByUserIdAsync(UserId);
            }

            string version;
            string res;

            if (await GroupVipRepository.GetByIdAsync(GroupId) != null)
            {
                if (await GroupVipRepository.IsYearVIPAsync(GroupId))
                    version = "年费版";
                else
                    version = "VIP版";
                int valid_days = await GroupVipRepository.RestDaysAsync(GroupId);
                if (valid_days >= 1850)
                    res = "『永久版』";
                else
                    res = $"『{version}』有效期：{valid_days}天";
            }
            else
            {
                if (await GroupVipRepository.IsVipOnceAsync(GroupId))
                    return "已过期，请及时续费";
                else
                    version = "体验版";
                res = $"『{version}』";
            }

            return res;
        }
}
