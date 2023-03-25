package logic

import (
	"context"
	"github.com/xh-polaris/sts-rpc/internal/svc"
	"github.com/xh-polaris/sts-rpc/pb"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"sync"
	"time"
)

var (
	cli *redis.Redis
	mu  sync.Mutex
)

const (
	delay      = time.Hour
	delayQueue = "sts:dq:timeOutObjectUrl"
)

// checkSingletonRedis singleton redis pattern
func checkSingletonRedis(redisConf *redis.RedisConf) {
	if cli == nil {
		mu.Lock()
		if cli == nil {
			cli = redis.MustNewRedis(*redisConf)
		}
		mu.Unlock()
	}
}

func DeleteTimeoutObjectLogic(svc *svc.ServiceContext) {
	for {
		// check if the cli is alive
		checkSingletonRedis(&svc.Config.Redis)
		ctx := context.Background()
		cnt, _ := cli.Zcard(delayQueue)
		if cnt == 0 {
			// 查看是否有数据，如果没有进入睡眠，降低cpu和redis的占用
			time.Sleep(time.Millisecond * 300)
			continue
		}
		f, _ := cli.ZrangeWithScores(delayQueue, 0, 0)
		first := f[0]
		// 查看该数据是否过期
		if first.Score > time.Now().Unix() {
			// 队列头部预定时间大于当前时间，进入睡眠，降低cpu和redis的占用
			// 睡眠至delayQueue队头
			time.Sleep(time.Second * time.Duration(first.Score-time.Now().Unix()))
			continue
		}
		suc, _ := cli.Zrem(delayQueue, first.Key)
		if suc == 0 {
			// 该url已经被其他实例获取，无需睡眠，因为还有可能存在数据
			continue
		}
		// 查看是否存在于已经使用的url中
		l := NewDeleteObjectLogic(ctx, svc)
		fakeReq := pb.DeleteObjectReq{
			Path: first.Key,
		}
		// 删除不成功的话只能不再删除, 所以这里没有处理异常
		_, _ = l.DeleteObject(&fakeReq)
	}
}

func addToDelayQueue(svc *svc.ServiceContext, path string) error {
	checkSingletonRedis(&svc.Config.Redis)
	_, err := cli.Zadd(delayQueue, time.Now().Add(delay).Unix(), path)
	if err != nil {
		return err
	}
	return nil
}
