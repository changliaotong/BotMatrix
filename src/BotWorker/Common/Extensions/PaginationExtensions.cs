namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class PaginationExtensions
    {
        public static IQueryable<T> PageBy<T>(this IQueryable<T> query, int pageIndex, int pageSize)
        {
            return query.Skip((pageIndex - 1) * pageSize).Take(pageSize);
        }
    }
}


