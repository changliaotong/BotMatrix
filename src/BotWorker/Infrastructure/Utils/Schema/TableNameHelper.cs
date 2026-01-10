namespace BotWorker.Infrastructure.Utils.Schema;

public static class TableNameHelper
{
    public static string GetTableName<T>()
    {
        var type = typeof(T);
        
        // 尝试获取 TableName 属性（兼容 MetaData 结构）
        var prop = type.GetProperty("TableName", System.Reflection.BindingFlags.Public | System.Reflection.BindingFlags.Instance | System.Reflection.BindingFlags.Static | System.Reflection.BindingFlags.FlattenHierarchy);
        if (prop != null)
        {
            try 
            {
                // 如果是静态属性直接获取
                if (prop.GetGetMethod()?.IsStatic == true)
                {
                    var value = prop.GetValue(null)?.ToString();
                    if (!string.IsNullOrEmpty(value)) return value;
                }
                else 
                {
                    // 如果是实例属性尝试创建临时实例
                    var instance = System.Activator.CreateInstance(type);
                    var value = prop.GetValue(instance)?.ToString();
                    if (!string.IsNullOrEmpty(value)) return value;
                }
            }
            catch (Exception ex)
            {
                System.Console.WriteLine($"[TableNameHelper] Error getting TableName for {type.Name}: {ex.Message}");
            }
        }

        // 尝试获取静态 FullName 字段并解析
        var fullNameField = type.GetField("FullName", System.Reflection.BindingFlags.Public | System.Reflection.BindingFlags.Static | System.Reflection.BindingFlags.FlattenHierarchy);
        if (fullNameField != null)
        {
            var fullName = fullNameField.GetValue(null)?.ToString();
            if (!string.IsNullOrEmpty(fullName))
            {
                // FullName 格式通常是 [Db].[dbo].[Table]
                var parts = fullName.Split('.');
                var lastPart = parts.Last().Trim('[', ']');
                if (!string.IsNullOrEmpty(lastPart)) return lastPart;
            }
        }

        return type.Name.ToLower(); 
    }
}
