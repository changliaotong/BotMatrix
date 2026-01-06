package server

import (
	"BotMatrix/common/log"
	"botworker/internal/onebot"
	"time"
)

// LoggingMiddleware 是一个示例中间件，用于记录所有进入系统的事件
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(event *onebot.Event) error {
		start := time.Now()
		
		// 调用下一个处理器
		err := next(event)
		
		duration := time.Since(start)
		
		// 记录简要日志
		// 注意：event.PostType 可能是 message, notice, request, meta_event
		log.Infof("[Middleware] Event Processed | Type: %s | Platform: %s | Duration: %v | Error: %v", 
			event.PostType, event.Platform, duration, err)
			
		return err
	}
}

// RecoveryMiddleware 捕获 Panic 防止崩溃
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
	return func(event *onebot.Event) error {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("[Middleware] Panic recovered in event handler: %v", r)
			}
		}()
		return next(event)
	}
}
