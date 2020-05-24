# easy_circuitbreaker
易用非阻塞无锁熔断器

#使用场景
学习了开源库的熔断器后，发现都需要加锁，在实际使用的时候担心加锁和计算的并发高了后对本身业务有影响，所以会异步进行熔断器状态计算。
而且精确度可能没要求有那么高，读的时候也可以不加锁直接读就行了，最大限度降低对业务的影响

设计思路主要是基于https://github.com/rubyist/circuitbreaker 进行改造

#benchmark

#基于下面这个熔断器改造
1. https://github.com/rubyist/circuitbreaker master 4afb8473475bafc430776dcaafd1e6d84a8032ad
2. https://github.com/cenkalti/backoff v4.0.2
2. https://github.com/facebookgo/clock master 2020-05-24 600d898af40aa09a7a93ecb9265d87b0504b6f03

#熔断器资料
1. https://zhuanlan.zhihu.com/p/58428026 

#参考学习的熔断器库
1. https://github.com/rubyist/circuitbreaker 
2. https://github.com/sony/gobreaker
3. https://github.com/streadway/handy/tree/master/breaker 
4. https://github.com/afex/hystrix-go 

