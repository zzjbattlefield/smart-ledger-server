package ai

import (
	"context"

	"golang.org/x/time/rate"
)

// RPMLimiter RPM限流器
// 基于令牌桶算法实现每分钟请求数限制
type RPMLimiter struct {
	limiter *rate.Limiter
	rpm     int
}

// NewRPMLimiter 创建RPM限流器
// rpm: 每分钟允许的请求数
func NewRPMLimiter(rpm int) *RPMLimiter {
	// 计算每秒令牌生成速率: rpm / 60 = 每秒请求数
	r := rate.Limit(float64(rpm) / 60.0)

	// 桶大小设置为 rpm/10，允许小范围突发
	// 但不超过 rpm/6（即10秒的量），最小为1
	burstSize := rpm / 10
	if burstSize < 1 {
		burstSize = 1
	}
	if burstSize > rpm/6 {
		burstSize = rpm / 6
	}

	return &RPMLimiter{
		limiter: rate.NewLimiter(r, burstSize),
		rpm:     rpm,
	}
}

// Wait 等待获取令牌
// 阻塞直到获得令牌或context取消
func (l *RPMLimiter) Wait(ctx context.Context) error {
	return l.limiter.Wait(ctx)
}

// Allow 非阻塞检查是否允许请求
func (l *RPMLimiter) Allow() bool {
	return l.limiter.Allow()
}

// RPM 获取配置的RPM值
func (l *RPMLimiter) RPM() int {
	return l.rpm
}
