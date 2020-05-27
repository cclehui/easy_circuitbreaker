package easy_circuitbreaker

import (
	"sync/atomic"
	"time"
)

var breakMap map[string]*EasyBreaker

func init() {
	breakMap = make(map[string]*EasyBreaker)
}

const (
	MQ_EVENT_SUCCESS     = 1
	MQ_EVENT_FAIL        = 2
	MQ_EVENT_KILL_WORKER = 99 //退出当前worker

	DEFAULT_MQ_SIZE    = 10000
	DEFAULT_WORKER_NUM = 1

	LOG_TAG = "easy_breaker"
)

type EasyBreaker struct {
	breaker *circuitbreaker.Breaker

	mq              chan int
	mqCapacity      int
	workerNum       int32
	workerNumRuning int32

	isRuning bool

	logger *log.Log
}

func GetEasyBreaker(name string, breaker *circuitbreaker.Breaker) {

}

func NewBreakerWithOptions(options *circuitbreaker.Options) *EasyBreaker {
	eb := &EasyBreaker{
		breaker:   circuitbreaker.NewBreakerWithOptions(options),
		mq:        make(chan int, DEFAULT_MQ_SIZE),
		workerNum: DEFAULT_WORKER_NUM,
	}

	eb.mqCapacity = DEFAULT_MQ_SIZE

	return eb
}

func (eb *EasyBreaker) Run() {

	if eb.workerNum < 1 {
		panic("workerNum 不能少于一个")
	}

	for i := 0; i < eb.workerNum; i++ {
		go eb.runWorker()
	}

	//监控和控制程序
	go eb.runMonitor()

}

//系统监控和控制程序
func (eb *EasyBreaker) runMonitor() {

	defer func() {
		if err := recover(); err != nil {

			if eb.logger != nil {
				logger.Errorf("%s, breaker runMonitor is shutdown, err:%#v", LOG_TAG, err)
			}
		}

		eb.isRuning = false //关闭系统

		//breaker 不工作后需要重置breaker状态
		eb.breaker.Reset()
	}()

	maxQueuedCount := int(eb.mqCapacity * 0.9)
	warningCount := 0

	for {

		if atomic.LoadInt32(eb.workerNumRuning) < 1 {
			go eb.runWorker() //增大worker数量

		} else if len(eb.mq) > maxQueuedCount {
			//监控队列中未消费数量
			//超过阈值 3 然后增大worker数量
			if warningCount >= 3 {
				eb.runWorker()
				warningCount = 0
			} else {
				warningCount++
			}

		} else {
			//正常状态
			if atomic.LoadInt32(eb.workerNumRuning) > 1 {
				//超过1个 减少worker数量, 达到1个的时候可以采用无锁方式
				eb.mq <- MQ_EVENT_KILL_WORKER
			}
		}

		//控制worker数量

		time.Sleep(time.Millisecond * 300)
	}

}

//处理 事件worker
func (eb *EasyBreaker) runWorker() {

	//运行中的worker数量计算
	atomic.AddInt32(eb.workerNumRuning, 1)

	defer func() {
		if err := recover(); err != nil {

			if eb.logger != nil {
				logger.Errorf("%s, breaker runWorker is shutdown, err:%#v", LOG_TAG, err)
			}
		}

		//breaker 不工作后需要重置breaker状态
		eb.breaker.Reset()

		atomic.AddInt32(eb.workerNumRuning, -1)
	}()

workerLoop:
	for {
		messageData := <-eb.mq

		switch messageData {
		case MQ_EVENT_SUCCESS:
			if eb.workerNumRuning < 2 {
				//无竞争
				eb.breaker.SuccessNoLock()

			} else {
				eb.breaker.Success()
			}

		case MQ_EVENT_FAIL:
			eb.breaker.Fail()
		case MQ_EVENT_KILL_WORKER:
			if eb.logger != nil {
				logger.Errorf("%s, breaker is shutdown by MQ_EVENT_KILL_WORKER")
			}
			break workerLoop
		}
	}

	return
}

//执行函数
func (eb *EasyBreaker) Do(circuit func() error) error {
	if !eb.isRuning() {
		return circuit()
	}

	if !eb.Ready() {
		return circuitbreaker.ErrBreakerOpen
	}

	err = circuit()

	if err != nil {
		eb.Fail()
	} else {

		eb.Success()
	}

	return err
}

//熔断器是否关闭或者可用
func (eb *EasyBreaker) Ready() bool {
	return !eb.isRuning() || eb.breaker.Ready()
}

//采用无锁 success
func (eb *EasyBreaker) Success() {
	if eb.isRuning() {
		eb.mq <- MQ_EVENT_SUCCESS
	}
}

func (eb *EasyBreaker) Fail() {
	if eb.isRuning() {
		eb.mq <- MQ_EVENT_FAIL
	}
}

//系统是否正常运行中
func (eb *EasyBreaker) isRuning() bool {
	return eb.isRuning
}
