using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.Data.SqlClient;

namespace sz84.Core.Database
{
    public static partial class SQLConn
    {
        public static int ExecTrans(params string[] sqls)
        {
            return ExecTrans(true, sqls);
        }

        public static int ExecTrans(params (string Sql, SqlParameter[] Parameters)[] sqls)
        {
            return ExecTrans(true, sqls);
        }

        public static int ExecTrans(bool isDebug, params (string Sql, SqlParameter[] Parameters)[] sqls)
        {
            using SqlConnection conn = new(GetConn());
            conn.Open();
            SqlTransaction trans = conn.BeginTransaction();

            try
            {
                foreach (var (sql, parameters) in sqls)
                {
                    if (sql.IsNull()) continue;

                    using SqlCommand cmd = new(sql, conn, trans)
                    {
                        CommandType = CommandType.Text
                    };

                    if (parameters != null)
                    {
                        cmd.Parameters.AddRange(parameters);
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
            using SqlConnection conn = new(GetConn());
            conn.Open();
            SqlTransaction trans = conn.BeginTransaction();
            SqlCommand cmd = new(sqls[0], conn, trans)
            {
                CommandType = CommandType.Text
            };
            try
            {
                foreach (string sql in sqls)
                {
                    if (!sql.IsNull())
                    {
                        cmd.CommandText = sql;
                        cmd.ExecuteNonQuery();
                    }
                }
                trans.Commit();
                return 0;
            }
            catch (Exception ex)
            {
                Console.WriteLine(ex.Message);
                if (isDebug)
                    DbDebug($"{ex.Message}\n{cmd.CommandText}", "ExecTrans");

                trans.Rollback();
                return -1;
            }
        }

        public static int ExecTrans(List<(string sql, SqlParameter[] parameters)> sqls, bool isDebug = true)
        {
            using SqlConnection conn = new(GetConn());
            conn.Open();
            using SqlTransaction trans = conn.BeginTransaction();

            try
            {
                foreach (var (sql, parameters) in sqls)
                {
                    using SqlCommand cmd = new(sql, conn, trans)
                    {
                        CommandType = CommandType.Text
                    };

                    if (parameters != null)
                    {
                        var paras = ProcessParameters(parameters);
                        cmd.Parameters.AddRange(paras);
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
                    Debug($"ExecTrans:{ex.Message}\n{ex.StackTrace}", "ExecTrans");
                trans.Rollback();
                return -1;
            }
        }

        // 事务封装，避免共享静态变量，避免并发问题
        public static SqlTransaction BeginTransaction(SqlConnection conn)
        {
            if (conn.State != ConnectionState.Open)
                conn.Open();
            return conn.BeginTransaction();
        }

        public static void CommitTransaction(SqlTransaction trans)
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

        public static void RollbackTransaction(SqlTransaction trans)
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
