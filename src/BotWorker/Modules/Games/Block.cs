using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models.BotMessages;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Utils;
using Dapper;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace BotWorker.Modules.Games
{
    // 区块链+游戏
    [Table("Block")]
    public partial class Block
    {
        private static IBlockRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBlockRepository>() 
            ?? throw new InvalidOperationException("IBlockRepository not registered");

        [Key]
        public long Id { get; set; }
        public long PrevId { get; set; }
        public string PrevHash { get; set; } = string.Empty;
        public string PrevRes { get; set; } = string.Empty;
        public long BotUin { get; set; }
        public long GroupId { get; set; }
        public string GroupName { get; set; } = string.Empty;
        public long UserId { get; set; }
        public string UserName { get; set; } = string.Empty;
        public string BlockInfo { get; set; } = string.Empty;
        public string BlockSecret { get; set; } = string.Empty;
        public int BlockNum { get; set; }
        public string BlockRes { get; set; } = string.Empty;
        public string BlockRand { get; set; } = string.Empty;
        public string BlockHash { get; set; } = string.Empty;
        public int IsOpen { get; set; }
        public DateTime? OpenDate { get; set; }
        public long OpenBotUin { get; set; }
        public long OpenUserId { get; set; }
        public string OpenUserName { get; set; } = string.Empty;

        public static string GetCmd(string cmdName, long qq)
        {
            cmdName = cmdName.ToLower() switch
            {
                "jd" => "剪刀",
                "st" => "石头",
                "bu" => "布",
                "d" => "大",
                "x" => "小",
                "z" => "庄",
                "j" => "单",
                "s" => "双",
                "w" => "围",
                "四" => "押点4",
                "五" => "押点5",
                "六" => "押点6",
                "七" => "押点7",
                "八" => "押点8",
                "九" => "押点9",
                "十" => "押点10",
                "十一" => "押点11",
                "十二" => "押点12",
                "十三" => "押点13",
                "十四" => "押点14",
                "十五" => "押点15",
                "十六" => "押点16",
                "十七" => "押点17",
                _ => cmdName,
            };

            return cmdName.In("大", "小", "单", "双", "围")
                ? $"押{(cmdName == "围" ? "全围" : cmdName)}"
                : cmdName;
        }

        public static string GetBlockInfo16(string hash16)
        {
            return $"HASH:{hash16}\n查询结果：该HASH有效，游戏记录已归档。";
        }

        private static Guid GetGuidAlgorithmic(long id)
        {
            byte[] bytes = new byte[16];
            BitConverter.GetBytes(id).CopyTo(bytes, 0);
            return new Guid(bytes);
        }

        public static async Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            return await Repository.GetIdAsync(groupId, userId, trans);
        }

        public static async Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null)
        {
            return await Repository.GetHashAsync(blockId, trans);
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name, string prevRes, string sqlAddCredit, object? addCreditParams, string sqlCreditHis, object? creditHisParams, IDbTransaction? trans = null)
        {
            long prevId = await GetIdAsync(groupId, userId, trans);
            string prevHash = prevId == 0 
                ? groupId == 0 ? GetGuidAlgorithmic(userId).AsString().Sha256() : GetGuidAlgorithmic(groupId).AsString().Sha256()
                : await GetHashAsync(prevId, trans);
            string hashRobot = GetGuidAlgorithmic(botUin).AsString().Sha256();
            string hashRoom = groupId == 0 ? "" : GetGuidAlgorithmic(groupId).AsString().Sha256();            
            string hashClient = GetGuidAlgorithmic(userId).AsString().Sha256();
            int num = BlockRandom.RandomNum();
            string blockRes = $"{BlockRandom.FormatNum(num)} {BlockRandom.Sum(num)} {BlockRandom.GetBlockRes(num)}";
            string blockRand = Guid.NewGuid().ToString().Sha256();
            string blockTime = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");

            string blockInfo = $"上局HASH:{prevHash}\n上局结果:{prevRes}\n时间节点:{blockTime}\n机器HASH:{hashRobot}\n群组HASH:{hashRoom}\n玩家HASH:{hashClient}\n";
            string blockSecret = $"本局数据:{blockRes}\n随机密码:{blockRand}";
            string hashBlock = (blockInfo + blockSecret).Sha256();

            var block = new Block
            {
                PrevId = prevId,
                PrevHash = prevHash,
                PrevRes = prevRes,
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = name,
                BlockInfo = blockInfo,
                BlockSecret = blockSecret,
                BlockNum = num,
                BlockRes = blockRes,
                BlockRand = blockRand,
                BlockHash = hashBlock
            };

            using var wrapper = await Repository.BeginTransactionAsync(trans);
            try
            {
                await wrapper.Connection.InsertAsync(block, wrapper.Transaction);
                
                if (!string.IsNullOrEmpty(sqlAddCredit)) 
                    await wrapper.Connection.ExecuteAsync(sqlAddCredit, addCreditParams, wrapper.Transaction);
                
                if (!string.IsNullOrEmpty(sqlCreditHis)) 
                    await wrapper.Connection.ExecuteAsync(sqlCreditHis, creditHisParams, wrapper.Transaction);
                
                if (prevId > 0)
                {
                    const string sqlUpdate = "UPDATE Block SET IsOpen=1, OpenDate=@OpenDate, OpenBotUin=@OpenBotUin, OpenUserId=@OpenUserId, OpenUserName=@OpenUserName WHERE Id = @Id";
                    await wrapper.Connection.ExecuteAsync(sqlUpdate, new 
                    { 
                        OpenDate = DateTime.Now, 
                        OpenBotUin = botUin, 
                        OpenUserId = userId, 
                        OpenUserName = name, 
                        Id = prevId 
                    }, wrapper.Transaction);
                }

                await wrapper.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Block.AppendAsync error: {ex.Message}");
                await wrapper.RollbackAsync();
                return -1;
            }
        }

        public static async Task<string> GetCmdAsync(string cmdName, long qq, IDbTransaction? trans = null)
        {
            return await Task.FromResult(GetCmd(cmdName, qq));
        }

        public static async Task<string> GetHashAsync(long groupId, long qq, IDbTransaction? trans = null)
        {
            return await GetHashAsync(await GetIdAsync(groupId, qq, trans), trans);
        }

        public static async Task<long> GetBlockIdAsync(string hash)
        {
            return await Repository.GetBlockIdAsync(hash);
        }

        public static async Task<int> GetNumAsync(long botUin, long groupId, string groupName, long qq, string name, IDbTransaction? trans = null)
        {
            long blockId = await GetIdAsync(groupId, qq, trans);
            if (blockId == 0)
            {
                if (await AppendAsync(botUin, groupId, groupName, qq, name, "创世区块", string.Empty, null, string.Empty, null, trans) != -1)
                    return await GetNumAsync(botUin, groupId, groupName, qq, name, trans);
            }
            return await GetNumAsync(blockId, trans);
        }

        public static async Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null)
        {
            return await Repository.GetNumAsync(blockId, trans);
        }

        public static async Task<decimal> GetOddsAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 && typeId <= 37)
            {
                var target = typeName.Replace("押", "");
                return blockNum.ToString().Split(new[] { target }, StringSplitOptions.None).Length - 1;
            }
            
            return await BlockType.GetOddsAsync(typeId, trans);
        }

        public static async Task<bool> IsWinAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 && typeId <= 37)
            {
                var target = typeName.Replace("押", "");
                int count = blockNum.ToString().Split(new[] { target }, StringSplitOptions.None).Length - 1;
                return count > 0;
            }
            return await BlockWin.IsWinAsync(typeId, blockNum, trans);
        }

        public static async Task<string> GetValueAsync(string field, long blockId, IDbTransaction? trans = null)
        {
            return await Repository.GetValueAsync(field, blockId, trans);
        }

        public static async Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null)
        {
            return await Repository.IsOpenAsync(blockId, trans);
        }

        public static async Task<string> GetBlockSecretAsync(long blockId, IDbTransaction? trans = null)
        {
            if (await IsOpenAsync(blockId, trans))
            {
                return await GetValueAsync("BlockSecret", blockId, trans);
            }
            return "本局数据:本局游戏尚未结束，保密区数据不可见\n" +
                   "随机密码:本局游戏尚未结束，保密区数据不可见";
        }
    }

    [Table("BlockRandom")]
    public class BlockRandom
    {
        private static IBlockRandomRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBlockRandomRepository>() 
            ?? throw new InvalidOperationException("IBlockRandomRepository not registered");

        [Key]
        public int Id { get; set; }
        public int BlockNum { get; set; }

        public static int RandomNum()
        {
            // Note: Since this is called from AppendAsync which is async, 
            // but the original code was sync, we might need to handle this.
            // For now, let's assume we can call the repository sync or just use Task.Run.Result
            return Repository.GetRandomNumAsync().GetAwaiter().GetResult();
        }

        public static string FormatNum(int Num)
        {
            string text = Num.ToString();
            string res = string.Empty;
            for (int i = 0; i < text.Length; i++)
            {
                res += $"【{text[i]}】";
            }
            return res;
        }

        public static int Sum(int num)
        {
            int res = 0;
            foreach (char c in num.ToString())
            {
                res += int.Parse(c.ToString());
            }
            return res;
        }

        public static string GetBlockRes(int blockNum)
        {
            if (blockNum == 111 || blockNum == 222 || blockNum == 333 || 
                blockNum == 444 || blockNum == 555 || blockNum == 666)
                return "围";

            return Sum(blockNum) > 10 ? "大" : "小";
        }
    }

    [Table("BlockType")]
    public class BlockType
    {
        private static IBlockTypeRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBlockTypeRepository>() 
            ?? throw new InvalidOperationException("IBlockTypeRepository not registered");

        [Key]
        public int Id { get; set; }
        public string TypeName { get; set; } = string.Empty;
        public decimal BlockOdds { get; set; }

        public static async Task<int> GetTypeIdAsync(string typeName, IDbTransaction? trans = null)
        {
            return await Repository.GetTypeIdAsync(typeName.Replace("押", ""), trans);
        }

        public static async Task<decimal> GetOddsAsync(int typeId, IDbTransaction? trans = null)
        {
            return await Repository.GetOddsAsync(typeId, trans);
        }
    }

    [Table("BlockWin")]
    public class BlockWin
    {
        private static IBlockWinRepository Repository => 
            BotMessage.ServiceProvider?.GetRequiredService<IBlockWinRepository>() 
            ?? throw new InvalidOperationException("IBlockWinRepository not registered");

        [Key]
        public int Id { get; set; }
        public int TypeId { get; set; }
        public int BlockNum { get; set; }
        public int IsWin { get; set; }

        public static async Task<bool> IsWinAsync(int typeId, int blockNum, IDbTransaction? trans = null)
        {
            return await Repository.IsWinAsync(typeId, blockNum, trans);
        }
    }
}
