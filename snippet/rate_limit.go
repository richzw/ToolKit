package snippet

import (
	"math"
	"sync"
	"time"
)

/* 1. 计数器是一种比较简单粗暴的限流算法
   - 其思想是在固定时间窗口内对请求进行计数，与阀值进行比较判断是否需要限流，一旦到了时间临界点，将计数器清零
   - 计数器算法存在“时间临界点”缺陷, 计数器算法实现限流的问题是没有办法应对突发流量
*/
type LimitRate struct {
	rate  int           //阀值
	begin time.Time     //计数开始时间
	cycle time.Duration //计数周期
	count int           //收到的请求数
	lock  sync.Mutex
}

func (limit *LimitRate) Allow() bool {
	limit.lock.Lock()
	defer limit.lock.Unlock()

	if limit.count == limit.rate-1 {
		now := time.Now()
		if now.Sub(limit.begin) >= limit.cycle {
			limit.Reset(now)
			return true
		}
		return false
	} else {
		limit.count++
		return true
	}
}

func (limit *LimitRate) Set(rate int, cycle time.Duration) {
	limit.rate = rate
	limit.begin = time.Now()
	limit.cycle = cycle
	limit.count = 0
}

func (limit *LimitRate) Reset(begin time.Time) {
	limit.begin = begin
	limit.count = 0
}

/* 滑动窗口算法
- 将一个大的时间窗口分成多个小窗口，每次大窗口向后滑动一个小窗口，并保证大的窗口内流量不会超出最大值，这种实现比固定窗口的流量曲线更加平滑。
- 滑动窗口算法是固定窗口的一种改进，但从根本上并没有真正解决固定窗口算法的临界突发流量问题
- kratos框架里circuit breaker用循环列表保存timeSlot对象的实现，他们这个实现的好处是不用频繁的创建和销毁timeslot对象
*/
type timeSlot struct {
	timestamp time.Time // 这个timeSlot的时间起点
	count     int       // 落在这个timeSlot内的请求数
}

// 统计整个时间窗口中已经发生的请求次数
func countReq(win []*timeSlot) int {
	var count int
	for _, ts := range win {
		count += ts.count
	}
	return count
}

type SlidingWindowLimiter struct {
	mu           sync.Mutex
	SlotDuration time.Duration // time slot的长度
	WinDuration  time.Duration // sliding window的长度
	numSlots     int           // window内最多有多少个slot
	windows      []*timeSlot
	maxReq       int // 大窗口时间内允许的最大请求数
}

func NewSliding(slotDuration time.Duration, winDuration time.Duration, maxReq int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		SlotDuration: slotDuration,
		WinDuration:  winDuration,
		numSlots:     int(winDuration / slotDuration),
		maxReq:       maxReq,
	}
}

func (l *SlidingWindowLimiter) validate() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	// 已经过期的time slot移出时间窗
	timeoutOffset := -1
	for i, ts := range l.windows {
		if ts.timestamp.Add(l.WinDuration).After(now) {
			break
		}
		timeoutOffset = i
	}
	if timeoutOffset > -1 {
		l.windows = l.windows[timeoutOffset+1:]
	}

	// 判断请求是否超限
	var result bool
	if countReq(l.windows) < l.maxReq {
		result = true
	}

	// 记录这次的请求数
	var lastSlot *timeSlot
	if len(l.windows) > 0 {
		lastSlot = l.windows[len(l.windows)-1]
		if lastSlot.timestamp.Add(l.SlotDuration).Before(now) {
			// 如果当前时间已经超过这个时间插槽的跨度，那么新建一个时间插槽
			lastSlot = &timeSlot{timestamp: now, count: 1}
			l.windows = append(l.windows, lastSlot)
		} else {
			lastSlot.count++
		}
	} else {
		lastSlot = &timeSlot{timestamp: now, count: 1}
		l.windows = append(l.windows, lastSlot)
	}

	return result
}

/* [漏桶算法](https://github.com/kevinyan815/gocookbook/issues/28) 是流量最均匀的限流实现方式
- 有一个木桶，桶的容量是固定的。当有请求到来时先放到木桶中，处理请求的worker以固定的速度从木桶中取出请求进行相应。如果木桶已经满了，直接返回请求频率超限的错误码或者页面。
- 一般用于流量“整形”。例如保护数据库的限流，先把对数据库的访问加入到木桶中，worker再以db能够承受的qps从木桶中取出请求，去访问数据库。
- 木桶流入请求的速率是不固定的，但是流出的速率是恒定的。这样的话能保护系统资源不被打满，但是面对突发流量时会有大量请求失败，不适合电商抢购和微博出现热点事件等场景的限流。
*/

type LeakyBucket struct {
	rate       float64 // 每秒固定流出速率
	capacity   float64 // 桶的容量
	water      float64 // 当前桶中请求量
	lastLeakMs int64   // 桶上次漏水微秒数
	lock       sync.Mutex
}

func (leaky *LeakyBucket) Allow() bool {
	leaky.lock.Lock()
	defer leaky.lock.Unlock()

	now := time.Now().UnixNano() / 1e6
	// 计算剩余水量,两次执行时间中需要漏掉的水
	leakyWater := leaky.water - (float64(now-leaky.lastLeakMs) * leaky.rate / 1000)
	leaky.water = math.Max(0, leakyWater)
	leaky.lastLeakMs = now
	if leaky.water+1 <= leaky.capacity {
		leaky.water++
		return true
	} else {
		return false
	}
}

func (leaky *LeakyBucket) Set(rate, capacity float64) {
	leaky.rate = rate
	leaky.capacity = capacity
	leaky.water = 0
	leaky.lastLeakMs = time.Now().UnixNano() / 1e6
}

/* 令牌桶是反向的"漏桶"
- 以恒定的速度往木桶里加入令牌，木桶满了则不再加入令牌。
- 服务收到请求时尝试从木桶中取出一个令牌，如果能够得到令牌则继续执行后续的业务逻辑。如果没有得到令牌，直接返回访问频率超限的错误码或页面等，不继续执行后续的业务逻辑

- 适合电商抢购或者微博出现热点事件这种场景，因为在限流的同时可以应对一定的突发流量。如果采用漏桶那样的均匀速度处理请求的算法，在发生热点时间的时候，会造成大量的用户无法访问，对用户体验的损害比较大。

- `golang.org/x/time/rate`。该限流器也是基于 Token Bucket(令牌桶) 实现的。
- `uber-go/ratelimit`也是一个很好的选择，与Golang官方限流器不同的是Uber的限流器是通过漏桶算法实现的
*/
type TokenBucket struct {
	rate         int64 //固定的token放入速率, r/s
	capacity     int64 //桶的容量
	tokens       int64 //桶中当前token数量
	lastTokenSec int64 //上次向桶中放令牌的时间的时间戳，单位为秒
	lock         sync.Mutex
}

func (bucket *TokenBucket) Take() bool {
	bucket.lock.Lock()
	defer bucket.lock.Unlock()

	now := time.Now().Unix()
	bucket.tokens = bucket.tokens + (now-bucket.lastTokenSec)*bucket.rate // 先添加令牌
	if bucket.tokens > bucket.capacity {
		bucket.tokens = bucket.capacity
	}
	bucket.lastTokenSec = now
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	} else {
		return false
	}
}

func (bucket *TokenBucket) Init(rate, cap int64) {
	bucket.rate = rate
	bucket.capacity = cap
	bucket.tokens = 0
	bucket.lastTokenSec = time.Now().Unix()
}
