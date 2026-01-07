using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using Microsoft.Data.SqlClient;

namespace BotWorker.Core.Database
{
    public static partial class SQLConn
    {
        // 不用共享静态连接，改为每次新建连接
        public static string GetConn()
        {
            return ConnString;
        }

        // 打开连接时返回 SqlConnection，方便使用 using
        public static SqlConnection OpenConnection()
        {
            var connection = new SqlConnection(ConnString);
            connection.Open();
            return connection;
        }

        public static void CloseConnection()
        {
            if (conn.State != ConnectionState.Closed)
                conn.Close();
        }
    }
}
