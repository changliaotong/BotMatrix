using System;
using System.Collections.Generic;
using System.Linq;
using Microsoft.Data.SqlClient;

namespace BotWorker.Core.MetaDatas
{
    /// <summary>
    /// 解决外部系统直接修改数据库导致的缓存不一致问题
    /// </summary>
    public class MetaDataSync
    {
        public static string TableName => "CacheSyncQueue";

        /// <summary>
        /// 扫描同步队列并更新缓存
        /// 由 BotWorker 后台线程每秒调用
        /// </summary>
        public static void ScanAndSync()
        {
            // 1. 获取并立即删除记录（使用 OUTPUT 子句确保原子性，性能极高）
            string sql = $@"
                DELETE TOP (100) FROM {TableName} 
                OUTPUT DELETED.* 
                WHERE IsProcessed = 0";
            
            var list = SQLConn.Query(sql);
            if (list.Count == 0) return;

            foreach (var item in list)
            {
                SyncCache(item.TableName.AsString(), item.RecordId, item.RecordId2, item.FieldName.AsString(), item.NewValue);
            }
        }

        private static void SyncCache(string table, object id, object? id2, string field, object value)
        {
            // 将表名映射到具体的 MetaData 类，调用其 SyncCacheField
            if (table.Equals("User", StringComparison.OrdinalIgnoreCase))
                Bots.Entries.UserInfo.SyncCacheField(id, id2, field, value);
            else if (table.Equals("GroupMember", StringComparison.OrdinalIgnoreCase))
                Bots.Groups.GroupMember.SyncCacheField(id, id2, field, value);
            else if (table.Equals("Group", StringComparison.OrdinalIgnoreCase))
                Bots.Groups.GroupInfo.SyncCacheField(id, id2, field, value);
        }
    }
}
