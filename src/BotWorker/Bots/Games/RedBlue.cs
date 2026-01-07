using Microsoft.Data.SqlClient;
using System.Security.Cryptography;
using System.Text;
using BotWorker.Common.Exts;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Games
{
    internal class RedBlue
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

    class Card(int id, string rank, string suit)
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
        {
            string sql = $"DELETE FROM {FullName} WHERE GroupId = {groupId}";
            Exec(sql);
        }

        public static void ClearShuffledDeck(long groupId, long id)
        {
            string sql = $"DELETE FROM {FullName} WHERE GroupId = {groupId} and Id = {id}";
            Exec(sql);
        }

        public static void SaveShuffledDeck(long groupId, List<Card> deck)
        {
            ClearShuffledDeck(groupId);
            for (int i = 0; i < deck.Count; i++)
                Insert([
                    new Cov("GroupId", groupId),
                    new Cov("Id", deck[i].Id),
                    new Cov("Rank", deck[i].Rank),
                    new Cov("Suit", deck[i].Suit),
                    new Cov("DeckOrder", i),
                ]);
        }

        public static bool IsShuffledDeckExists(long groupId)
        {
            return CountWhere($"groupId = {groupId}") >= 6;
        }

        public static List<Card> ReadShuffledDeck(long groupId)
        {
            List<Card> deck = [];
            string query = $"SELECT Id, Rank, Suit FROM {FullName} WHERE groupId = @groupId ORDER BY DeckOrder";
            foreach (var reader in QueryReader(query, new SqlParameter("@groupId", groupId)))
            {
                while (reader.Read())
                {
                    deck.Add(new Card(
                        reader.GetInt32(0),
                        reader.GetString(1),
                        reader.GetString(2)
                    ));
                }
            }
            return deck;
        }

    }
}
