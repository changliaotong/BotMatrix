namespace BotWorker.Common.Exts
{
    public static class CollectionExtensions
    {
        public static bool IsNullOrEmpty<T>(this IEnumerable<T>? source)
            => source == null || !source.Any();

        public static void ForEach<T>(this IEnumerable<T> source, Action<T> action)
        {
            ArgumentNullException.ThrowIfNull(source);
            ArgumentNullException.ThrowIfNull(action);
            foreach (var item in source) action(item);
        }

        public static IEnumerable<IEnumerable<T>> ChunkBy<T>(this IEnumerable<T> source, int size)
        {
            return source.Select((x, i) => new { Index = i, Value = x })
                         .GroupBy(x => x.Index / size)
                         .Select(g => g.Select(x => x.Value));
        }

        public static IEnumerable<T> DistinctBy<T, TKey>(this IEnumerable<T> source, Func<T, TKey> keySelector)
            => source.GroupBy(keySelector).Select(g => g.First());

        // 将一个对象包装成 IEnumerable<T>
        public static IEnumerable<T> AsEnumerable<T>(this T item)
        {
            yield return item;
        }

        // 分页简单实现
        public static IEnumerable<T> Paginate<T>(this IEnumerable<T> source, int pageIndex, int pageSize)
        {
            if (pageIndex < 0) throw new ArgumentException("pageIndex must be >= 0", nameof(pageIndex));
            if (pageSize <= 0) throw new ArgumentException("pageSize must be > 0", nameof(pageSize));
            return source.Skip(pageIndex * pageSize).Take(pageSize);
        }
    }
}
