using Microsoft.Data.SqlClient;

namespace BotWorker.Infrastructure.Persistence.ORM
{    public class QueryOptions
    {
        public int? PageIndex { get; set; }     // 从1开始
        public int? PageSize { get; set; }      // 每页数量
        public int? Top { get; set; }           // 指定取前几条
        public bool? GetAll { get; set; }       // 是否全部获取
        public string? FilterSql { get; set; }  // 可直接传 SQL WHERE 条件
        public SqlParameter[] Parameters { get; set; } = [];

        // 可选排序字段和方向
        public string? OrderBy { get; set; }
    }
}

