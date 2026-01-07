namespace sz84.Infrastructure.Caching
{
    public interface ICacheRepository<T>
    {
        Task<T?> GetAsync(string id);
        Task SetAsync(string id, T value, TimeSpan? expiry = null);
        Task RemoveAsync(string id);
        Task<T?> GetOrSetAsync(string id, Func<Task<T?>> dbFetcher, TimeSpan? expiry = null);
    }
}
