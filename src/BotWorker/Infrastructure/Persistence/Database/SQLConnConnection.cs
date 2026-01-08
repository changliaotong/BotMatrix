using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
namespace BotWorker.Infrastructure.Persistence.Database
{
    public static partial class SQLConn
    {
        // 不用共享静态连接，改为每次新建连接
        public static string GetConn()
        {
            return ConnString;
        }

        // 打开连接时返回 IDbConnection，方便使用 using
        public static IDbConnection OpenConnection()
        {
            var connection = DbProviderFactory.CreateConnection();
            connection.Open();
            return connection;
        }
    }
}

