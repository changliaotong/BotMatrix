using System.Diagnostics;
using System.Reflection;
using Newtonsoft.Json;
using Newtonsoft.Json.Serialization;

namespace BotWorker.Common.Exts
{
    public static class AdvancedExtensions
    {
        // ---------- 反射 & 属性复制 ----------

        public static void CopyPropertiesFrom<T>(this T target, object source)
        {
            if (source == null) return;
            var targetType = typeof(T);
            var sourceType = source.GetType();

            var props = from t in targetType.GetProperties(BindingFlags.Public | BindingFlags.Instance)
                        join s in sourceType.GetProperties(BindingFlags.Public | BindingFlags.Instance)
                        on t.Name equals s.Name
                        where t.CanWrite && s.CanRead
                        select new { Target = t, Source = s };

            foreach (var pair in props)
            {
                var value = pair.Source.GetValue(source);
                pair.Target.SetValue(target, value);
            }
        }

        // ---------- IEnumerable 扩展 ----------

        // 安全 ForEach（带索引）
        public static void ForEach<T>(this IEnumerable<T> source, Action<T, int> action)
        {
            int index = 0;
            foreach (var item in source)
            {
                action(item, index++);
            }
        }

        // 去重（基于指定属性）
        public static IEnumerable<T> DistinctBy<T, TKey>(this IEnumerable<T> source, Func<T, TKey> keySelector)
        {
            return source.GroupBy(keySelector).Select(g => g.First());
        }

        // 分组批处理
        public static IEnumerable<List<T>> ChunkBy<T>(this IEnumerable<T> source, int chunkSize)
        {
            var chunk = new List<T>(chunkSize);
            foreach (var item in source)
            {
                chunk.Add(item);
                if (chunk.Count == chunkSize)
                {
                    yield return chunk;
                    chunk = new List<T>(chunkSize);
                }
            }
            if (chunk.Count > 0)
                yield return chunk;
        }

        // ---------- 异步扩展 ----------

        // 忽略异常的异步执行（用于日志上传、无关紧要的后台任务）
        public static async Task FireAndForget(this Task task, Action<Exception>? onError = null)
        {
            try
            {
                await task.ConfigureAwait(false);
            }
            catch (Exception ex)
            {
                onError?.Invoke(ex);
            }
        }

        // 并行执行所有任务，带最大并发限制
        public static async Task ParallelForEachAsync<T>(
            this IEnumerable<T> source,
            Func<T, Task> taskSelector,
            int maxDegreeOfParallelism = 4)
        {
            using var semaphore = new SemaphoreSlim(maxDegreeOfParallelism);
            var tasks = source.Select(async item =>
            {
                await semaphore.WaitAsync();
                try
                {
                    await taskSelector(item);
                }
                finally
                {
                    semaphore.Release();
                }
            });

            await Task.WhenAll(tasks);
        }

        // ---------- JSON 扩展 ----------

        public static string ToJson<T>(this T obj, bool indented = false)
        {
            var settings = new JsonSerializerSettings
            {
                ContractResolver = new CamelCasePropertyNamesContractResolver(),
                Formatting = indented ? Formatting.Indented : Formatting.None
            };

            return JsonConvert.SerializeObject(obj, settings);
        }

        public static T? FromJson<T>(this string json)
        {
            if (string.IsNullOrWhiteSpace(json))
                return default;

            try
            {
                return JsonConvert.DeserializeObject<T>(json);
            }
            catch (JsonException ex)
            {
                // Handle the JSON exception, for example, log the error
                Console.WriteLine($"Error deserializing JSON: {ex.Message}");
                return default;
            }
        }

        // ---------- 通用对象扩展 ----------

        // 判断对象是否在一个集合中（类似 SQL 的 IN）
        public static bool In<T>(this T item, params T[] items)
        {
            if (typeof(T) == typeof(string))
            {
                var strItem = item as string;
                return items.Cast<string>().Any(i => string.Equals(i, strItem, StringComparison.OrdinalIgnoreCase));
            }
            else
            {
                return items.Contains(item);
            }
        }

        // 判断对象是否为空（null 或 空字符串/集合）
        public static bool IsNullOrEmpty<T>(this T? obj)
        {
            if (obj == null) return true;
            if (obj is string str) return string.IsNullOrWhiteSpace(str);
            if (obj is IEnumerable<object> list) return !list.Any();
            return false;
        }

        // ---------- 时间相关 ----------

        public static string ToFriendlyTime(this DateTime dt)
        {
            var ts = DateTime.Now - dt;
            if (ts.TotalSeconds < 60) return "刚刚";
            if (ts.TotalMinutes < 60) return $"{(int)ts.TotalMinutes} 分钟前";
            if (ts.TotalHours < 24) return $"{(int)ts.TotalHours} 小时前";
            if (ts.TotalDays < 7) return $"{(int)ts.TotalDays} 天前";
            return dt.ToString("yyyy-MM-dd HH:mm");
        }

        // ---------- 性能计时器 ----------

        public static TResult TimeIt<TResult>(this Func<TResult> func, Action<TimeSpan>? onComplete = null)
        {
            var sw = Stopwatch.StartNew();
            var result = func();
            sw.Stop();
            onComplete?.Invoke(sw.Elapsed);
            return result;
        }

        public static async Task<TResult> TimeItAsync<TResult>(this Func<Task<TResult>> func, Action<TimeSpan>? onComplete = null)
        {
            var sw = Stopwatch.StartNew();
            var result = await func();
            sw.Stop();
            onComplete?.Invoke(sw.Elapsed);
            return result;
        }
    }
}