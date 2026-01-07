using Microsoft.Data.SqlClient;
using System.Data;
using System.Text;
using BotWorker.Core.Database;

namespace BotWorker.Infrastructure.Utils
{
    public class CodeGenerator
    {
        public static string GenerateClasses()
        {
            string res = string.Empty;
            using var connection = new SqlConnection(GetConn());
            connection.Open();
            var schema = connection.GetSchema("Tables");
            foreach (DataRow row in schema.Rows)
            {
                string tableName = row["TABLE_NAME"].AsString();
                res += GenerateClassForTable(connection, tableName) + "\n";
            }
            return res;
        }

        private static string GenerateClassForTable(SqlConnection connection, string tableName)
        {
            var sb = new StringBuilder();
            sb.AppendLine($"public class {tableName}");
            sb.AppendLine("{");

            var command = new SqlCommand($"SELECT * FROM {tableName} WHERE 1 = 0", connection);
            using (var reader = command.ExecuteReader(CommandBehavior.SchemaOnly))
            {
                var schemaTable = reader.GetSchemaTable();
                foreach (DataRow row in schemaTable.Rows)
                {
                    string columnName = row["ColumnName"].AsString();
                    string columnType = GetCSharpType(row["DataType"].AsString());
                    sb.AppendLine($"    public {columnType} {columnName} {{ get; set; }}");
                }
            }

            sb.AppendLine("}");
            return sb.AsString();
        }

        private static string GetCSharpType(string sqlType)
        {
            // 处理可空类型
            if (sqlType.StartsWith("System.Nullable"))
            {
                // 提取内部类型
                string innerType = sqlType.Split('<')[1].TrimEnd('>').Trim();
                return innerType switch
                {
                    "System.Int32" => "int?",
                    "System.String" => "string?",// string 是可空类型
                    "System.DateTime" => "DateTime?",
                    "System.Boolean" => "bool?",
                    "System.Double" => "double?",
                    "System.Single" => "float?",
                    "System.Decimal" => "decimal?",
                    "System.Int64" => "long?",
                    "System.Byte" => "byte?",
                    "System.Char" => "char?",
                    "System.Guid" => "Guid?",
                    "System.TimeSpan" => "TimeSpan?",
                    "System.Int16" => "short?",
                    "System.SByte" => "sbyte?",
                    // 可根据需要添加更多的可空类型映射
                    _ => "object?",// 默认返回可空对象
                };
            }

            // 处理常规类型
            return sqlType switch
            {
                "System.Int32" => "int",
                "System.String" => "string",// string 在 C# 中是可空类型
                "System.DateTime" => "DateTime",
                "System.Boolean" => "bool",
                "System.Double" => "double",
                "System.Single" => "float",
                "System.Decimal" => "decimal",
                "System.Int64" => "long",
                "System.Byte" => "byte",
                "System.Char" => "char",
                "System.Guid" => "Guid",
                "System.Object" => "object",
                "System.TimeSpan" => "TimeSpan",
                "System.Int16" => "short",
                "System.SByte" => "sbyte",
                "System.DBNull" => "null",// 处理DBNull
                                          // 可根据需要添加更多映射
                _ => "object",// 默认返回object
            };
        }
    }

}
