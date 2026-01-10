package collaboration

import (
	"log"
	"sync"
	"time"
)

// MessageBusImpl 实现通用消息总线
type MessageBusImpl struct {
	mu          sync.RWMutex
	subscribers map[string][]MessageHandler
	messages    chan Message
	stopChan    chan struct{}
}

// NewMessageBus 创建新的消息总线实例
func NewMessageBus() *MessageBusImpl {
	mb := &MessageBusImpl{
		subscribers: make(map[string][]MessageHandler),
		messages:    make(chan Message, 1000),
		stopChan:    make(chan struct{}),
	}
	
	// 启动消息处理协程
	go mb.processMessages()
	
	return mb
}

// processMessages 处理消息
tfunc (mb *MessageBusImpl) processMessages() {
	for {
		select {
		case msg := <-mb.messages:
			mb.handleMessage(msg)
		case <-mb.stopChan:
			return
		}
	}
}

// handleMessage 处理单个消息
tfunc (mb *MessageBusImpl) handleMessage(msg Message) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	
	// 处理定向消息
	if msg.ToRoleID != "" {
		if handlers, ok := mb.subscribers[msg.ToRoleID]; ok {
			for _, handler := range handlers {
				go func(h MessageHandler, m Message) {
					err := h(m)
					if err != nil {
						log.Printf("Error handling message: %v", err)
					}
				}(handler, msg)
			}
		}
		return
	}
	
	// 处理广播消息
	for _, handlers := range mb.subscribers {
		for _, handler := range handlers {
			go func(h MessageHandler, m Message) {
				err := h(m)
				if err != nil {
					log.Printf("Error handling message: %v", err)
				}
			}(handler, msg)
		}
	}
}

// SendMessage 发送消息
tfunc (mb *MessageBusImpl) SendMessage(message Message) error {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	
	mb.messages <- message
	return nil
}

// Subscribe 订阅消息
tfunc (mb *MessageBusImpl) Subscribe(roleID string, handler MessageHandler) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	mb.subscribers[roleID] = append(mb.subscribers[roleID], handler)
	return nil
}

// Unsubscribe 取消订阅
tfunc (mb *MessageBusImpl) Unsubscribe(roleID string, handler MessageHandler) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	if handlers, ok := mb.subscribers[roleID]; ok {
		for i, h := range handlers {
			if h == handler {
				mb.subscribers[roleID] = append(handlers[:i], handlers[i+1:]...)
				break
			}
		}
	}
	return nil
}

// Broadcast 广播消息
tfunc (mb *MessageBusImpl) Broadcast(message Message) error {
	message.ToRoleID = "" // 清空目标角色ID，表示广播
	return mb.SendMessage(message)
}

// SendDelayedMessage 发送延迟消息
tfunc (mb *MessageBusImpl) SendDelayedMessage(message Message, delay time.Duration) error {
	if message.ID == "" {
		message.ID = generateMessageID()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now().Add(delay)
	}
	
	go func() {
		time.Sleep(delay)
		mb.messages <- message
	}()
	
	return nil
}

// Stop 停止消息总线
tfunc (mb *MessageBusImpl) Stop() {
	close(mb.stopChan)
	close(mb.messages)
}

// generateMessageID 生成唯一消息ID
tfunc generateMessageID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
tfunc randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}