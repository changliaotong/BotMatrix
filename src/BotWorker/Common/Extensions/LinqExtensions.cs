using System.Linq.Expressions;

namespace BotWorker.Common.Extensions
{
    public static class LinqExtensions
    {        
        /// <summary>
        /// 判断集合是否为空（null �?Count == 0�?
        /// </summary>
        public static bool IsNullOrEmpty<T>(this IEnumerable<T>? source)
            => source == null || !source.Any();

        /// <summary>
        /// 判断集合是否不为空（�?null 且至少有一个元素）
        /// </summary>
        public static bool IsNotEmpty<T>(this IEnumerable<T>? source)
            => source != null && source.Any();

        /// <summary>
        /// 对集合执行指定操作（foreach 的链式语法）
        /// </summary>
        public static IEnumerable<T> ForEach<T>(this IEnumerable<T> source, Action<T> action)
        {
            foreach (var item in source)
                action(item);
            return source;
        }

        /// <summary>
        /// 将对象集合转为逗号分隔的字符串（可自定义格式）
        /// </summary>
        public static string JoinAsString<T>(this IEnumerable<T> source, string separator = ",", Func<T, string>? selector = null)
            => string.Join(separator, selector == null ? source : source.Select(selector));

        // 动态构造条件：判断某属性是否等于指定�?
        public static IQueryable<T> WhereEquals<T>(this IQueryable<T> source, string propertyName, object? value)
        {
            ArgumentNullException.ThrowIfNull(source);
            if (string.IsNullOrEmpty(propertyName)) throw new ArgumentNullException(nameof(propertyName));

            var parameter = Expression.Parameter(typeof(T), "x");
            var property = Expression.Property(parameter, propertyName);
            var constant = Expression.Constant(value, property.Type);

            var equal = Expression.Equal(property, constant);
            var lambda = Expression.Lambda<Func<T, bool>>(equal, parameter);

            return source.Where(lambda);
        }

        // 支持多条件动态构�?AND 组合
        public static IQueryable<T> WhereAll<T>(this IQueryable<T> source, IEnumerable<(string Property, object? Value)> conditions)
        {
            ArgumentNullException.ThrowIfNull(source);
            ArgumentNullException.ThrowIfNull(conditions);

            var parameter = Expression.Parameter(typeof(T), "x");
            Expression? body = null;

            foreach (var (Property, Value) in conditions)
            {
                var property = Expression.Property(parameter, Property);
                var constant = Expression.Constant(Value, property.Type);
                var equal = Expression.Equal(property, constant);
                body = body == null ? equal : Expression.AndAlso(body, equal);
            }

            if (body == null)
                return source;

            var lambda = Expression.Lambda<Func<T, bool>>(body, parameter);
            return source.Where(lambda);
        }
    }
}


