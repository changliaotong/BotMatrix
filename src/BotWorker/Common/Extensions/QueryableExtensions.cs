using System.Linq.Expressions;

namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class QueryableExtensions
    {
        public static IQueryable<T> WhereIf<T>(this IQueryable<T> source, bool condition, Expression<Func<T, bool>> predicate)
            => condition ? source.Where(predicate) : source;

        public static IOrderedQueryable<T> OrderByPropertyName<T>(this IQueryable<T> source, string propertyName, bool descending = false)
        {
            var param = Expression.Parameter(typeof(T), "x");
            var prop = Expression.PropertyOrField(param, propertyName);
            var lambda = Expression.Lambda(prop, param);
            string method = descending ? "OrderByDescending" : "OrderBy";
            var result = typeof(Queryable).GetMethods()
                .First(m => m.Name == method && m.GetParameters().Length == 2)
                .MakeGenericMethod(typeof(T), prop.Type)
                .Invoke(null, new object[] { source, lambda });
            return (IOrderedQueryable<T>)result!;
        }
    }

}


