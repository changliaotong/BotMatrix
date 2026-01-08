using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Models.Messages.BotMessages
{
    //红和蓝
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string GetRedBlueRes(bool isDetail = true)
        {
            IsCancelProxy = true;

            if (!Group.IsCreditSystem)
                return CreditSystemClosed;

            if (IsTooFast())
                return RetryMsgTooFast;

            long creditValue = UserInfo.GetCredit(GroupId, UserId);
            if (!CmdPara.IsNum())
            {
                if (CmdPara.ToUpper().In("梭哈", "SH"))
                    CmdPara = creditValue.ToString();
                else
                    return "请押积分，您的积分：{积分}";
            }

            long blockCredit = CmdPara.AsLong();
            if (blockCredit < Group.BlockMin)
                return $"至少押{Group.BlockMin}分";

            if (creditValue < blockCredit)
                return $"您只有{creditValue}分";

            string res = "";
            List<Card> deck;

            // 检查数据库中是否存在已洗好的牌堆，如果不存在，则进行洗牌并保存到数据库
            if (!ShuffledDeck.IsShuffledDeckExists(GroupId))
            {
                deck = RedBlue.InitializeDeck();
                RedBlue.ShuffleDeck(deck);
                ShuffledDeck.SaveShuffledDeck(GroupId, deck);
            }
            else
                deck = ShuffledDeck.ReadShuffledDeck(GroupId);

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
            for (int i = 0; i < playerHand.Count + bankerHand.Count; i++)
                ShuffledDeck.ClearShuffledDeck(GroupId, deck[i].Id);

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

            creditValue += creditAdd;

            if (creditAdd != 0)
            {
                var sql = UserInfo.SqlAddCredit(SelfId, GroupId, UserId, creditAdd);
                var sql2 = CreditLog.SqlHistory(SelfId, GroupId, GroupName, UserId, Name, creditAdd, "红和蓝得分");
                if (ExecTrans(sql, sql2) == -1)
                    return RetryMsg;
            }

            res += $"✅ 得分：{creditGet}，累计：{creditValue}";

            // 判断是否需要重新洗牌
            if (deck.Count < 6)
            {
                deck = RedBlue.InitializeDeck();
                RedBlue.ShuffleDeck(deck);
                ShuffledDeck.SaveShuffledDeck(GroupId, deck);
            }

            // 发布下一局游戏的 Hash
            //string nextGameHash = ComputeNextGameHash(deck.GetRange(0, 6)); // 发布下一局游戏要发的牌
            //res += $"下一局游戏的 Hash: {nextGameHash}";

            return res;

        }
    }
}
