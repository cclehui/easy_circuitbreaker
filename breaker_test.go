package easy_circuitbreaker

import (
	"testing"
	"time"
)

func TestFunctionBase(t *testing.T) {

	breaker := NewRateBreaker(0.9, 100)
	breaker.SetLogger(DefaultLogger{})

	startTime := time.Now()

	for {

		for i := 1; i <= 100; i++ {
			breaker.Success()
		}
		time.Sleep(time.Millisecond * 300)

		for i := 1; i <= 900; i++ {
			breaker.Fail()
		}

		if time.Now().Sub(startTime) > time.Second*1 {
			break
		}
	}

	//select {}

	if breaker.Ready() {
		t.Fail()
	}

	time.Sleep(time.Millisecond * 500)

	if !breaker.Ready() {
		t.Fail()
	}

}
