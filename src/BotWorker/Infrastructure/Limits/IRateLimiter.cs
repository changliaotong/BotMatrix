namespace sz84.Infrastructure.Limits
{
    // IRateLimiter.cs
    public interface IRateLimiter
    {
        /// <summary>
        /// 检查并记录访问，返回是否允许访问
        /// </summary>
        /// <param name="key">限流唯一标识</param>
        /// <param name="maxCount">单位时间内最大次数</param>
        /// <param name="period">时间窗口</param>
        bool CheckLimit(string key, int maxCount, TimeSpan period);
    }

}
