package mrdb

import (
	"context"
	"errors"

	asRedis "github.com/redis/go-redis/v9"
)

type Redis struct {
	conf *redisConf
	*asRedis.Client
}

type redisConf struct {
	Url string
}

type redisOpts func(*redisConf)

func NewClient(url string, opts ...redisOpts) (*Redis, error) {
	if url == "" {
		return nil, errors.New("url is required")
	}

	t := &Redis{
		conf: &redisConf{
			Url: url,
		},
	}

	for _, opt := range opts {
		opt(t.conf)
	}

	var err error
	// 使用 url 解析
	opt, err := asRedis.ParseURL(t.conf.Url)
	if err != nil {
		return nil, err
	}

	// 启用 RESP3 响应
	opt.UnstableResp3 = true

	// 创建 Redis 客户端
	client := asRedis.NewClient(opt)
	// 测试 Redis 连接
	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	t.Client = client
	return t, nil
}
