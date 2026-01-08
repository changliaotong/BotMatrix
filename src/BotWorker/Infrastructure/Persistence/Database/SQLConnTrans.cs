using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using BotWorker.Infrastructure.Extensions;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
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
            using var conn = DbProviderFactory.CreateConnection();
            conn.Open();
            using var trans = conn.BeginTransaction();

            try
            {
                foreach (var (sql, parameters) in sqls)
                {
                    if (sql.IsNull()) continue;

                    using var cmd = conn.CreateCommand();
                    cmd.CommandText = sql;
                    cmd.Transaction = trans;

                    if (parameters != null)
                    {
                        var processedParameters = ProcessParameters(parameters);
                        foreach (var p in processedParameters) cmd.Parameters.Add(p);
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
                        trans.Rollback();
                        return -1;
                    }
                }

                trans.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTrans");
                else
                    Debug($"ExecTrans:{ex.Message}\nTransaction failed.");
                trans.Rollback();
                return -1;
            }
        }


        // 批量执行sql命令 事务提交
        public static int ExecTrans(bool isDebug, params string[] sqls)
        {
            using var conn = DbProviderFactory.CreateConnection();
            conn.Open();
            using var trans = conn.BeginTransaction();
            try
            {
                foreach (string sql in sqls)
                {
                    if (!sql.IsNull())
                    {
                        using var cmd = conn.CreateCommand();
                        cmd.CommandText = sql;
                        cmd.Transaction = trans;
                        cmd.ExecuteNonQuery();
                    }
                }
                trans.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\nTransaction failed.", "ExecTrans");
                else
                    Debug($"ExecTrans:{ex.Message}\nTransaction failed.");
                trans.Rollback();
                return -1;
            }
        }

        public static int ExecTrans(List<(string sql, IDataParameter[] parameters)> sqls, bool isDebug = true)
        {
            using var conn = DbProviderFactory.CreateConnection();
            conn.Open();
            using var trans = conn.BeginTransaction();

            try
            {
                foreach (var (sql, parameters) in sqls)
                {
                    if (sql.IsNull()) continue;

                    using var cmd = conn.CreateCommand();
                    cmd.CommandText = sql;
                    cmd.Transaction = trans;

                    if (parameters != null)
                    {
                        var processedParameters = ProcessParameters(parameters);
                        foreach (var p in processedParameters) cmd.Parameters.Add(p);
                    }

                    cmd.ExecuteNonQuery();
                }

                trans.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                if (isDebug)
                    DbDebug($"{ex.Message}\n{ex.StackTrace}", "ExecTrans");
                else
                    DbDebug($"ExecTrans:{ex.Message}\n{ex.StackTrace}", "ExecTrans");
                trans.Rollback();
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

