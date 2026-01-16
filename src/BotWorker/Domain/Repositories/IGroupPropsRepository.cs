using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IGroupPropsRepository
    {
        Task<long> GetIdAsync(long groupId, long userId, long propId);
        Task<bool> HavePropAsync(long groupId, long userId, long propId);
        Task<int> UsePropAsync(long groupId, long userId, long propId, long qqProp);
        Task<string> GetMyPropListAsync(long groupId, long userId);
        Task<int> InsertAsync(long groupId, long userId, long propId, IDbTransaction? trans = null);
    }

    public interface IPropRepository
    {
        Task<long> GetIdAsync(string propName);
        Task<string> GetPropListAsync();
        Task<int> GetPropPriceAsync(long propId);
    }
}
