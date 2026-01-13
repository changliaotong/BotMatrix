using System;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false)]
    public class DisplayInListAttribute(bool show) : Attribute
    {
        public bool Show { get; } = show;
    }

    [AttributeUsage(AttributeTargets.Property, AllowMultiple = false)]
    public class DisplayInEditAttribute(bool show) : Attribute
    {
        public bool Show { get; } = show;
    }

    /// <summary>
    /// 标记实体使用 PostgreSQL 数据库（AI 相关功能专用）
    /// </summary>
    [AttributeUsage(AttributeTargets.Class, AllowMultiple = false)]
    public class UsePostgreSqlAttribute : Attribute
    {
    }
}