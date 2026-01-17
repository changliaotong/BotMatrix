using System;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common.Extensions;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Interfaces;
using BotWorker.Domain.Repositories;
using Dapper;
using Dapper.Contrib.Extensions;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.Games
{
    public class BlockService : IBlockService
    {
        private readonly IBlockRepository _blockRepo;
        private readonly IBlockRandomRepository _randomRepo;
        private readonly IBlockTypeRepository _typeRepo;
        private readonly IBlockWinRepository _winRepo;
        private readonly ILogger<BlockService> _logger;

        public BlockService(
            IBlockRepository blockRepo,
            IBlockRandomRepository randomRepo,
            IBlockTypeRepository typeRepo,
            IBlockWinRepository winRepo,
            ILogger<BlockService> logger)
        {
            _blockRepo = blockRepo;
            _randomRepo = randomRepo;
            _typeRepo = typeRepo;
            _winRepo = winRepo;
            _logger = logger;
        }

        public async Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            return await _blockRepo.GetIdAsync(groupId, userId, trans);
        }

        public async Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null)
        {
            return await _blockRepo.GetHashAsync(blockId, trans);
        }

        public (string sql, object paras) SqlAppend(long botUin, long groupId, string groupName, long userId, string name, string prevRes, string blockRes, string blockRand, string blockInfo, string blockHash, long prevId)
        {
            const string sql = @"INSERT INTO Block (PrevId, PrevHash, PrevRes, BotUin, GroupId, GroupName, UserId, UserName, BlockInfo, BlockSecret, BlockNum, BlockRes, BlockRand, BlockHash) 
                               VALUES (@PrevId, @PrevHash, @PrevRes, @BotUin, @GroupId, @GroupName, @UserId, @UserName, @BlockInfo, @BlockSecret, @BlockNum, @BlockRes, @BlockRand, @BlockHash)";
            
            var paras = new
            {
                PrevId = prevId,
                PrevHash = "", // This will be set by the caller if needed, or we can calculate it here
                PrevRes = prevRes,
                BotUin = botUin,
                GroupId = groupId,
                GroupName = groupName,
                UserId = userId,
                UserName = name,
                BlockInfo = blockInfo,
                BlockSecret = $"本局数据:{blockRes}\n随机密码:{blockRand}",
                BlockNum = 0, // Should be passed in
                BlockRes = blockRes,
                BlockRand = blockRand,
                BlockHash = blockHash
            };

            return (sql, paras);
        }

        public (string sql, object paras) SqlUpdateOpen(long botUin, long userId, string name, long prevId)
        {
            const string sql = "UPDATE Block SET IsOpen=1, OpenDate=@OpenDate, OpenBotUin=@OpenBotUin, OpenUserId=@OpenUserId, OpenUserName=@OpenUserName WHERE Id = @Id";
            var paras = new
            {
                OpenDate = DateTime.Now,
                OpenBotUin = botUin,
                OpenUserId = userId,
                OpenUserName = name,
                Id = prevId
            };
            return (sql, paras);
        }

        public async Task<long> GetBlockIdAsync(string hash)
        {
            return await _blockRepo.GetBlockIdAsync(hash);
        }

        public async Task<int> GetNumAsync(long botUin, long groupId, string groupName, long qq, string name, IDbTransaction? trans = null)
        {
            long blockId = await GetIdAsync(groupId, qq, trans);
            if (blockId == 0)
            {
                // Create genesis block
                int num = await _randomRepo.GetRandomNumAsync();
                string blockRes = $"{FormatNum(num)} {Sum(num)} {GetBlockRes(num)}";
                string blockRand = Guid.NewGuid().ToString().Sha256();
                string blockTime = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss");
                string prevHash = groupId == 0 ? GetGuidAlgorithmic(qq).AsString().Sha256() : GetGuidAlgorithmic(groupId).AsString().Sha256();
                
                string hashRobot = GetGuidAlgorithmic(botUin).AsString().Sha256();
                string hashRoom = groupId == 0 ? "" : GetGuidAlgorithmic(groupId).AsString().Sha256();            
                string hashClient = GetGuidAlgorithmic(qq).AsString().Sha256();

                string blockInfo = $"上局HASH:{prevHash}\n上局结果:创世区块\n时间节点:{blockTime}\n机器HASH:{hashRobot}\n群组HASH:{hashRoom}\n玩家HASH:{hashClient}\n";
                string blockSecret = $"本局数据:{blockRes}\n随机密码:{blockRand}";
                string hashBlock = (blockInfo + blockSecret).Sha256();

                var block = new Block
                {
                    PrevId = 0,
                    PrevHash = prevHash,
                    PrevRes = "创世区块",
                    BotUin = botUin,
                    GroupId = groupId,
                    GroupName = groupName,
                    UserId = qq,
                    UserName = name,
                    BlockInfo = blockInfo,
                    BlockSecret = blockSecret,
                    BlockNum = num,
                    BlockRes = blockRes,
                    BlockRand = blockRand,
                    BlockHash = hashBlock
                };

                using var wrapper = await _blockRepo.BeginTransactionAsync(trans);
                try
                {
                    await wrapper.Connection.InsertAsync(block, wrapper.Transaction);
                    await wrapper.CommitAsync();
                    return await GetNumAsync(botUin, groupId, groupName, qq, name, trans);
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "Block.GetNumAsync (Genesis) error: {Message}", ex.Message);
                    await wrapper.RollbackAsync();
                    return 0;
                }
            }
            return await GetNumAsync(blockId, trans);
        }

        public async Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null)
        {
            return await _blockRepo.GetNumAsync(blockId, trans);
        }

        public async Task<decimal> GetOddsAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 && typeId <= 37)
            {
                var target = typeName.Replace("押", "");
                return blockNum.ToString().Split(new[] { target }, StringSplitOptions.None).Length - 1;
            }
            
            return await _typeRepo.GetOddsAsync(typeId, trans);
        }

        public async Task<bool> IsWinAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 && typeId <= 37)
            {
                var target = typeName.Replace("押", "");
                int count = blockNum.ToString().Split(new[] { target }, StringSplitOptions.None).Length - 1;
                return count > 0;
            }
            return await _winRepo.IsWinAsync(typeId, blockNum, trans);
        }

        public async Task<string> GetValueAsync(string field, long blockId, IDbTransaction? trans = null)
        {
            return await _blockRepo.GetValueAsync(field, blockId, trans);
        }

        public async Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null)
        {
            return await _blockRepo.IsOpenAsync(blockId, trans);
        }

        public async Task<string> GetBlockSecretAsync(long blockId, IDbTransaction? trans = null)
        {
            if (await IsOpenAsync(blockId, trans))
            {
                return await GetValueAsync("BlockSecret", blockId, trans);
            }
            return "本局数据:本局游戏尚未结束，保密区数据不可见\n" +
                   "随机密码:本局游戏尚未结束，保密区数据不可见";
        }

        public async Task<string> GetCmdAsync(string cmdName, long qq, IDbTransaction? trans = null)
        {
            return await Task.FromResult(GetCmd(cmdName, qq));
        }

        public string GetCmd(string cmdName, long qq)
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

        public string GetBlockInfo16(string hash16)
        {
            return $"HASH:{hash16}\n查询结果：该HASH有效，游戏记录已归档。";
        }

        private Guid GetGuidAlgorithmic(long id)
        {
            byte[] bytes = new byte[16];
            BitConverter.GetBytes(id).CopyTo(bytes, 0);
            return new Guid(bytes);
        }

        public string FormatNum(int Num)
        {
            string text = Num.ToString();
            string res = string.Empty;
            for (int i = 0; i < text.Length; i++)
            {
                res += $"【{text[i]}】";
            }
            return res;
        }

        public int Sum(int num)
        {
            int res = 0;
            foreach (char c in num.ToString())
            {
                res += int.Parse(c.ToString());
            }
            return res;
        }

        public string GetBlockRes(int blockNum)
        {
            if (blockNum == 111 || blockNum == 222 || blockNum == 333 || 
                blockNum == 444 || blockNum == 555 || blockNum == 666)
                return "围";

            return Sum(blockNum) > 10 ? "大" : "小";
        }
    }
}
