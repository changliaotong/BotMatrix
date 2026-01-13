using System.Data;
using System.Data.Common;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        public static async Task<int> ExecTransAsync(params string[] sqls)
        {
            return await ExecTransAsync(true, sqls);
        }

        public static async Task<int> ExecTransAsync(bool isDebug, params (string Sql, IDataParameter[] Parameters)[] sqls)
        {
            // 预处理参数，减少事务内耗时
            var processedSqls = new List<(string Sql, IDataParameter[] Parameters)>();
            foreach (var item in sqls)
            {
                if (item.Sql.IsNull()) continue;
                processedSqls.Add((item.Sql, ProcessParameters(item.Parameters)));
            }

            if (processedSqls.Count == 0) return 0;

            using var wrapper = await MetaData.BeginTransactionAsync();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (var (sql, parameters) in processedSqls)
                {
                    LogSql(sql, parameters);
                    using var cmdObject = conn.CreateCommand();
                    if (cmdObject is not DbCommand cmd)
                        throw new NotSupportedException("Command must be a DbCommand to support async operations.");

                    cmd.CommandText = sql;
                    var unwrappedTrans = MetaData.Unwrap(trans);
                    if (unwrappedTrans == null)
                        throw new InvalidOperationException("Could not unwrap transaction.");
                    cmd.Transaction = (DbTransaction)unwrappedTrans;

                    if (parameters != null)
                    {
                        foreach (var p in parameters) cmd.Parameters.Add(p);
                    }

                    try
                    {
                        await cmd.ExecuteNonQueryAsync();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTransAsync");
                        else
                            Debug($"ExecTransAsync:{ex.Message}\n{cmd.CommandText}");
                        await wrapper.RollbackAsync();
                        return -1;
                    }
                }

                await wrapper.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTransAsync");
                else
                    Debug($"ExecTransAsync:{ex.Message}\nTransaction failed.");
                await wrapper.RollbackAsync();
                return -1;
            }
        }

        // 批量执行sql命令 事务提交
        public static async Task<int> ExecTransAsync(bool isDebug, params string[] sqls)
        {
            var validSqls = sqls.Where(s => !s.IsNull()).ToList();
            if (validSqls.Count == 0) return 0;

            using var wrapper = await MetaData.BeginTransactionAsync();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (string sql in validSqls)
                {
                    LogSql(sql, null);
                    using var cmdObject = conn.CreateCommand();
                    if (cmdObject is not DbCommand cmd)
                        throw new NotSupportedException("Command must be a DbCommand to support async operations.");

                    cmd.CommandText = sql;
                    var unwrappedTrans = MetaData.Unwrap(trans);
                    if (unwrappedTrans == null)
                        throw new InvalidOperationException("Could not unwrap transaction.");
                    cmd.Transaction = (DbTransaction)unwrappedTrans;
                    
                    try
                    {
                        await cmd.ExecuteNonQueryAsync();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTransAsync");
                        else
                            Debug($"ExecTransAsync:{ex.Message}\n{cmd.CommandText}");
                        await wrapper.RollbackAsync();
                        return -1;
                    }
                }
                await wrapper.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTransAsync");
                else
                    Debug($"ExecTransAsync:{ex.Message}\nTransaction failed.");
                await wrapper.RollbackAsync();
                return -1;
            }
        }

        public static async Task<int> ExecTransAsync(List<(string sql, IDataParameter[] parameters)> sqls, bool isDebug = true)
        {
            // 预处理参数
            var processedSqls = new List<(string Sql, IDataParameter[] Parameters)>();
            foreach (var item in sqls)
            {
                if (item.sql.IsNull()) continue;
                processedSqls.Add((item.sql, ProcessParameters(item.parameters)));
            }

            if (processedSqls.Count == 0) return 0;

            using var wrapper = await MetaData.BeginTransactionAsync();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (var (sql, parameters) in processedSqls)
                {
                    LogSql(sql, parameters);
                    using var cmdObject = conn.CreateCommand();
                    if (cmdObject is not DbCommand cmd)
                        throw new NotSupportedException("Command must be a DbCommand to support async operations.");

                    cmd.CommandText = sql;
                    var unwrappedTrans = MetaData.Unwrap(trans);
                    if (unwrappedTrans == null)
                        throw new InvalidOperationException("Could not unwrap transaction.");
                    cmd.Transaction = (DbTransaction)unwrappedTrans;

                    if (parameters != null)
                    {
                        foreach (var p in parameters) cmd.Parameters.Add(p);
                    }

                    try
                    {
                        await cmd.ExecuteNonQueryAsync();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTransAsync");
                        else
                            Debug($"ExecTransAsync:{ex.Message}\n{cmd.CommandText}");
                        await wrapper.RollbackAsync();
                        return -1;
                    }
                }

                await wrapper.CommitAsync();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{ex.StackTrace}", "ExecTransAsync");
                else
                    DbDebug($"ExecTransAsync:{ex.Message}\n{ex.StackTrace}", "ExecTransAsync");
                await wrapper.RollbackAsync();
                return -1;
            }
        }

        public static int ExecTrans(params string[] sqls)
        {
            return ExecTrans(true, sqls);
        }

        public static int ExecTrans(params (string Sql, IDataParameter[] Parameters)[] sqls)
        {
            return ExecTrans(true, sqls);
        }

        public static int ExecTrans(bool isDebug, params (string Sql, IDataParameter[] Parameters)[] sqls)
        {
            // 预处理参数
            var processedSqls = new List<(string Sql, IDataParameter[] Parameters)>();
            foreach (var item in sqls)
            {
                if (item.Sql.IsNull()) continue;
                processedSqls.Add((item.Sql, ProcessParameters(item.Parameters)));
            }

            if (processedSqls.Count == 0) return 0;

            using var wrapper = MetaData.BeginTransaction();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (var (sql, parameters) in processedSqls)
                {
                    LogSql(sql, parameters);
                    using var cmd = conn.CreateCommand();
                    cmd.CommandText = sql;
                    cmd.Transaction = MetaData.Unwrap(trans);

                    if (parameters != null)
                    {
                        foreach (var p in parameters) cmd.Parameters.Add(p);
                    }

                    try
                    {
                        cmd.ExecuteNonQuery();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTrans");
                        else
                            Debug($"ExecTrans:{ex.Message}\n{cmd.CommandText}");
                        wrapper.Rollback();
                        return -1;
                    }
                }

                wrapper.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTrans");
                else
                    Debug($"ExecTrans:{ex.Message}\nTransaction failed.");
                wrapper.Rollback();
                return -1;
            }
        }


        // 批量执行sql命令 事务提交
        public static int ExecTrans(bool isDebug, params string[] sqls)
        {
            var validSqls = sqls.Where(s => !s.IsNull()).ToList();
            if (validSqls.Count == 0) return 0;

            using var wrapper = MetaData.BeginTransaction();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (string sql in validSqls)
                {
                    LogSql(sql, null);
                    using var cmd = conn.CreateCommand();
                    cmd.CommandText = sql;
                    cmd.Transaction = MetaData.Unwrap(trans);
                    
                    try
                    {
                        cmd.ExecuteNonQuery();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTrans");
                        else
                            Debug($"ExecTrans:{ex.Message}\n{cmd.CommandText}");
                        wrapper.Rollback();
                        return -1;
                    }
                }
                wrapper.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTrans");
                else
                    Debug($"ExecTrans:{ex.Message}\nTransaction failed.");
                wrapper.Rollback();
                return -1;
            }
        }

        public static int ExecTrans(List<(string sql, IDataParameter[] parameters)> sqls, bool isDebug = true)
        {
            // 预处理参数
            var processedSqls = new List<(string Sql, IDataParameter[] Parameters)>();
            foreach (var item in sqls)
            {
                if (item.sql.IsNull()) continue;
                processedSqls.Add((item.sql, ProcessParameters(item.parameters)));
            }

            if (processedSqls.Count == 0) return 0;

            using var wrapper = MetaData.BeginTransaction();
            var trans = wrapper.Transaction;
            var conn = trans.Connection;

            if (conn == null)
                throw new InvalidOperationException("Transaction connection is null.");

            try
            {
                foreach (var (sql, parameters) in processedSqls)
                {
                    LogSql(sql, parameters);
                    using var cmd = conn.CreateCommand();
                    cmd.CommandText = sql;
                    cmd.Transaction = MetaData.Unwrap(trans);

                    if (parameters != null)
                    {
                        foreach (var p in parameters) cmd.Parameters.Add(p);
                    }

                    try
                    {
                        cmd.ExecuteNonQuery();
                    }
                    catch (Exception ex)
                    {
                        if (isDebug)
                            DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTrans");
                        else
                            Debug($"ExecTrans:{ex.Message}\n{cmd.CommandText}");
                        wrapper.Rollback();
                        return -1;
                    }
                }

                wrapper.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{ex.StackTrace}", "ExecTrans");
                else
                    DbDebug($"ExecTrans:{ex.Message}\n{ex.StackTrace}", "ExecTrans");
                wrapper.Rollback();
                return -1;
            }
        }

        // 事务封装，避免共享静态变量，避免并发问题
        public static IDbTransaction BeginTransaction(IDbConnection conn)
        {
            if (conn.State != ConnectionState.Open)
                conn.Open();
            return conn.BeginTransaction();
        }

        public static void CommitTransaction(IDbTransaction trans)
        {
            if (trans == null) return;
            try
            {
                trans.Commit();
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "commit trans");
                trans.Rollback();
                throw;
            }
            finally
            {
                trans.Dispose();
                trans.Connection?.Close();
            }
        }

        public static void RollbackTransaction(IDbTransaction trans)
        {
            if (trans == null) return;
            try
            {
                trans.Rollback();
            }
            catch (Exception ex)
            {
                DbDebug(ex.Message, "rollback trans");
                throw;
            }
            finally
            {
                trans.Dispose();
                trans.Connection?.Close();
            }
        }
    }
}

