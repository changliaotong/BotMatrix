using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Reflection;
using System.Security.Cryptography;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using BotWorker.Modules.Plugins;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.redblue",
        Name = "红蓝战士",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "经典的红蓝点数博弈游戏，支持押红、押蓝、押和",
        Category = "Games"
    )]
    public class RedBluePlugin : IPlugin
    {
        private readonly IUserCreditService _creditService;
        private readonly IShuffledDeckRepository _shuffledDeckRepo;
        private readonly ILogger<RedBluePlugin> _logger;

        public RedBluePlugin(
            IUserCreditService creditService, 
            IShuffledDeckRepository shuffledDeckRepo,
            ILogger<RedBluePlugin> logger)
        {
            _creditService = creditService;
            _shuffledDeckRepo = shuffledDeckRepo;
            _logger = logger;
        }

        public BotPluginAttribute Metadata => GetType().GetCustomAttribute<BotPluginAttribute>()!;

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(
                new SkillCapability("红蓝博弈", ["红", "蓝", "和"]),
                HandleRedBlueAsync
            );
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleRedBlueAsync(IPluginContext ctx, string[] args)
        {
            // 积分检查
            if (ctx.Group == null || !ctx.Group.IsCreditSystem)
                return "❌ 本群未开启积分系统";

            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var botId = long.Parse(ctx.BotId);
            var cmdName = ctx.RawMessage.Trim().Split(' ')[0];
            var cmdPara = args.Length > 0 ? args[0] : "";

            using var wrapper = await _shuffledDeckRepo.BeginTransactionAsync();
            try
            {
                // 1. 获取积分并锁定用户
                long creditValue = await _creditService.GetCreditForUpdateAsync(botId, groupId, userId, wrapper.Transaction);

                if (string.IsNullOrEmpty(cmdPara))
                {
                    await wrapper.RollbackAsync();
                    return $"请押积分，您的积分：{creditValue:N0}";
                }

                if (cmdPara.ToUpper().In("梭哈", "SH"))
                    cmdPara = creditValue.ToString();

                if (!long.TryParse(cmdPara, out long blockCredit))
                {
                    await wrapper.RollbackAsync();
                    return "押注积分必须是数字";
                }

                if (blockCredit < ctx.Group.BlockMin)
                {
                    await wrapper.RollbackAsync();
                    return $"至少押{ctx.Group.BlockMin}分";
                }

                if (creditValue < blockCredit)
                {
                    await wrapper.RollbackAsync();
                    return $"您只有{creditValue:N0}分";
                }

                // 2. 加载并锁定牌堆
                List<Card> deck = await _shuffledDeckRepo.ReadShuffledDeckAsync(groupId, wrapper.Transaction, true);

                if (deck.Count < 6)
                {
                    deck = RedBlue.InitializeDeck();
                    RedBlue.ShuffleDeck(deck);
                    await SaveShuffledDeckAsync(groupId, deck, wrapper.Transaction);
                }

                // 发牌
                List<Card> playerHand = [deck[0], deck[2]];
                List<Card> bankerHand = [deck[1], deck[3]];

                string result;
                int payout;

                // 检查天牌
                if (RedBlue.HasNatural(playerHand) || RedBlue.HasNatural(bankerHand))
                {
                    result = RedBlue.CalculateResult(playerHand, bankerHand);
                    payout = RedBlue.CalculatePayout(result);
                }
                else
                {
                    // 补牌逻辑（简化版）
                    if (RedBlue.CalculateTotal(playerHand) <= 5)
                        playerHand.Add(deck[4]);
                    
                    if (RedBlue.CalculateTotal(bankerHand) <= 5)
                        bankerHand.Add(deck[5]);

                    result = RedBlue.CalculateResult(playerHand, bankerHand);
                    payout = RedBlue.CalculatePayout(result);
                }

                // 3. 计算收益并更新积分
                bool isWin = (cmdName == "蓝" && result == "蓝赢") || 
                             (cmdName == "红" && result == "红赢") || 
                             (cmdName == "和" && result == "和");

                long profit = isWin ? blockCredit * payout : -blockCredit;
                
                var addRes = await _creditService.AddCreditAsync(botId, groupId, ctx.GroupName ?? "", userId, ctx.UserName, profit, $"红蓝博弈:{cmdName}", wrapper.Transaction);
                
                // 4. 移除已使用的牌
                List<int> usedIds = playerHand.Concat(bankerHand).Select(c => c.Id).ToList();
                await _shuffledDeckRepo.ClearShuffledDeckAsync(groupId, usedIds, wrapper.Transaction);

                await wrapper.CommitAsync();

                // 5. 构建响应消息
                StringBuilder sb = new();
                sb.AppendLine($"【红蓝博弈】结果：{result}");
                sb.AppendLine($"蓝方：{string.Join(" ", playerHand.Select(c => $"[{c.Suit}{c.Rank}]"))} ({RedBlue.CalculateTotal(playerHand)}点)");
                sb.AppendLine($"红方：{string.Join(" ", bankerHand.Select(c => $"[{c.Suit}{c.Rank}]"))} ({RedBlue.CalculateTotal(bankerHand)}点)");
                sb.AppendLine("------------------");
                sb.AppendLine(isWin ? $"💰 恭喜！您赢得了 {profit:N0} 积分" : $"💸 很遗憾，您输掉了 {blockCredit:N0} 积分");
                sb.Append($"当前积分：{addRes.CreditValue:N0}");

                return sb.ToString();
            }
            catch (Exception ex)
            {
                await wrapper.RollbackAsync();
                _logger.LogError(ex, "红蓝博弈发生异常");
                return $"❌ 游戏发生异常：{ex.Message}";
            }
        }

        private async Task SaveShuffledDeckAsync(long groupId, List<Card> deck, IDbTransaction trans)
        {
            await _shuffledDeckRepo.ClearShuffledDeckAsync(groupId, trans);
            foreach (var (card, i) in deck.Select((c, i) => (c, i)))
            {
                var item = new ShuffledDeck
                {
                    GroupId = groupId,
                    Id = card.Id,
                    Rank = card.Rank,
                    Suit = card.Suit,
                    DeckOrder = i
                };
                await trans.Connection.InsertAsync(item, trans);
            }
        }
    }

    public static class RedBlue
    {
        public static List<Card> InitializeDeck()
        {
            string[] suits = ["♠", "♥", "♣", "♦"];
            string[] ranks = ["A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"];
            List<Card> deck = [];
            int id = 0;
            foreach (var suit in suits)
            {
                foreach (var rank in ranks)
                {
                    deck.Add(new Card(++id, rank, suit));
                }
            }
            return deck;
        }

        public static void ShuffleDeck(List<Card> deck)
        {
            for (int i = deck.Count - 1; i > 0; i--)
            {
                int j = RandomNumberGenerator.GetInt32(i + 1);
                (deck[i], deck[j]) = (deck[j], deck[i]);
            }
        }

        public static int CalculateTotal(List<Card> hand)
        {
            int total = hand.Sum(CalculatePoint);
            return total % 10;
        }

        public static int CalculatePoint(Card card)
        {
            if (card.Rank == "J" || card.Rank == "Q" || card.Rank == "K")
            {
                return 10;
            }
            else if (card.Rank == "A")
            {
                return 1;
            }
            else
            {
                return int.Parse(card.Rank);
            }
        }

        public static bool HasNatural(List<Card> hand)
        {
            int total = CalculateTotal(hand);
            return total == 8 || total == 9;
        }

        public static string CalculateResult(List<Card> playerHand, List<Card> bankerHand)
        {
            int playerTotal = CalculateTotal(playerHand);
            int bankerTotal = CalculateTotal(bankerHand);

            if (playerTotal > bankerTotal) return "蓝赢";
            if (playerTotal < bankerTotal) return "红赢";
            return "和";
        }

        public static int CalculatePayout(string result)
        {
            return result == "和" ? 8 : 1;
        }
    }

    public class Card(int id, string rank, string suit)
    {
        public int Id { get; set; } = id;
        public string Rank { get; set; } = rank;
        public string Suit { get; set; } = suit;
    }

    [Table("shuffled_deck")]
    public class ShuffledDeck
    {
        [ExplicitKey]
        public long GroupId { get; set; }
        [ExplicitKey]
        public int Id { get; set; }
        public string Rank { get; set; } = string.Empty;
        public string Suit { get; set; } = string.Empty;
        public int DeckOrder { get; set; }
    }
}
