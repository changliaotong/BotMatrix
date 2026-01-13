using System.Data;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public class DbParameterDTO
    {
        public string Name { get; set; } = "";
        public object? Value { get; set; }
        public string DbType { get; set; } = "";

        public static List<IDataParameter> ToParameters(List<DbParameterDTO> dtos)
        {
            return dtos.Select(dto => DbProviderFactory.CreateParameter(dto.Name, dto.Value ?? DBNull.Value)).ToList();
        }

        public static List<DbParameterDTO> From(IDataParameter[]? parameters)
        {
            if (parameters == null) return new();
            return parameters.Select(p => new DbParameterDTO
            {
                Name = p.ParameterName,
                Value = p.Value,
                DbType = p.DbType.ToString()
            }).ToList();
        }
    }
}

