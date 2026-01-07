using sz84.Bots.Entries;
using sz84.Infrastructure.Caching;

namespace sz84.Core.Services
{
    public class GroupInfoService(ICacheRepository<GroupInfo> cache)
    {
        private readonly ICacheRepository<GroupInfo> _cache = cache;

        public async Task RemoveGroupInfoCacheAsync(string groupId)
        {
            await _cache.RemoveAsync(groupId);
        }
    }

}
