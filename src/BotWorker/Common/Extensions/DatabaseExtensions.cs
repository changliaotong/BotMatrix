using System.Data;
using System.Data.Common;

namespace BotWorker.Common.Extensions
{
    public static class DatabaseExtensions
    {
        public static async Task CommitAsync(this IDbTransaction transaction, CancellationToken cancellationToken = default)
        {
            if (transaction is DbTransaction dbTransaction)
            {
                await dbTransaction.CommitAsync(cancellationToken);
            }
            else
            {
                transaction.Commit();
                await Task.CompletedTask;
            }
        }

        public static async Task RollbackAsync(this IDbTransaction transaction, CancellationToken cancellationToken = default)
        {
            if (transaction is DbTransaction dbTransaction)
            {
                await dbTransaction.RollbackAsync(cancellationToken);
            }
            else
            {
                transaction.Rollback();
                await Task.CompletedTask;
            }
        }
    }
}
