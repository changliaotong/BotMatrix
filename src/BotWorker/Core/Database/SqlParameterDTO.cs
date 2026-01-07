using System.Data;
using Microsoft.Data.SqlClient;

namespace BotWorker.Core.Database
{
    public class SqlParameterDTO
    {
        public string Name { get; set; } = "";
        public object? Value { get; set; }
        public string DbType { get; set; } = "";

        public static List<SqlParameter> ToParameters(List<SqlParameterDTO> dtos)
        {
            return dtos.Select(dto => new SqlParameter(dto.Name, ConvertDbType(dto.DbType)) { Value = dto.Value ?? DBNull.Value }).ToList();
        }

        public static List<SqlParameterDTO> From(SqlParameter[]? parameters)
        {
            if (parameters == null) return new();
            return parameters.Select(p => new SqlParameterDTO
            {
                Name = p.ParameterName,
                Value = p.Value,
                DbType = p.DbType.ToString()
            }).ToList();
        }

        private static SqlDbType ConvertDbType(string type)
        {
            return Enum.TryParse<SqlDbType>(type, out var result) ? result : SqlDbType.VarChar;
        }
    }
}
