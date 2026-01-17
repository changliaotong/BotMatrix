using System;
using System.Collections.Generic;
using System.Data;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Persistence;

namespace BotWorker.Domain.Repositories
{
    public interface IBaseRepository<T> where T : class
    {
        Task<T?> GetByIdAsync(long id);
        Task<IEnumerable<T>> GetAllAsync();
        Task<bool> DeleteAsync(long id);
        Task<int> DeleteAsync(string conditions, object? parameters = null, IDbTransaction? trans = null);
        Task<SqlHelper.TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null);
        Task<TValue> GetValueAsync<TValue>(string field, long id, IDbTransaction? trans = null);
        Task<int> SetValueAsync(string field, object value, long id, IDbTransaction? trans = null);
        Task<int> IncrementValueAsync(string field, object value, long id, IDbTransaction? trans = null);
        Task<int> UpdateAsync(string fieldsSql, long id, IDbTransaction? trans = null);
        Task<int> UpdateAsync(string fieldsSql, string conditions, object? parameters = null, IDbTransaction? trans = null);
        Task<long> InsertAsync(T entity, IDbTransaction? trans = null);
        Task<bool> UpdateEntityAsync(T entity, IDbTransaction? trans = null);
        Task<long> CountAsync(string conditions = "", object? parameters = null, IDbTransaction? trans = null);
        Task EnsureTableCreatedAsync();
    }
}
