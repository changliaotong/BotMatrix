using System.Data;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Interfaces
{
    public interface IBlockService
    {
        Task<long> GetIdAsync(long groupId, long userId, IDbTransaction? trans = null);
        Task<string> GetHashAsync(long blockId, IDbTransaction? trans = null);
        Task<long> GetBlockIdAsync(string hash);
        Task<int> GetNumAsync(long botUin, long groupId, string groupName, long qq, string name, IDbTransaction? trans = null);
        Task<int> GetNumAsync(long blockId, IDbTransaction? trans = null);
        Task<decimal> GetOddsAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null);
        Task<bool> IsWinAsync(int typeId, string typeName, int blockNum, IDbTransaction? trans = null);
        Task<string> GetValueAsync(string field, long blockId, IDbTransaction? trans = null);
        Task<bool> IsOpenAsync(long blockId, IDbTransaction? trans = null);
        Task<string> GetBlockSecretAsync(long blockId, IDbTransaction? trans = null);
        Task<string> GetCmdAsync(string cmdName, long qq, IDbTransaction? trans = null);
        string GetCmd(string cmdName, long qq);
        string GetBlockInfo16(string hash16);
        (string sql, object paras) SqlAppend(long botUin, long groupId, string groupName, long userId, string name, string prevRes, string blockRes, string blockRand, string blockInfo, string blockHash, long prevId);
        (string sql, object paras) SqlUpdateOpen(long botUin, long userId, string name, long prevId);
    }
}
