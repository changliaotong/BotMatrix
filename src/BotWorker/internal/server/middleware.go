package server

import (
	"botworker/internal/onebot"
)

// HandlerFunc 定义事件处理函数签名
// 注意：这与 onebot.EventHandler (func(event *onebot.Event) error) 签名匹配
type HandlerFunc func(event *onebot.Event) error

// MiddlewareFunc 定义中间件函数签名
// 它接收一个处理器，并返回一个新的处理器
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// ChainMiddleware 将多个中间件链接在一起
// 顺序：ChainMiddleware(m1, m2, m3) -> m1(m2(m3(finalHandler)))
func ChainMiddleware(middlewares []MiddlewareFunc, finalHandler HandlerFunc) HandlerFunc {
	// 如果没有中间件，直接返回最终处理器
	if len(middlewares) == 0 {
		return finalHandler
	}

	// 倒序包装，这样第一个中间件最先执行
	h := finalHandler
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
