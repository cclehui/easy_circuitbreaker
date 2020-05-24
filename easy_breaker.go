package easy_circuitbreaker

import "sync/atomic"

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
	workerNum       int32
	workerNumRuning int32

	isRuning bool

	logger *log.Log
}

func GetEasyBreaker(name string, breaker *circuitbreaker.Breaker) {

}

func NewBreakerWithOptions(options *circuitbreaker.Options) *EasyBreaker {
	return &EasyBreaker{
		breaker:   circuitbreaker.NewBreakerWithOptions(options),
		mq:        make(chan int, DEFAULT_MQ_SIZE),
		workerNum: DEFAULT_WORKER_NUM,
	}
}

func (eb *EasyBreaker) Run() {

	if eb.workerNum < 1 {
		panic("workerNum 不能少于一个")
	}

	for i := 0; i < eb.workerNum; i++ {
		go runWorker()
	}

	//监控和控制程序
	go runMonitor()

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

	//监控队列中未消费数量

	//控制worker数量

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
			eb.breaker.SuccessNoLock()
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
