using BotWorker.Common;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using Microsoft.Extensions.DependencyInjection;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.Games
{
    [BotPlugin(
        Id = "game.2048",
        Name = "2048游戏",
        Version = "1.0.0",
        Author = "Matrix",
        Description = "经典的2048数字合并游戏",
        Category = "Games"
    )]
    public class Game2048Plugin : IPlugin
    {
        private readonly IGroupRepository _groupRepo;

        public Game2048Plugin(IGroupRepository groupRepo)
        {
            _groupRepo = groupRepo;
        }

        public async Task InitAsync(IRobot robot)
        {
            await robot.RegisterSkillAsync(new SkillCapability
            {
                Name = "2048游戏",
                Commands = ["2048", "w", "a", "s", "d", "上", "下", "左", "右", "开始", "结束"],
                Description = "发送【2048】进入游戏，发送【wsad】或【上下左右】控制"
            }, HandleGameAsync);
        }

        public Task StopAsync() => Task.CompletedTask;

        private async Task<string> HandleGameAsync(IPluginContext ctx, string[] args)
        {
            var userId = long.Parse(ctx.UserId);
            var groupId = long.Parse(ctx.GroupId ?? "0");
            var cmdPara = ctx.RawMessage.Trim();

            if (cmdPara.Equals("2048", StringComparison.OrdinalIgnoreCase))
            {
                cmdPara = ""; // 触发进入游戏
            }

            return await Game2048.GetGameResAsync(_groupRepo, groupId, userId, cmdPara);
        }
    }

    /// <summary>
    /// 2048游戏
    /// </summary>
    internal class Game2048
    {
        public static Dictionary<long, bool> dict = new();

        public static async Task<string> GetGameResAsync(IGroupRepository repository, long groupId, long qq, string cmdPara)
        {
            if (string.IsNullOrEmpty(cmdPara))
            {
                int i = await UserInfo.SetStateAsync(UserInfo.States.G2048, qq);
                return i == -1
                    ? "系统繁忙，请稍后再试"
                    : "发【开始】，发送【上下左右】或【wsad】控制游戏";
            }
            else if (cmdPara == "结束")
            {
                int i = await UserInfo.SetStateAsync(UserInfo.States.Chat, qq);
                return i == -1
                    ? "系统繁忙，请稍后再试"
                    : "2048游戏结束";
            }

            int[,] tiles = await GetTilesAsync(repository, groupId);
            if (new[] { "上", "w", "8" }.Contains(cmdPara))
                TurnTo(tiles, Direct.Left);

            if (new[] { "下", "s", "2" }.Contains(cmdPara))
                TurnTo(tiles, Direct.Right);

            if (new[] { "左", "a", "4" }.Contains(cmdPara))
            {
                if (dict.ContainsKey(groupId))
                    dict[groupId] = true;
                else
                    dict.Add(groupId, true);

                TurnTo(tiles, Direct.Up);
            }

            if (new[] { "右", "d", "6" }.Contains(cmdPara))
            {
                if (dict.ContainsKey(groupId))
                    dict[groupId] = false;
                else
                    dict.Add(groupId, false);
                TurnTo(tiles, Direct.Down);
            }

            if (cmdPara == "开始")
            {
                InitTiles(tiles);
                RandomValue(tiles, 2);
            }                                         
            
            string res = PrintTiles(groupId, tiles);
            if (IsGameOver(tiles))
                res += "Game Over!";
            await SaveTilesAsync(repository, groupId, tiles);
            return res;
        }

        public enum Direct
        {
            Up,
            Left,
            Down,
            Right,
            Other
        }

        public static int TurnTo(int[,] tiles, Direct direct)
        {
            int res = Slide(tiles, (int)direct, out _, out _);
            Console.WriteLine("direct:" + direct.ToString());

            if (res > 0)
            {
                RandomValue(tiles);
            }
            
            return res;
        }

        public static void InitTiles(int[,] tiles)
        {
            for (int i = 0; i < 4; i++)
            {
                for (int j = 0; j < 4; j++)
                {
                    tiles[i, j] = 0;
                }
            }            
        }

        public static async Task SaveTilesAsync(IGroupRepository repository, long groupId, int[,] tiles)
        {
            string res = "";
            for (int i = 0; i < 4; i++)
            {
                for (int j = 0; j < 4; j++)
                {
                    res += $" {tiles[i, j]}";
                }
            }
            await repository.SetValueAsync("game_2048", res.Trim(), groupId);
        }

        public static async Task<int[,]> GetTilesAsync(IGroupRepository repository, long groupId)
        {
            int[,] tiles = new int[4, 4];
            string res = await repository.GetValueAsync("game_2048", groupId);
            if (string.IsNullOrEmpty(res))
            {
                InitTiles(tiles);
            }
            else
            {
                var items = res.Split(" ", StringSplitOptions.RemoveEmptyEntries);
                int k = 0;
                for (int i = 0; i <= 3; i++)
                {
                    for (int j = 0; j <= 3; j++)
                    {
                        if (k < items.Length)
                        {
                            int.TryParse(items[k], out tiles[i, j]);
                        }
                        k++;
                    }
                }
            }
            return tiles;
        }

        // 最大值
        public static int GetMax(int[,] tiles, out int x, out int y)
        {
            int max = 0;
            x = 0;
            y = 0;
            for (int i = 0; i < 4; i++)
            {
                for (int j = 0; j < 4; j++)
                {
                    if (tiles[i, j] > max)
                    {
                        max = tiles[i, j];
                        x = i;
                        y = j;
                    }
                }
            }
            return max;
        }

        public static string PrintTiles(long groupId, int[,] tiles)
        {
            int max = GetMax(tiles, out _, out _);
            string res = string.Empty;
            int k = 1;
            for (int i = 0; i <= 3; i++)
            {
                for (int j = 0; j <= 3; j++)
                {
                    int value = tiles[i, j];
                    if (value == 0)
                        res += $" ".Times(max.AsString().Length + 1);
                    else
                    {
                        bool leftAlign = true;
                        if (dict.TryGetValue(groupId, out bool val))
                        {
                            leftAlign = val;
                        }

                        res += leftAlign
                            ? $"{value}{" ".Times(max.AsString().Length - value.AsString().Length + 1)}"
                            : $"{" ".Times(max.AsString().Length - value.AsString().Length + 1)}{value}";
                    }
                    Console.Write(res);
                    if (k % 4 == 0)
                    {
                        res += "\n";
                        Console.WriteLine();
                    }
                    k++;
                }
            }
            return res;
        }

        public static int RandomValue(int[,] tiles, int count = 1)
        {
            //0的数量
            int zeroCount = ZeroCount(tiles);
            if (zeroCount == 0) return 0;

            //随机位置赋值2或4
            int posRandom = Random.Shared.Next(1, zeroCount + 1);
            int k = 0;
            for (int i = 0; i <= 3; i++)
            {
                for (int j = 0; j <= 3; j++)
                {
                    if (tiles[i, j] == 0)
                    {
                        k++;
                        if (k == posRandom)
                        {
                            tiles[i, j] = Random.Shared.Next(1, 10) == 4 ? 4 : 2;
                            break;
                        }
                    }
                }
                if (k == posRandom) break;
            }
            if (count == 2)
            {
                RandomValue(tiles);
            }
            return k;
        }

        public static int ZeroCount(int[,] tiles)
        {
            int res = 0;
            for (int i = 0; i <= 3; i++)
            {
                for (int j = 0; j <= 3; j++)
                {
                    if (tiles[i, j] == 0)
                        res++;
                }
            }
            return res;
        }

        public static int Slide(int[,] tiles, int direct, out int slide, out int merge)
        {
            if (direct > 0)
            {
                for (int i = 0; i < 4 - direct; i++)
                    Rotating(tiles);
            }

            slide = 0;
            merge = 0;
            for (int i = 0; i <= 3; i++)
            {
                for (int j = 0; j <= 2; j++)
                {
                    if (tiles[i, j] == 0 && tiles[i, j + 1] != 0)
                    {
                        tiles[i, j] = tiles[i, j + 1];
                        tiles[i, j + 1] = 0;
                        slide++;
                    }
                    if (tiles[i, j] != 0 && tiles[i, j] == tiles[i, j + 1])
                    {
                        tiles[i, j] *= 2;
                        tiles[i, j + 1] = 0;
                        merge++;
                    }
                }
            }
            int k = slide + merge;
            int res = k;
            while (k > 0)
            {
                k = Slide(tiles, 0, out int newSlide, out int newMerge);
                res += k;
                slide += newSlide;
                merge += newMerge;
            }

            if (direct > 0)
            {
                for (int i = 0; i < direct; i++)
                    Rotating(tiles);
            }
            return res;
        }

        // Game over 
        public static bool IsGameOver(int[,] tiles)
        {
            if (tiles == null)
                return false;

            foreach (int tile in tiles)
            {                
                if (tile == 0)
                    return false;
            }
            int k = 0;
            for (int i = 0; i <= 3; i++)
            {
                for (int j = 0; j <= 3; j++)
                {
                    if (IsHaveSame(tiles, i, j))
                        return false;
                    k++;
                }
            }
            return true;
        }

        // i,j 点周围是否有相同数值的格子
        public static bool IsHaveSame(int[,] tiles, int i, int j)
        {
            return i - 1 >= 0 && tiles[i, j] == tiles[i - 1, j] ||
                   i + 1 <= 3 && tiles[i, j] == tiles[i + 1, j] ||
                   j - 1 >= 0 && tiles[i, j] == tiles[i, j - 1] ||
                   j + 1 <= 3 && tiles[i, j] == tiles[i, j + 1];
        }

        /// <summary>
        /// 右转90度
        /// </summary>
        public static void Rotating(int[,] tiles)
        {
            int n = tiles.GetLength(0);
            for (int i = 0; i < n / 2; i++)
            {
                for (int j = i; j < n - i - 1; j++)
                {
                    int top =  tiles[i, j]; 

                    //向左移动到顶部
                    tiles[i, j] = tiles[n - 1 - j, i];

                    //从底部向左移动
                    tiles[n - 1 - j, i] = tiles[n - i - 1, n - 1 - j];

                    //向右移动到底部
                    tiles[n - i - 1, n - 1 - j] = tiles[j, n - i - 1];

                    //从上往右移动
                    tiles[j, n - i - 1] = top;
                }
            }
        }
        /// <summary>
        /// 右转90度
        /// </summary>
        public static void RotatingIt(int[,] tiles)
        {
            //for (int i = 0; i <= 2; i++)
            //{
                Rotating(tiles);
            //}
        }

    }
}
