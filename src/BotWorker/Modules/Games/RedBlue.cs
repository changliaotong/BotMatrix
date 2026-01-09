using System.Data;
using System.Security.Cryptography;
using System.Text;
using System.Reflection;

using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Modules.Plugins;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Entities;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.redblue",
        Name = "红蓝博弈",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "经典的红蓝点数博弈游戏，支持押红、押蓝、押和",
        Category = "Games"
    )]
    public class RedBluePlugin : IPlugin
    {
        public BotPluginAttribute Metadata => GetType().GetCustomAttribute<BotPluginAttribute>()!;

        public async Task InitAsync(IRobot robot)
        {
            await EnsureTablesCreatedAsync();
            await robot.RegisterSkillAsync(
                new SkillCapability("红蓝博弈", ["红", "蓝", "和"]),
                HandleRedBlueAsync
            );
        }

        private async Task EnsureTablesCreatedAsync()
        {
            await ShuffledDeck.EnsureTableCreatedAsync();
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleRedBlueAsync(IPluginContext ctx, string[] args)
        {
            // 积分检查
            if (ctx.Group == null || !ctx.Group.IsCreditSystem)
                return "❌ 本群未开启积分系统";

            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var cmdName = ctx.RawMessage.Trim().Split(' ')[0];
            var cmdPara = args.Length > 0 ? args[0] : "";

            long creditValue = UserInfo.GetCredit(groupId, userId);
            
            if (string.IsNullOrEmpty(cmdPara))
                return "请押积分，您的积分：" + creditValue;

            if (cmdPara.ToUpper() == "梭哈" || cmdPara.ToUpper() == "SH")
                cmdPara = creditValue.ToString();

            if (!long.TryParse(cmdPara, out long blockCredit))
                return "押注积分必须是数字";

            if (blockCredit < ctx.Group.BlockMin)
                return $"至少押{ctx.Group.BlockMin}分";

            if (creditValue < blockCredit)
                return $"您只有{creditValue}分";

            List<Card> deck;
            if (!await ShuffledDeck.IsShuffledDeckExistsAsync(groupId))
            {
                deck = RedBlue.InitializeDeck();
                RedBlue.ShuffleDeck(deck);
                await ShuffledDeck.SaveShuffledDeckAsync(groupId, deck);
            }
            else
            {
                deck = await ShuffledDeck.ReadShuffledDeckAsync(groupId);
            }

            // 发牌
            List<Card> playerHand = [deck[0], deck[2]];
            List<Card> bankerHand = [deck[1], deck[3]];

            string result;
            int payout;

            bool isNatural = RedBlue.HasNatural(playerHand) || RedBlue.HasNatural(bankerHand);
            if (!isNatural)
            {
                int playerTotal = RedBlue.CalculateTotal(playerHand);
                int bankerTotal = RedBlue.CalculateTotal(bankerHand);
                int playerThirdCard = -1;

                if (playerTotal <= 5)
                {
                    playerHand.Add(deck[4]);
                    playerThirdCard = RedBlue.CalculatePoint(deck[4]);
                    playerTotal = RedBlue.CalculateTotal(playerHand);
                }

                bool bankerDrawCard = false;
                if (playerThirdCard != -1)
                {
                    if (bankerTotal <= 2) bankerDrawCard = true;
                    else if (bankerTotal == 3 && playerThirdCard != 8) bankerDrawCard = true;
                    else if (bankerTotal == 4 && (playerThirdCard >= 2 && playerThirdCard <= 7)) bankerDrawCard = true;
                    else if (bankerTotal == 5 && (playerThirdCard >= 4 && playerThirdCard <= 7)) bankerDrawCard = true;
                    else if (bankerTotal == 6 && (playerThirdCard == 6 || playerThirdCard == 7)) bankerDrawCard = true;
                }
                else if (bankerTotal <= 5)
                {
                    bankerDrawCard = true;
                }

                if (bankerDrawCard)
                    bankerHand.Add(deck[5]);
            }

            result = RedBlue.CalculateResult(playerHand, bankerHand);
            payout = RedBlue.CalculatePayout(result);

            foreach (var card in playerHand.Concat(bankerHand))
                await ShuffledDeck.ClearShuffledDeckAsync(groupId, card.Id);

            var sb = new StringBuilder();
            sb.AppendLine($"蓝：{string.Join(" ", playerHand.Select(c => c.Suit + c.Rank))}【{RedBlue.CalculateTotal(playerHand)}】");
            sb.AppendLine($"红：{string.Join(" ", bankerHand.Select(c => c.Suit + c.Rank))}【{RedBlue.CalculateTotal(bankerHand)}】");
            sb.AppendLine($"结果：{result}");

            bool isWin = result.Contains(cmdName);
            long creditAdd = 0;
            if (isWin)
            {
                if (cmdName == "红" && RedBlue.CalculateTotal(bankerHand) == 6)
                    creditAdd = blockCredit / 2;
                else
                    creditAdd = blockCredit * payout;
            }
            else if (result == "和")
                creditAdd = 0;
            else
                creditAdd = -blockCredit;

            if (creditAdd != 0)
            {
                var addRes = await UserInfo.AddCreditAsync(long.Parse(ctx.BotId), groupId, ctx.Group.GroupName, userId, ctx.User?.Name ?? "", creditAdd, "红和蓝得分");
                creditValue = addRes.CreditValue;
            }

            sb.Append($"✅ 得分：{(isWin ? blockCredit + creditAdd : (result == "和" ? blockCredit : 0))}，累计：{creditValue}");

            if (deck.Count < 6)
            {
                deck = RedBlue.InitializeDeck();
                RedBlue.ShuffleDeck(deck);
                await ShuffledDeck.SaveShuffledDeckAsync(groupId, deck);
            }

            return sb.ToString();
        }
    }

    public static class RedBlue
    {
        //static readonly string[] suits = { "红桃", "方块", "梅花", "黑桃" };
        static readonly string[] suits = { "♥️", "♦️", "♣", "♠" };
        static readonly string[] ranks = { "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A" };


        public static string ComputeNextGameHash(List<Card> deck)
        {
            StringBuilder data = new();
            foreach (var card in deck)
            {
                data.Append(card.Rank).Append(card.Suit);
            }
            byte[] hashBytes = SHA384.HashData(Encoding.UTF8.GetBytes(data.ToString()));
            return BitConverter.ToString(hashBytes).Replace("-", "");
        }

        public static List<Card> InitializeDeck()
        {
            List<Card> deck = [];

            int j = 0;
            foreach (var suit in suits)
            {
                foreach (var rank in ranks)
                {
                    for (int i = 0; i < 8; i++) // 循环8次
                    {
                        j++;
                        deck.Add(new Card(j, rank, suit));
                    }
                }
            }

            return deck;
        }

        public static void ShuffleDeck(List<Card> deck)
        {
            Random rng = new();
            int n = deck.Count;
            while (n > 1)
            {
                n--;
                int k = rng.Next(n + 1);
                (deck[n], deck[k]) = (deck[k], deck[n]);
            }
        }

        public static int CalculateTotal(List<Card> hand)
        {
            int total = 0;
            foreach (var card in hand)
            {
                total += CalculatePoint(card);
            }
            return total % 10; // 只保留个位数
        }

        public static int CalculatePoint(Card card)
        {
            if (card.Rank == "J" || card.Rank == "Q" || card.Rank == "K")
            {
                return 10;
            }
            else if (card.Rank == "A")
            {
                return 1; // A先算1点
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

            if (playerTotal > bankerTotal)
            {
                return "蓝赢";
            }
            else if (playerTotal < bankerTotal)
            {
                return "红赢";
            }
            else
            {
                return "和";
            }
        }

        public static int CalculatePayout(string result)
        {
            if (result == "蓝赢" || result == "红赢")
            {
                return 1; // 1倍
            }
            else
            {
                return 8; // 8倍
            }
        }

    }

    public class Card(int id, string rank, string suit)
    {
        public int Id { get; set; } = id;
        public string Rank { get; } = rank;
        public string Suit { get; } = suit;
    }

    class ShuffledDeck : MetaData<ShuffledDeck>
    {
        public override string TableName => "ShuffledDeck";
        public override string KeyField => "DeckId";

        public static void ClearShuffledDeck(long groupId)
            => ClearShuffledDeckAsync(groupId).GetAwaiter().GetResult();

        public static async Task ClearShuffledDeckAsync(long groupId)
        {
            string sql = $"DELETE FROM {FullName} WHERE GroupId = {groupId}";
            await ExecAsync(sql);
        }

        public static void ClearShuffledDeck(long groupId, long id)
            => ClearShuffledDeckAsync(groupId, id).GetAwaiter().GetResult();

        public static async Task ClearShuffledDeckAsync(long groupId, long id)
        {
            string sql = $"DELETE FROM {FullName} WHERE GroupId = {groupId} and Id = {id}";
            await ExecAsync(sql);
        }

        public static void SaveShuffledDeck(long groupId, List<Card> deck)
            => SaveShuffledDeckAsync(groupId, deck).GetAwaiter().GetResult();

        public static async Task SaveShuffledDeckAsync(long groupId, List<Card> deck)
        {
            await ClearShuffledDeckAsync(groupId);
            for (int i = 0; i < deck.Count; i++)
                await InsertAsync([
                    new Cov("GroupId", groupId),
                    new Cov("Id", deck[i].Id),
                    new Cov("Rank", deck[i].Rank),
                    new Cov("Suit", deck[i].Suit),
                    new Cov("DeckOrder", i),
                ]);
        }

        public static bool IsShuffledDeckExists(long groupId)
            => IsShuffledDeckExistsAsync(groupId).GetAwaiter().GetResult();

        public static async Task<bool> IsShuffledDeckExistsAsync(long groupId)
        {
            return await CountWhereAsync($"groupId = {groupId}") >= 6;
        }

        public static List<Card> ReadShuffledDeck(long groupId)
            => ReadShuffledDeckAsync(groupId).GetAwaiter().GetResult();

        public static async Task<List<Card>> ReadShuffledDeckAsync(long groupId)
        {
            List<Card> deck = [];
            string query = $"SELECT Id, Rank, Suit FROM {FullName} WHERE groupId = @groupId ORDER BY DeckOrder";
            
            var ds = await QueryDatasetAsync(query, null, CreateParameter("@groupId", groupId));
            if (ds != null && ds.Tables.Count > 0)
            {
                foreach (DataRow row in ds.Tables[0].Rows)
                {
                    deck.Add(new Card(
                        Convert.ToInt32(row[0]),
                        row[1].ToString() ?? "",
                        row[2].ToString() ?? ""
                    ));
                }
            }
            return deck;
        }
    }
}
