using System.Data;

namespace BotWorker.Infrastructure.Extensions;

public static class DataTableExtensions
{
    public static List<T> ToList<T>(this DataTable table) where T : new()
    {
        var list = new List<T>();
        foreach (DataRow row in table.Rows)
        {
            list.Add(row.ToModel<T>());
        }
        return list;
    }
}

