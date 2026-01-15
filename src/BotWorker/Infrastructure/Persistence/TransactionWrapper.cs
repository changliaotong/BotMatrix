using System.Data;
using System.Data.Common;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Persistence;

/// <summary>
/// 事务包装器，确保即使忘记调用 Commit/Rollback，连接也能通过 Dispose 正确释放
/// </summary>
public class TransactionWrapper : IDbTransaction, IAsyncDisposable
{
    private readonly IDbTransaction _inner;
    private readonly bool _ownConnection;
    private bool _disposed;
    private bool _committed;
    private bool _rolledBack;

    public TransactionWrapper(IDbTransaction inner, bool ownConnection = true)
    {
        _inner = inner;
        _ownConnection = ownConnection;
    }

    public IDbTransaction Transaction => _inner;
    public IDbConnection? Connection => _inner.Connection;
    public IsolationLevel IsolationLevel => _inner.IsolationLevel;

    public static async Task<TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
    {
        if (existingTrans != null) 
        {
            var trans = existingTrans;
            // 如果已经是包装器，解包以避免多层包装
            if (trans is TransactionWrapper wrapper)
                trans = wrapper.Transaction;

            return new TransactionWrapper(trans!, false);
        }
        
        var conn = Database.DbProviderFactory.CreateConnection();
        if (conn is DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
        var newTrans = conn.BeginTransaction(level);
        return new TransactionWrapper(newTrans, true);
    }

    public static async Task<T> ExecuteAsync<T>(Func<TransactionWrapper, Task<T>> action, IDbTransaction? trans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
    {
        await using var wrapper = await BeginTransactionAsync(trans, level);
        try
        {
            var result = await action(wrapper);
            await wrapper.CommitAsync();
            return result;
        }
        catch
        {
            await wrapper.RollbackAsync();
            throw;
        }
    }

    public static async Task ExecuteAsync(Func<TransactionWrapper, Task> action, IDbTransaction? trans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
    {
        await using var wrapper = await BeginTransactionAsync(trans, level);
        try
        {
            await action(wrapper);
            await wrapper.CommitAsync();
        }
        catch
        {
            await wrapper.RollbackAsync();
            throw;
        }
    }

    public void Commit()
    {
        if (_disposed || _committed || _rolledBack) return;
        if (_ownConnection) _inner.Commit();
        _committed = true;
    }

    public async Task CommitAsync()
    {
        if (_disposed || _committed || _rolledBack) return;
        if (_ownConnection) 
        {
            if (_inner is DbTransaction dbTrans) await dbTrans.CommitAsync();
            else _inner.Commit();
        }
        _committed = true;
    }

    public void Rollback()
    {
        if (_disposed || _committed || _rolledBack) return;
        _inner.Rollback();
        _rolledBack = true;
    }

    public async Task RollbackAsync()
    {
        if (_disposed || _committed || _rolledBack) return;
        if (_inner is DbTransaction dbTrans) await dbTrans.RollbackAsync();
        else _inner.Rollback();
        _rolledBack = true;
    }

    public void Dispose()
    {
        if (!_disposed)
        {
            if (!_committed && !_rolledBack && _ownConnection)
            {
                try { Rollback(); } catch { /* Ignore */ }
            }

            if (_ownConnection)
            {
                var conn = _inner.Connection;
                _inner.Dispose();
                conn?.Dispose();
            }
            _disposed = true;
        }
    }

    public async ValueTask DisposeAsync()
    {
        if (!_disposed)
        {
            if (!_committed && !_rolledBack && _ownConnection)
            {
                try { await RollbackAsync(); } catch { /* Ignore */ }
            }

            if (_ownConnection)
            {
                var conn = _inner.Connection;
                if (_inner is DbTransaction dbTrans)
                    await dbTrans.DisposeAsync();
                else
                    _inner.Dispose();

                if (conn is DbConnection dbConn)
                    await dbConn.DisposeAsync();
                else
                    conn?.Dispose();
            }
            _disposed = true;
        }
    }
}
