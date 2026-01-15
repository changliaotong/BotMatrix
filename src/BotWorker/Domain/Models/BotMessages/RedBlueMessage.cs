namespace BotWorker.Domain.Models.BotMessages
{
    //红和蓝
    public partial class BotMessage : MetaData<BotMessage>
    {
        public async Task<string> GetRedBlueResAsync(bool isDetail = true)
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (await IsTooFastAsync())
                return RetryMsgTooFast;

            using var wrapper = await BeginTransactionAsync();
            try
            {
                // 1. 在事务内获取积分并锁定用户
                long creditValue = await UserInfo.GetCreditForUpdateAsync(SelfId, GroupId, UserId, wrapper.Transaction);

                // 2. 在事务内加载并锁定牌堆
                List<Card> deck = await ShuffledDeck.ReadShuffledDeckAsync(GroupId, wrapper.Transaction, true);

                if (deck.Count < 6)
                {
                    deck = RedBlue.InitializeDeck();
                    RedBlue.ShuffleDeck(deck);
                    await ShuffledDeck.SaveShuffledDeckAsync(GroupId, deck, wrapper.Transaction);
                }

                if (!CmdPara.IsNum())
                {
                    if (CmdPara.ToUpper().In("梭哈", "SH"))
                        CmdPara = creditValue.ToString();
                    else
                    {
                        await wrapper.RollbackAsync();
                        return $"请押积分，您的积分：{creditValue:N0}";
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
                    return $"您只有{creditValue:N0}分";
                }

                string res = "";

                // 发牌
                List<Card> playerHand = [deck[0], deck[2]]; // 第1张和第3张牌给红方
                List<Card> bankerHand = [deck[1], deck[3]]; // 第2张和第4张牌给蓝方

                string result;
                int payout;

                // 判断是否出现自然赢
                bool isNatural = RedBlue.HasNatural(playerHand) || RedBlue.HasNatural(bankerHand);
                if (!isNatural)
                {
                    // 判断是否需要蓝方补牌
                    bool playerDrawCard = false;
                    int playerThirdCard = 0;
                    int playerTotal = RedBlue.CalculateTotal(playerHand);
                    int bankerTotal = RedBlue.CalculateTotal(bankerHand);
                    if (playerTotal <= 5)
                    {
                        playerDrawCard = true;
                    }

                    // 如果蓝方补牌，重新计算点数
                    if (playerDrawCard)
                    {
                        playerHand.Add(deck[4]); // 补一张牌
                        playerThirdCard = RedBlue.CalculatePoint(deck[4]);
                        playerTotal = RedBlue.CalculateTotal(playerHand);
                    }

                    // 判断是否需要红方补牌
                    bool bankerDrawCard = false;
                    if (playerDrawCard)
                    {
                        if (bankerTotal <= 5)
                        {
                            bankerDrawCard = true;
                        }
                    }
                    else
                    {
                        // 判断红方是否赢
                        bool bankerWin = bankerTotal > playerTotal;

                        if (!bankerWin)
                        {

                            if (bankerTotal <= 2)
                            {
                                bankerDrawCard = true;
                            }
                            else if (bankerTotal == 3 && playerThirdCard != 8)
                            {
                                bankerDrawCard = true;
                            }
                            else if (bankerTotal == 4 && (playerThirdCard >= 2 && playerThirdCard <= 7))
                            {
                                bankerDrawCard = true;
                            }
                            else if (bankerTotal == 5 && (playerThirdCard >= 4 && playerThirdCard <= 7))
                            {
                                bankerDrawCard = true;
                            }
                            else if (bankerTotal == 6 && (playerThirdCard == 6 || playerThirdCard == 7))
                            {
                                bankerDrawCard = true;
                            }
                        }
                    }

                    // 补发剩余的牌
                    if (bankerDrawCard)
                        bankerHand.Add(deck[5]);
                }
                // 结算
                result = RedBlue.CalculateResult(playerHand, bankerHand);
                payout = RedBlue.CalculatePayout(result);
                // 批量清除已使用的牌（减少数据库交互次数）
                var usedCardIds = playerHand.Concat(bankerHand).Select(c => c.Id).ToList();
                await ShuffledDeck.ClearShuffledDeckAsync(GroupId, usedCardIds, wrapper.Transaction);

                var bres = $"蓝：{string.Concat(playerHand.Select(card => " " + card.Suit + card.Rank)).Trim()}【{RedBlue.CalculateTotal(playerHand)}】\n";
                bres += $"红：{string.Concat(bankerHand.Select(card => " " + card.Suit + card.Rank)).Trim()}【{RedBlue.CalculateTotal(bankerHand)}】\n";
                bres += $"结果：{result}\n";

                // 显示结果
                if (isDetail) res += bres;

                bool isWin = result.Contains(CmdName);
                long creditGet = 0;
                long creditAdd;
                if (isWin)
                {
                    int odds = payout;
                    if (CmdName == "红" && RedBlue.CalculateTotal(bankerHand) == 6)
                        creditAdd = blockCredit / 2;
                    else
                        creditAdd = blockCredit * odds;
                    creditGet = blockCredit + creditAdd;
                }
                else
                {
                    if (result == "和")
                    {
                        creditGet = blockCredit;
                        creditAdd = 0;
                    }
                    else
                        creditAdd = -blockCredit;
                }

                if (creditAdd != 0)
                {
                    var addRes = await UserInfo.AddCreditAsync(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "红和蓝得分", wrapper.Transaction);
                    if (addRes.Result == -1)
                    {
                        await wrapper.RollbackAsync();
                        return RetryMsg;
                    }
                    creditValue = addRes.CreditValue;
                }

                await wrapper.CommitAsync();

                // 同步缓存
                await UserInfo.SyncCreditCacheAsync(SelfId, GroupId, UserId, creditValue);

                res += $"✅ 得分：{creditGet:N0}，累计：{creditValue:N0}";

                return res;
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                Logger.Error($"[GetRedBlueRes Error] {ex.Message}");
                return RetryMsg;
            }

        }
    }
}
