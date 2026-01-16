using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.Repositories;

using System.Data;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence.ORM;
using BotWorker.Infrastructure.Persistence.Repositories;
using Dapper;

namespace BotWorker.Domain.Models.BotMessages
{
    public partial class BotMessage
    {
        // 临时使用静态属性，后续应通过依赖注入解决
        public static IUserRepository UserRepository { get; } = new UserRepository();
        public static IGroupRepository GroupRepository { get; } = new GroupRepository();
        public static IGroupMemberRepository GroupMemberRepository { get; } = new GroupMemberRepository();
        public static ISignInRepository SignInRepository { get; } = new SignInRepository();
        public static IBotRepository BotRepository { get; } = new BotRepository();
        public static ITokensLogRepository TokenLogRepository { get; } = new TokenLogRepository();
        public static ICreditLogRepository CreditLogRepository { get; } = new CreditLogRepository();
        public static IBotMessageRepository BotMessageRepository { get; } = new BotMessageRepository();

        // 兼容旧代码的辅助方法
        public static async Task<MetaData.TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null, IsolationLevel level = IsolationLevel.ReadCommitted)
            => await MetaData.BeginTransactionAsync(existingTrans, level);

        public static async Task<int> ExecAsync(string sql, params object?[] args)
        {
            var (trans, actualArgs, explicitParams) = ParseArgs(args);
            var (resolvedSql, parameters) = MetaData.ResolveSql(sql, actualArgs);
            
            using var conn = Persistence.Database.DbProviderFactory.CreateConnection();
            if (conn is System.Data.Common.DbConnection dbConn) await dbConn.OpenAsync(); else conn.Open();
            
            var dapperParams = new DynamicParameters();
            foreach (var p in parameters) dapperParams.Add(p.ParameterName, p.Value);
            if (explicitParams != null)
                foreach (var p in explicitParams) dapperParams.Add(p.ParameterName, p.Value);

            return await conn.ExecuteAsync(resolvedSql, dapperParams, trans);
        }

        private static (IDbTransaction? trans, object?[] actualArgs, IDataParameter[]? parameters) ParseArgs(object?[] args)
        {
            IDbTransaction? trans = null;
            object?[] actualArgs = args;
            IDataParameter[]? parameters = null;

            int start = 0;
            if (args.Length > 0 && args[0] is IDbTransaction t)
            {
                trans = t;
                start = 1;
            }

            if (args.Length > start && args[^1] is IDataParameter[] p)
            {
                parameters = p;
                actualArgs = args[start..^1];
            }
            else
            {
                actualArgs = args[start..];
            }

            return (trans, actualArgs, parameters);
        }

        public static string RetryMsg => "⚠️ 操作失败，请稍后再试";
        public static string CreditSystemClosed => "⚠️ 积分系统未开启";
    }
}
