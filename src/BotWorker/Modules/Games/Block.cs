using Microsoft.Data.SqlClient;
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
                
        public static ((string sql, SqlParameter[] paras) sqlInsert, (string sql, SqlParameter[] paras) sqlUpdatePrev) SqlAppend(long botUin, long groupId, string groupName, long userId, string name, string prevRes)
        {
            long prevId = GetId(groupId, userId);
            string prevHash = prevId == 0
                ? groupId == 0 ? UserInfo.GetGuid(userId).AsString().Sha256() : GroupInfo.GetGuid(groupId).AsString().Sha256()
                : GetHash(prevId);
            string hashRobot = BotInfo.GetBotGuid(botUin).Sha256();
            string hashRoom = groupId == 0 ? "" : GroupInfo.GetGuid(groupId).AsString().Sha256();
            string hashClient = UserInfo.GetGuid(userId).AsString().Sha256();
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

            var sql2 = SqlSetValues($"IsOpen=1, OpenDate=GETDATE(), OpenBotUin={botUin}, OpenUserId={userId}, OpenUserName={name.Quotes()}", prevId);
            return (sql1, sql2);
        }

        public static async Task<long> GetIdAsync(long groupId, long userId)
        {
            string sql = $"SELECT ISNULL(MAX(Id),0) AS res FROM {FullName} WHERE GroupId = {groupId} AND IsOpen = 0";
            var res = await ExecScalarAsync<long>(groupId == 0 ? $"{sql} AND UserId = {userId} " : sql);
            return res ?? 0;
        }

        public static async Task<string> GetHashAsync(long blockId)
        {
            if (blockId == 0) return string.Empty;
            var res = await QueryScalarAsync<string>($"SELECT BlockHash FROM {FullName} WHERE Id = {blockId}");
            return res ?? string.Empty;
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, string groupName, long userId, string name, string prevRes, (string sql, SqlParameter[] paras) sqlAddCredit, (string sql, SqlParameter[] paras) sqlCreditHis)
        {
            long prevId = await GetIdAsync(groupId, userId);
            string prevHash = prevId == 0 
                ? groupId == 0 ? UserInfo.GetGuid(userId).AsString().Sha256() : GroupInfo.GetGuid(groupId).AsString().Sha256()
                : await GetHashAsync(prevId);
            string hashRobot = BotInfo.GetBotGuid(botUin).Sha256();
            string hashRoom = groupId == 0 ? "" : GroupInfo.GetGuid(groupId).AsString().Sha256();            
            string hashClient = UserInfo.GetGuid(userId).AsString().Sha256();
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

            var sql2 = SqlSetValues($"IsOpen=1, OpenDate=GETDATE(), OpenBotUin={botUin}, OpenUserId={userId}, OpenUserName={name.Quotes()}", prevId);
            
            using var trans = await BeginTransactionAsync();
            try
            {
                await ExecAsync(sql.sql, trans, sql.paras);
                if (!sqlAddCredit.sql.IsNull()) await ExecAsync(sqlAddCredit.sql, trans, sqlAddCredit.paras);
                if (!sqlCreditHis.sql.IsNull()) await ExecAsync(sqlCreditHis.sql, trans, sqlCreditHis.paras);
                await ExecAsync(sql2.sql, trans, sql2.paras);
                await trans.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Block.AppendAsync error: {ex.Message}");
                await trans.RollbackAsync();
                return -1;
            }
        }


        public static string GetCmd(string cmdName, long qq)
        {
            cmdName = cmdName.ToLower() switch
            {
                "jd" => "剪刀",
                "st" => "石头",
                "bu" => "布",
                "d" => "大",
                "x" => UserInfo.GetBool("Xxian", qq) ? "闲" : "小",
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

        public static string GetHash(long groupId, long qq)
            => GetHashAsync(groupId, qq).GetAwaiter().GetResult();

        public static async Task<string> GetHashAsync(long groupId, long qq)
        {
            return await GetHashAsync(await GetIdAsync(groupId, qq));
        }

        public static long GetId(long groupId, long userId)
            => GetIdAsync(groupId, userId).GetAwaiter().GetResult();

        public static async Task<long> GetBlockIdAsync(string hash)
        {
            return (await GetWhereAsync("Id", $"BlockHash = {hash.Quotes()}")).AsLong();
        }

        public static int GetNum(long botUin, long groupId, string groupName, long qq, string name)
            => GetNumAsync(botUin, groupId, groupName, qq, name).GetAwaiter().GetResult();

        public static async Task<int> GetNumAsync(long botUin, long groupId, string groupName, long qq, string name)
        {
            long blockId = await GetIdAsync(groupId, qq);
            if (blockId == 0)
            {
                if (await AppendAsync(botUin, groupId, groupName, qq, name, "创世区块", (string.Empty, Array.Empty<SqlParameter>()), (string.Empty, Array.Empty<SqlParameter>())) != -1)
                    return await GetNumAsync(botUin, groupId, groupName, qq, name);
            }
            return GetNum(blockId);
        }

        public static int GetOdds(int typeId, string typeName, int blockNum)
            => GetOddsAsync(typeId, typeName, blockNum).GetAwaiter().GetResult();

        public static async Task<int> GetOddsAsync(int typeId, string typeName, int blockNum)
        {
            if (typeId >= 32 & typeId <= 37)
                return blockNum.ToString().Split([typeName.Replace("押", "")], StringSplitOptions.None).Length - 1;
            else
                return int.Parse(await QueryAsync($"SELECT BlockOdds FROM {BlockType.FullName} WHERE Id = {typeId}"));
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
            if (blockNum.In(111, 222, 333, 444, 555, 666))
                return "围";

            return Sum(blockNum) > 10 ? "大" : "小";
        }       

        public static string GetBlockInfo16(string hash16)
        {
            long block_id = GetWhere($"ISNULL(Id, 0)", $"BlockHash LIKE {hash16.QuotesLike()}").AsLong();
            return block_id == 0
                ? ""
                : GetBlockInfo(block_id) + GetBlockSecret(block_id);
        }

        public static string GetBlockInfo(long blockId)
        {
            return GetValue("BlockInfo", blockId);
        }

        public static bool IsOpen(long blockId)
        {
            return GetBool("IsOpen", blockId);
        }

        public static string GetHash(long blockId)
        {
            return GetValue("BlockHash", blockId);
        }

        public static int GetNum(long blockId)
        {
            return GetInt("BlockNum", blockId);
        }

        public static string GetBlockSecret(long blockId)
        {
            return IsOpen(blockId)
                    ? GetValue("BlockSecret", blockId)
                    : $"本局数据:本局游戏尚未结束，保密区数据不可见\n" +
                      $"随机密码:本局游戏尚未结束，保密区数据不可见";
        }

        public static bool IsWin(int typeId, string typeName, int blockNum)
        {
            if (typeId >= 32 & typeId <= 37)
            {
                int i = blockNum.ToString().Split([typeName.Replace("押", "")], StringSplitOptions.None).Length - 1;
                return i > 0;
            }
            return (QueryScalar<string>($"select IsWin from {BlockWin.FullName} where TypeId = {typeId} and BlockNum = {blockNum}") ?? "").AsBool();
        }
    }

    public class BlockRandom : MetaData<BlockRandom>
    {
        public override string TableName => "BlockRandom";
        public override string KeyField => "Id";

        // 随机一组数字（三个1-6）
        public static int RandomNum()
        {
            return GetInt("BlockNum", RandomInt(1, 216));
        }
    }
    public class BlockType : MetaData<BlockType>
    {
        public override string TableName => "Blocktype";
        public override string KeyField => "Id";

        //类型
        public static int GetTypeId(string TypeName)
        {
            return GetWhere($"Id", $"TypeName = {TypeName.Replace("押", "").Quotes()}").AsInt();
        }

    }
    public class BlockWin : MetaData<BlockWin>
    {
        public override string TableName => "BlockWin";
        public override string KeyField => "Id";
    }
}
