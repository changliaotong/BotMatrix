package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"botworker/internal/config"
	"botworker/internal/onebot"
	"botworker/internal/redis"

	"BotMatrix/common/log"

	goredis "github.com/redis/go-redis/v9"
)

// RedisStreamServer 处理 Redis Stream 消息
type RedisStreamServer struct {
	config          *config.StreamConfig
	redisClient     *redis.Client
	messageHandlers []onebot.EventHandler
	noticeHandlers  []onebot.EventHandler
	requestHandlers []onebot.EventHandler
	eventHandlers   map[string][]onebot.EventHandler
	apiHandlers     map[string]onebot.RequestHandler
	middlewares     []MiddlewareFunc
	closeChan       chan struct{}
	wg              sync.WaitGroup
}

// NewRedisStreamServer 创建新的 Redis Stream 服务器
func NewRedisStreamServer(cfg *config.StreamConfig, redisClient *redis.Client) *RedisStreamServer {
	if cfg == nil {
		cfg = &config.StreamConfig{
			BatchSize: 10,
			BlockTime: 2 * time.Second,
		}
	}
	return &RedisStreamServer{
		config:          cfg,
		redisClient:     redisClient,
		eventHandlers:   make(map[string][]onebot.EventHandler),
		apiHandlers:     make(map[string]onebot.RequestHandler),
		closeChan:       make(chan struct{}),
		messageHandlers: []onebot.EventHandler{},
		noticeHandlers:  []onebot.EventHandler{},
		requestHandlers: []onebot.EventHandler{},
		middlewares:     []MiddlewareFunc{},
	}
}

// Use 注册中间件
func (s *RedisStreamServer) Use(m ...MiddlewareFunc) {
	s.middlewares = append(s.middlewares, m...)
}

func (s *RedisStreamServer) OnMessage(fn onebot.EventHandler) {
	s.messageHandlers = append(s.messageHandlers, fn)
}

func (s *RedisStreamServer) OnNotice(fn onebot.EventHandler) {
	s.noticeHandlers = append(s.noticeHandlers, fn)
}

func (s *RedisStreamServer) OnRequest(fn onebot.EventHandler) {
	s.requestHandlers = append(s.requestHandlers, fn)
}

func (s *RedisStreamServer) OnEvent(eventName string, fn onebot.EventHandler) {
	s.eventHandlers[eventName] = append(s.eventHandlers[eventName], fn)
}

func (s *RedisStreamServer) HandleAPI(action string, fn onebot.RequestHandler) {
	s.apiHandlers[action] = fn
}

// Run 启动消费者
func (s *RedisStreamServer) Run() error {
	if s.redisClient == nil {
		return fmt.Errorf("redis client is nil")
	}
	if len(s.config.Streams) == 0 {
		log.Warn("[RedisStream] No streams configured to consume")
		return nil
	}

	ctx := context.Background()

	// 确保消费者组存在
	for _, stream := range s.config.Streams {
		err := s.redisClient.XGroupCreateMkStream(ctx, stream, s.config.Group, "0").Err()
		if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
			log.Errorf("[RedisStream] Failed to create group for stream %s: %v", stream, err)
		}
	}

	// 启动消费者协程
	s.wg.Add(1)
	go s.consumeLoop()

	log.Info(fmt.Sprintf("[RedisStream] Started consuming streams: %v as group: %s", s.config.Streams, s.config.Group))
	return nil
}

func (s *RedisStreamServer) consumeLoop() {
	defer s.wg.Done()

	consumer := s.config.Consumer
	if consumer == "" {
		consumer = fmt.Sprintf("worker-%d", time.Now().UnixNano())
	}

	streams := s.config.Streams
	// XREADGROUP 需要 Streams 参数格式为 [stream1, stream2, ..., ID1, ID2, ...]
	// 使用 ">" 表示读取未被消费的消息
	readArgs := make([]string, len(streams)*2)
	for i, stream := range streams {
		readArgs[i] = stream
		readArgs[i+len(streams)] = ">"
	}

	for {
		select {
		case <-s.closeChan:
			return
		default:
			// 读取消息
			entries, err := s.redisClient.XReadGroup(context.Background(), &goredis.XReadGroupArgs{
				Group:    s.config.Group,
				Consumer: consumer,
				Streams:  readArgs,
				Count:    s.config.BatchSize,
				Block:    s.config.BlockTime,
				NoAck:    false, // 需要手动确认
			}).Result()

			if err != nil {
				if err != goredis.Nil {
					log.Errorf("[RedisStream] Read error: %v", err)
					time.Sleep(1 * time.Second) // 出错后稍作等待
				}
				continue
			}

			for _, streamData := range entries {
				for _, msg := range streamData.Messages {
					s.processStreamMessage(streamData.Stream, msg)
				}
			}
		}
	}
}

