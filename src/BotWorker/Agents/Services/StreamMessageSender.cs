using System.Text;

namespace BotWorker.Agents.Services
{
    //流式消息通过机器人多次发送 测试过程
    public class StreamMessageSender
    {
        private readonly int _maxMessageLength;
        private readonly Func<string, CancellationToken, Task> _sendMessageAsync;
        private readonly HashSet<char> _sentenceEndChars;
        private readonly StringBuilder _buffer = new();

        public StreamMessageSender(
            Func<string, CancellationToken, Task> sendMessageAsync,
            int maxMessageLength = 300,
            IEnumerable<char>? sentenceEndChars = null)
        {
            _sendMessageAsync = sendMessageAsync ?? throw new ArgumentNullException(nameof(sendMessageAsync));
            _maxMessageLength = maxMessageLength > 0 ? maxMessageLength : throw new ArgumentOutOfRangeException(nameof(maxMessageLength));
            _sentenceEndChars = sentenceEndChars != null
                ? [.. sentenceEndChars]
                : ['.', '。', '!', '！', '?', '？', '\n', '\r'];
        }

        private bool IsSentenceEnd(char c) => _sentenceEndChars.Contains(c);

        public async Task AppendAsync(string chunk, bool isStreaming, CancellationToken cancellationToken = default)
        {
            cancellationToken.ThrowIfCancellationRequested();

            _buffer.Append(chunk);

            int lastSplitIndex = -1;
            for (int i = _buffer.Length - 1; i >= 0; i--)
            {
                if (IsSentenceEnd(_buffer[i]))
                {
                    lastSplitIndex = i + 1; // split after this char
                    break;
                }
            }

            if (lastSplitIndex > 0 && lastSplitIndex >= _maxMessageLength / 2)
            {
                var msg = _buffer.ToString(0, lastSplitIndex);
                _buffer.Remove(0, lastSplitIndex);
                await _sendMessageAsync(msg, cancellationToken);
            }
            else if (_buffer.Length >= _maxMessageLength)
            {
                var msg = _buffer.ToString(0, _maxMessageLength);
                _buffer.Remove(0, _maxMessageLength);
                await _sendMessageAsync(msg, cancellationToken);
            }

            if (!isStreaming && _buffer.Length > 0)
            {
                await _sendMessageAsync(_buffer.ToString(), cancellationToken);
                _buffer.Clear();
                await _sendMessageAsync("[流式生成结束]", cancellationToken);
            }
        }

        public void Clear() => _buffer.Clear();
    }

}
