using System.Data;
using System.Reflection;
using System.Runtime.CompilerServices;
using Newtonsoft.Json;

namespace BotWorker.Common.Extensions
{
    public static class ExtT
    {
        public static T ShallowCopy<T>(this T source) where T : class
        {
            ArgumentNullException.ThrowIfNull(source);

            var method = source.GetType().GetMethod("MemberwiseClone", BindingFlags.Instance | BindingFlags.NonPublic);
            return method == null ? throw new InvalidOperationException("MemberwiseClone method not found.") : (T)method.Invoke(source, null)!;
        }

        public static T DeepCopy<T>(this T source)
        {
            if (source == null) return default!;

            var serialized = JsonConvert.SerializeObject(source); 
            return JsonConvert.DeserializeObject<T>(serialized)!;
        }
        public static T2? ConvertTo<T2>(this object source)
        {
            var settings = new JsonSerializerSettings
            {
                NullValueHandling = NullValueHandling.Ignore,
                MissingMemberHandling = MissingMemberHandling.Ignore
            };

            return JsonConvert.DeserializeObject<T2>(JsonConvert.SerializeObject(source, settings), settings);
        }


        /// <summary>
        /// 将异步数据流按指定大小分批�?
        /// 可选是否立即物化（避免懒加载带来的副作用）�?
        /// </summary>
        public static async IAsyncEnumerable<IEnumerable<T>> BatchAsync<T>(
            this IAsyncEnumerable<T> source,
            int batchSize,
            bool materialize = false,
            [EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            if (batchSize <= 0) throw new ArgumentOutOfRangeException(nameof(batchSize));

            List<T> batch = new(batchSize);
            await foreach (var item in source.WithCancellation(cancellationToken))
            {
                batch.Add(item);
                if (batch.Count == batchSize)
                {
                    yield return materialize ? batch.ToArray() : batch;
                    batch = new List<T>(batchSize);
                }
            }

            if (batch.Count > 0)
                yield return materialize ? batch.ToArray() : batch;
        }

        /// <summary>
        /// 对异步数据流按批次执行异步操作�?
        /// 支持节流与取消令牌�?
        /// </summary>
        public static async Task ForEachBatchAsync<T>(
            this IAsyncEnumerable<T> source,
            int batchSize,
            Func<IEnumerable<T>, Task> action,
            TimeSpan? delayBetweenBatches = null,
            CancellationToken cancellationToken = default)
        {
            await foreach (var batch in source.BatchAsync(batchSize, materialize: true, cancellationToken))
            {
                await action(batch);
                if (delayBetweenBatches.HasValue)
                    await Task.Delay(delayBetweenBatches.Value, cancellationToken);
            }
        }

        /// <summary>
        /// 返回一个随机布尔值（true �?false）�?
        /// </summary>
        /// <param name="random">Random 实例</param>
        /// <returns>返回 true �?false，概率各�?50%</returns>
        public static bool NextBool(this Random random)
            => random.Next(2) == 0;

        /// <summary>
        /// 截断字符串到指定最大长度，超出部分会被移除�?
        /// </summary>
        /// <param name="str">要截断的字符�?/param>
        /// <param name="maxLength">最大长�?/param>
        /// <returns>原始字符串或被截断后的字符串</returns>
        public static string Truncate(this string str, int maxLength)
            => string.IsNullOrEmpty(str) ? str : (str.Length <= maxLength ? str : str.Substring(0, maxLength));

        /// <summary>
        /// 将序列分批为多个子集合，每批最多包含指定数量的元素�?
        /// </summary>
        /// <typeparam name="T">集合元素类型</typeparam>
        /// <param name="source">原始集合</param>
        /// <param name="size">每批的大小（必须大于0�?/param>
        /// <returns>按批次分组后的序列，每个子集合最多包含指定数量的元素</returns>
        /// <exception cref="ArgumentException">�?size 小于等于 0 时抛�?/exception>
        public static IEnumerable<IEnumerable<T>> Batch<T>(this IEnumerable<T> source, int size)
        {
            if (size <= 0) throw new ArgumentException("Batch size must be greater than zero.", nameof(size));
            using var enumerator = source.GetEnumerator();
            while (enumerator.MoveNext())
            {
                yield return YieldBatchElements(enumerator, size);
            }
        }

        /// <summary>
        /// 辅助方法：从枚举器中提取一个批次的元素�?
        /// </summary>
        /// <typeparam name="T">集合元素类型</typeparam>
        /// <param name="source">元素枚举�?/param>
        /// <param name="size">每批的大�?/param>
        /// <returns>一个批次的元素集合</returns>
        private static IEnumerable<T> YieldBatchElements<T>(IEnumerator<T> source, int size)
        {
            int count = 0;
            do
            {
                yield return source.Current;
                count++;
            } while (count < size && source.MoveNext());
        }

        /// <summary>
        /// 在链式调用中插入一个操作（通常用于调试或副作用），
        /// 执行传入的动作后返回原始对象本身，方便连续调用�?
        /// </summary>
        /// <typeparam name="T">对象类型</typeparam>
        /// <param name="obj">当前对象</param>
        /// <param name="action">对对象执行的操作</param>
        /// <returns>返回原始对象本身，支持链式调�?/returns>
        public static T Tap<T>(this T obj, Action<T> action)
        {
            action(obj);
            return obj;
        }

        /// <summary>
        /// 将值限制在指定的最小值和最大值之间（适用于支持比较的泛型类型）�?
        /// 如果值小于最小值，返回最小值；
        /// 如果值大于最大值，返回最大值；
        /// 否则返回原值�?
        /// </summary>
        /// <typeparam name="T">实现�?IComparable&lt;T&gt; 的类�?/typeparam>
        /// <param name="value">要限制的�?/param>
        /// <param name="min">允许的最小�?/param>
        /// <param name="max">允许的最大�?/param>
        /// <returns>限制后的值，保证在[min, max]范围�?/returns>
        public static T Clamp<T>(this T value, T min, T max) where T : IComparable<T>
        {
            if (value.CompareTo(min) < 0) return min;
            if (value.CompareTo(max) > 0) return max;
            return value;
        }

        /// <summary>
        /// 从集合中随机抽取一个元素，支持排除元素和自定义默认返回值�?
        /// </summary>
        public static T? RandomOne<T>(
            this IEnumerable<T> source,
            IEnumerable<T>? exclude = null,
            T? defaultValue = default,
            bool throwIfEmpty = false)
        {
            ArgumentNullException.ThrowIfNull(source);

            if (exclude != null)
                source = source.Except(exclude);

            if (source is IReadOnlyList<T> list)
            {
                if (list.Count == 0)
                {
                    if (throwIfEmpty) throw new InvalidOperationException("Collection is empty after exclusions.");
                    return defaultValue;
                }
                return list[Random.Shared.Next(list.Count)];
            }

            T? result = defaultValue;
            int count = 0;
            foreach (var item in source)
            {
                count++;
                if (Random.Shared.Next(count) == 0)
                    result = item;
            }

            if (count == 0 && throwIfEmpty)
                throw new InvalidOperationException("Collection is empty after exclusions.");

            return result;
        }

        /// <summary>
        /// 从集合中随机抽取指定数量的不重复元素，支持排除元素�?
        /// </summary>
        public static List<T> RandomMany<T>(
            this IEnumerable<T> source,
            int count,
            IEnumerable<T>? exclude = null)
        {
            if (source == null)
                throw new ArgumentNullException(nameof(source));
            if (count < 0)
                throw new ArgumentOutOfRangeException(nameof(count), "Count must be non-negative.");

            if (exclude != null)
                source = source.Except(exclude);

            if (source is IReadOnlyList<T> list)
            {
                int n = Math.Min(count, list.Count);
                return Shuffle(list).Take(n).ToList();
            }

            var filteredList = source.ToList();
            int m = Math.Min(count, filteredList.Count);
            return Shuffle(filteredList).Take(m).ToList();
        }

        /// <summary>
        /// Fisher–Yates 洗牌算法，返回数组，性能更好，避�?CA1859 警告�?
        /// </summary>
        private static T[] Shuffle<T>(IReadOnlyList<T> source)
        {
            var buffer = source.ToArray();
            for (int i = buffer.Length - 1; i > 0; i--)
            {
                int j = Random.Shared.Next(i + 1);
                (buffer[i], buffer[j]) = (buffer[j], buffer[i]);
            }
            return buffer;
        }

        public static T As<T>(this string res)
        {
            try
            {
                // 使用switch表达式和模式匹配来处理类型转�?
                return typeof(T) switch
                {
                    Type t when t == typeof(bool) => (T)(object)res.AsBool(),
                    Type t when t == typeof(int) => (T)(object)res.AsInt(),
                    Type t when t == typeof(long) => (T)(object)res.AsLong(),
                    Type t when t == typeof(float) => (T)(object)res.AsFloat(),
                    Type t when t == typeof(double) => (T)(object)res.AsDouble(),
                    Type t when t == typeof(decimal) => (T)(object)res.AsDecimal(),
                    Type t when t == typeof(DateTime) => (T)(object)res.AsDateTime(),
                    Type t when t == typeof(string) => (T)(object)res,
                    Type t when t.IsEnum => (T)Enum.Parse(typeof(T), res),
                    _ => (T)Convert.ChangeType(res, typeof(T))
                };
            }
            catch (Exception ex)
            {
                throw new InvalidOperationException($"转换数据时出错：str to type {typeof(T)}.", ex);
            }
        }

        // �?DataTable 转换�?List
        public static List<T> ToList<T>(this DataTable table) where T : new()
        {
            IList<PropertyInfo> properties = [.. typeof(T).GetProperties()];
            List<T> result = [];

            foreach (var row in table.Rows)
            {
                var item = CreateItemFromRow<T>((DataRow)row, properties);
                result.Add(item);
            }

            return result;
        }

        private static T CreateItemFromRow<T>(DataRow row, IList<PropertyInfo> properties) where T : new()
        {
            T item = new();
            foreach (var property in properties)
            {
                if (property.PropertyType == typeof(DayOfWeek))
                {
                    DayOfWeek day = (DayOfWeek)Enum.Parse(typeof(DayOfWeek), row[property.Name].AsString());
                    property.SetValue(item, day, null);
                }
                else
                {
                    if (row[property.Name] == DBNull.Value)
                        property.SetValue(item, null, null);
                    else
                    {
                        if (Nullable.GetUnderlyingType(property.PropertyType) != null)
                        {
                            //nullable
                            object? convertedValue = null;
                            try
                            {
                                convertedValue = Convert.ChangeType(row[property.Name], Nullable.GetUnderlyingType(property.PropertyType) ?? typeof(string));
                            }
                            catch (Exception)
                            {
                            }
                            property.SetValue(item, convertedValue, null);
                        }
                        else
                            property.SetValue(item, row[property.Name], null);
                    }
                }
            }
            return item;
        }
    }
}