func (s *RedisStreamServer) processStreamMessage(stream string, msg goredis.XMessage) {
	// 假设消息体在 "event" 字段中，或者是整个 Values 映射
	// 这里我们尝试从 "data" 或 "payload" 或 "event" 字段解析 JSON
	// 如果没有特定字段，尝试将整个 Values 转为 JSON

	var eventJSON []byte
	var err error

	if val, ok := msg.Values["data"].(string); ok {
		eventJSON = []byte(val)
	} else if val, ok := msg.Values["payload"].(string); ok {
		eventJSON = []byte(val)
	} else if val, ok := msg.Values["event"].(string); ok {
		eventJSON = []byte(val)
	} else {
		// 尝试序列化整个 map
		eventJSON, err = json.Marshal(msg.Values)
		if err != nil {
			log.Errorf("[RedisStream] Failed to marshal message values: %v", err)
			return // 无法处理，不 Ack，稍后重试或进入死信
		}
	}

	var event onebot.Event
	if err := json.Unmarshal(eventJSON, &event); err != nil {
		log.Errorf("[RedisStream] Failed to unmarshal event: %v", err)
		// 格式错误的消息应该 Ack 掉以免死循环，或者移入死信队列
		s.redisClient.XAck(context.Background(), stream, s.config.Group, msg.ID)
		return
	}

	// 补充来源信息，方便调试
	if event.Raw == nil {
		event.Raw = msg.Values
	}

	// 处理 ID 生成 (这部分逻辑复用了 http/ws 的处理)
	processEventIDs(&event)

	// 构建并执行中间件链
	finalHandler := func(e *onebot.Event) error {
		s.dispatchEvent(*e)
		return nil
	}

	wrappedHandler := ChainMiddleware(s.middlewares, finalHandler)

	// 执行处理
	if err := wrappedHandler(&event); err != nil {
		log.Errorf("[RedisStream] Error processing event %s: %v", msg.ID, err)
		// 处理失败是否 Ack? 取决于策略。这里假设业务错误也视为已处理，避免卡死。
		// 如果是系统级错误（如 DB 连接失败），可能需要重试（不 Ack）。
		// 简单起见，我们目前都 Ack。
	}

	// 确认消息
	s.redisClient.XAck(context.Background(), stream, s.config.Group, msg.ID)
}

func (s *RedisStreamServer) dispatchEvent(event onebot.Event) {
	// 复用 WebSocket/HTTP 的分发逻辑
	// 注意：这里是单向通知，没有 Response 机制

	switch event.PostType {
	case "message":
		for _, handler := range s.messageHandlers {
			if err := handler(&event); err != nil {
				log.Errorf("[RedisStream] Message handler error: %v", err)
			}
		}
	case "notice":
		for _, handler := range s.noticeHandlers {
			if err := handler(&event); err != nil {
				log.Errorf("[RedisStream] Notice handler error: %v", err)
			}
		}
	case "request":
		for _, handler := range s.requestHandlers {
			if err := handler(&event); err != nil {
				log.Errorf("[RedisStream] Request handler error: %v", err)
			}
		}
	}

	if handlers, ok := s.eventHandlers[event.EventName]; ok {
		for _, handler := range handlers {
			if err := handler(&event); err != nil {
				log.Errorf("[RedisStream] Event handler error: %v", err)
			}
		}
	}
}

func (s *RedisStreamServer) Stop() {
	close(s.closeChan)
	s.wg.Wait()
}
