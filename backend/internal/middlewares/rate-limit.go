package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	Limit  int
	Window time.Duration
	Prefix string
}

type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(
	client *redis.Client,
) *RateLimiter {
	return &RateLimiter{
		client: client,
	}
}

func (r *RateLimiter) Middleware(
	config RateLimitConfig,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		if config.Limit <= 0 {
			c.Next()
			return
		}

		if config.Window <= 0 {
			c.Next()
			return
		}

		ip := c.ClientIP()

		if ip == "" {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"error": "request rejected",
				},
			)

			return
		}

		key := fmt.Sprintf(
			"ratelimit:%s:%s",
			config.Prefix,
			ip,
		)

		ctx := c.Request.Context()

		count, err := r.increment(
			ctx,
			key,
			config.Window,
		)

		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusServiceUnavailable,
				gin.H{
					"error": "rate limiting service unavailable",
				},
			)

			return
		}

		remaining := config.Limit - int(count)

		if remaining < 0 {
			remaining = 0
		}

		c.Header(
			"X-RateLimit-Limit",
			strconv.Itoa(config.Limit),
		)

		c.Header(
			"X-RateLimit-Remaining",
			strconv.Itoa(remaining),
		)

		if count > int64(config.Limit) {
			retryAfter := int(config.Window.Seconds())

			c.Header(
				"Retry-After",
				strconv.Itoa(retryAfter),
			)

			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "too many requests",
				},
			)

			return
		}

		c.Next()
	}
}

func (r *RateLimiter) increment(
	ctx context.Context,
	key string,
	window time.Duration,
) (int64, error) {

	script := redis.NewScript(`
local current = redis.call("INCR", KEYS[1])

if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
`)

	result, err := script.Run(
		ctx,
		r.client,
		[]string{key},
		int(window.Seconds()),
	).Result()

	if err != nil {
		return 0, err
	}

	count, ok := result.(int64)

	if !ok {
		return 0, fmt.Errorf(
			"unexpected Redis rate limit result",
		)
	}

	return count, nil
}
