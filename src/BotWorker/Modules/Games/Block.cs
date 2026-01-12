using System;
using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Modules.Games
{
    // 区块链+游戏
    public partial class Block : MetaDataGuid<Block>
    {
        public override string TableName => "Block";
        public override string KeyField => "Id";

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

        public static async Task<((string sql, IDataParameter[] paras) sqlInsert, (string sql, IDataParameter[] paras) sqlUpdatePrev)> SqlAppendAsync(long botUin, long groupId, string groupName, long userId, string name, string prevRes, IDbTransaction? trans = null)
        {
            long prevId = await GetIdAsync(groupId, userId, trans);
            string prevHash = prevId == 0
                ? groupId == 0 ? GetGuidAlgorithmic(userId).AsString().Sha256() : GetGuidAlgorithmic(groupId).AsString().Sha256()
                : await GetHashAsync(prevId, trans);
            string hashRobot = GetGuidAlgorithmic(botUin).AsString().Sha256();
            string hashRoom = groupId == 0 ? "" : GetGuidAlgorithmic(groupId).AsString().Sha256();
            string hashClient = GetGuidAlgorithmic(userId).AsString().Sha256();
            int num = BlockRandom.RandomNum();
            string blockRes = $"{FormatNum(num)} {Sum(num)} {GetBlockRes(num)}";
            string blockRand = GetNewId().Sha256();
            string blockTime = GetTimeStamp();

            string blockInfo = $"上局HASH:{prevHash}\n上局结果:{prevRes}\n时间节点:{blockTime}\n机器HASH:{hashRobot}\n群组HASH:{hashRoom}\n玩家HASH:{hashClient}\n";
            string blockSecret = $"本局数据:{blockRes}\n随机密码:{blockRand}";
            string hashBlock = (blockInfo + blockSecret).Sha256();

            var sql1 = SqlInsert(new List<Cov>
            {
                new Cov("PrevId", prevId),
                new Cov("PrevHash", prevHash),
                new Cov("PrevRes", prevRes),
                new Cov("BotUin", botUin),
                new Cov("GroupId", groupId),
                new Cov("GroupName", groupName),
                new Cov("UserId", userId),
                new Cov("UserName", name),
                new Cov("BlockInfo", blockInfo),
                new Cov("BlockSecret", blockSecret),
                new Cov("BlockNum", num),
                new Cov("BlockRes", blockRes),
                new Cov("BlockRand", blockRand),
                new Cov("BlockHash", hashBlock)
            });

            var sql2 = SqlSetValues($"IsOpen=1, OpenDate={SqlDateTime}, OpenBotUin={botUin}, OpenUserId={userId}, OpenUserName={name.Quotes()}", prevId);
            return (sql1, sql2);
        }

        public static async Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null)
        {
            string sql = $"SELECT {SqlIsNull("MAX(Id)", "0")} AS res FROM {FullName} WHERE GroupId = {groupId} AND IsOpen = 0";
            var res = await ExecScalarAsync<long>(groupId == 0 ? $"{sql} AND UserId = {userId} " : sql, trans);
            return res;
        }

        public static async Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null)
        {
            if (blockId == 0) return string.Empty;
            return await QueryScalarAsync<string>($"SELECT BlockHash FROM {FullName} WHERE Id = {blockId}", trans) ?? string.Empty;
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name, string prevRes, (string sql, IDataParameter[] paras) sqlAddCredit, (string sql, IDataParameter[] paras) sqlCreditHis, IDbTransaction? trans = null)
        {
            long prevId = await GetIdAsync(groupId, userId, trans);
            string prevHash = prevId == 0 
                ? groupId == 0 ? GetGuidAlgorithmic(userId).AsString().Sha256() : GetGuidAlgorithmic(groupId).AsString().Sha256()
                : await GetHashAsync(prevId, trans);
            string hashRobot = GetGuidAlgorithmic(botUin).AsString().Sha256();
            string hashRoom = groupId == 0 ? "" : GetGuidAlgorithmic(groupId).AsString().Sha256();            
            string hashClient = GetGuidAlgorithmic(userId).AsString().Sha256();
            int num = BlockRandom.RandomNum();
            string blockRes = $"{FormatNum(num)} {Sum(num)} {GetBlockRes(num)}";
            string blockRand = GetNewId().Sha256();
            string blockTime = GetTimeStamp();

            string blockInfo = $"上局HASH:{prevHash}\n上局结果:{prevRes}\n时间节点:{blockTime}\n机器HASH:{hashRobot}\n群组HASH:{hashRoom}\n玩家HASH:{hashClient}\n";
            string blockSecret = $"本局数据:{blockRes}\n随机密码:{blockRand}";
            string hashBlock = (blockInfo + blockSecret).Sha256();

            var sql = SqlInsert([
                                    new Cov("PrevId", prevId),
                                    new Cov("PrevHash", prevHash),
                                    new Cov("PrevRes", prevRes),
                                    new Cov("BotUin", botUin),
                                    new Cov("GroupId", groupId),
                                    new Cov("GroupName", groupName),
                                    new Cov("UserId", userId),
                                    new Cov("UserName", name),
                                    new Cov("BlockInfo", blockInfo),
                                    new Cov("BlockSecret", blockSecret),
                                    new Cov("BlockNum", num),
                                    new Cov("BlockRes", blockRes),
                                    new Cov("BlockRand", blockRand),
                                    new Cov("BlockHash", hashBlock)
                                ]);

            var sql2 = SqlSetValues($"IsOpen=1, OpenDate={SqlDateTime}, OpenBotUin={botUin}, OpenUserId={userId}, OpenUserName={name.Quotes()}", prevId);
            
            using var wrapper = await BeginTransactionAsync(trans);
            try
            {
                await ExecAsync(sql.sql, wrapper.Transaction, sql.paras);
                if (!sqlAddCredit.sql.IsNull()) await ExecAsync(sqlAddCredit.sql, wrapper.Transaction, sqlAddCredit.paras);
                if (!sqlCreditHis.sql.IsNull()) await ExecAsync(sqlCreditHis.sql, wrapper.Transaction, sqlCreditHis.paras);
                await ExecAsync(sql2.sql, wrapper.Transaction, sql2.parameters);
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
            return (await GetWhereAsync("Id", $"BlockHash = {hash.Quotes()}")).AsLong();
        }

        public static async Task<int> GetNumAsync(long botUin, long groupId, string groupName, long qq, string name, IDbTransaction? trans = null)
        {
            long blockId = await GetIdAsync(groupId, qq, trans);
            if (blockId == 0)
            {
                if (await AppendAsync(botUin, groupId, groupName, qq, name, "创世区块", (string.Empty, Array.Empty<IDataParameter>()), (string.Empty, Array.Empty<IDataParameter>()), trans) != -1)
                    return await GetNumAsync(botUin, groupId, groupName, qq, name, trans);
            }
            return await GetNumAsync(blockId, trans);
        }

        public static async Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null)
        {
            return await GetIntAsync("BlockNum", blockId, null, trans);
        }

        public static async Task<decimal> GetOddsAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 & typeId <= 37)
                return blockNum.ToString().Split([typeName.Replace("押", "")], StringSplitOptions.None).Length - 1;

            // 算法实现赔率
            return typeId switch
            {
                1 or 2 or 3 or 4 => 1.0m, // 大小单双
                5 => 24.0m, // 全围
                6 or 19 => 50.0m, // 点4, 点17
                7 or 18 => 18.0m, // 点5, 点16
                8 or 17 => 14.0m, // 点6, 点15
                9 or 16 => 12.0m, // 点7, 点14
                10 or 15 => 8.0m, // 点8, 点13
                11 or 14 => 6.0m, // 点9, 点12
                12 or 13 => 6.0m, // 点10, 点11
                _ => 1.0m
            };
        }

        public static string FormatNum(int Num)
        {
            string text = Num.ToString().PadLeft(3, '0');
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
            foreach (char c in num.ToString().PadLeft(3, '0'))
            {
                res += int.Parse(c.ToString());
            }
            return res;
        }

        public static string GetBlockRes(int blockNum)
        {
            if (blockNum.In(111, 222, 333, 444, 555, 666))
                return "围";

            return Sum(blockNum) > 10 ? "大" : "小";
        }       

        public static async Task<string> GetBlockInfo16Async(string hash16, IDbTransaction? trans = null)
        {
            long block_id = (await GetWhereAsync(SqlIsNull("Id", "0"), $"BlockHash LIKE {hash16.QuotesLike()}", "", trans)).AsLong();
            return block_id == 0
                ? ""
                : await GetBlockInfoAsync(block_id, trans) + await GetBlockSecretAsync(block_id, trans);
        }

        public static async Task<string> GetBlockInfoAsync(long blockId, IDbTransaction? trans = null)
        {
            return await GetValueAsync("BlockInfo", blockId, null, trans);
        }

        public static async Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null)
        {
            return await GetBoolAsync("IsOpen", blockId, null, trans);
        }

        public static async Task<string> GetBlockSecretAsync(long blockId, IDbTransaction? trans = null)
        {
            return await IsOpenAsync(blockId, trans)
                    ? await GetValueAsync("BlockSecret", blockId, null, trans)
                    : $"本局数据:本局游戏尚未结束，保密区数据不可见\n" +
                      $"随机密码:本局游戏尚未结束，保密区数据不可见";
        }

        public static async Task<bool> IsWinAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null)
        {
            if (typeId >= 32 & typeId <= 37)
            {
                int i = blockNum.ToString().Split([typeName.Replace("押", "")], StringSplitOptions.None).Length - 1;
                return i > 0;
            }

            int sum = Sum(blockNum);
            bool isWei = blockNum.In(111, 222, 333, 444, 555, 666);

            return typeId switch
            {
                1 => sum > 10 && !isWei, // 大
                2 => sum <= 10 && !isWei, // 小
                3 => sum % 2 != 0 && !isWei, // 单
                4 => sum % 2 == 0 && !isWei, // 双
                5 => isWei, // 围
                _ => false
            };
        }
    }

    public class BlockRandom : MetaData<BlockRandom>
    {
        public override string TableName => "BlockRandom";
        public override string KeyField => "Id";

        public static int RandomNum()
        {
            int[] dice = [RandomInt(1, 6), RandomInt(1, 6), RandomInt(1, 6)];
            Array.Sort(dice);
            return (dice[0] * 100) + (dice[1] * 10) + dice[2];
        }

        public static async Task<int> RandomNumAsync(IDbTransaction? trans = null)
        {
            return await Task.FromResult(RandomNum());
        }
    }
    public class BlockType : MetaData<BlockType>
    {
        public override string TableName => "Blocktype";
        public override string KeyField => "Id";

        public static async Task<int> GetTypeIdAsync(string TypeName, IDbTransaction? trans = null)
        {
            string name = TypeName.Replace("押", "");
            int id = name switch
            {
                "大" => 1,
                "小" => 2,
                "单" => 3,
                "双" => 4,
                "全围" => 5,
                "点4" => 6,
                "点5" => 7,
                "点6" => 8,
                "点7" => 9,
                "点8" => 10,
                "点9" => 11,
                "点10" => 12,
                "点11" => 13,
                "点12" => 14,
                "点13" => 15,
                "点14" => 16,
                "点15" => 17,
                "点16" => 18,
                "点17" => 19,
                "1" => 32,
                "2" => 33,
                "3" => 34,
                "4" => 35,
                "5" => 36,
                "6" => 37,
                _ => 0
            };
            return await Task.FromResult(id);
        }

    }
    public class BlockWin : MetaData<BlockWin>
    {
        public override string TableName => "BlockWin";
        public override string KeyField => "Id";
    }
}
