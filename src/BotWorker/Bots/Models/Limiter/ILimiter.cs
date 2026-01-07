namespace sz84.Bots.Models.Limiter
{
    public interface ILimiter
    {
        /// <summary>
        /// 已使用（限制行为）
        /// </summary>
        Task<bool> HasUsedAsync(long? groupId, long userId, string actionKey);

        /// <summary>
        /// 标记为已使用（限制行为）
        /// </summary>
        Task MarkUsedAsync(long? groupId, long userId, string actionKey);

        /// <summary>
        /// 获取最后使用时间（用于展示）
        /// </summary>
        Task<DateTime?> GetLastUsedAsync(long? groupId, long userId, string actionKey);

        /// <summary>
        /// 如果没用过，就标记使用并返回 true；否则 false
        /// </summary>
        Task<bool> TryUseAsync(long? groupId, long userId, string actionKey);
    }

}
