# Lockx 分布式锁应用库 

## 简介
Lockx 是一个基于 raftx 共识协议实现的分布式锁应用库，它为分布式系统中的资源访问提供了同步机制。通过确保任意时刻只有一个客户端能获取到锁并操作共享资源，Lockx 有助于保证数据一致性和任务执行的协调性。

## 特点
- **高效**：基于内存的操作使得获取和释放锁的过程更快。
- **海量锁创建能力**：支持同时创建成千上万乃至数百万个分布式锁。
- **极低的资源占用**：无论是CPU还是内存资源，Lockx 的占用都极少。
- **无自旋阻塞策略**：避免了不必要的CPU资源消耗。
- **抢占式获取锁**：允许快速响应锁的状态变化。
- **TTL（存活时间）支持**：防止因集群节点宕机造成的死锁问题。

## 使用场景
适用于需要高并发控制、数据一致性保障的任务，如库存扣减、定时任务调度、缓存更新、秒杀活动和文件上传等。

## 实现方式
Lockx 主要依赖于 raftx 提供的易失性数据API，利用其强一致性和键值增删改触发事件特性来实现高效的分布式锁逻辑。

## 快速开始
### 安装
```bash
go get github.com/donnie4w/lockx
```

### 创建分布式锁管理器
```go
import "github.com/donnie4w/lockx"

// 创建分布式锁管理器
mutex1 := lockx.NewMutex(":20001", []string{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003"})
```

### 获取与释放锁
```go
// 获取锁
mutex1.Lock("resourceName", ttlInSeconds)

// 尝试获取锁，若失败则立即返回false
success := mutex1.TryLock("resourceName", ttlInSeconds)

// 释放锁
mutex1.Unlock("resourceName")
```

## 示例代码
下面是一个简单的测试例子，展示了如何在三个不同的节点之间竞争同一个资源的分布式锁。

```go
func Test_lock(t *testing.T) {
	go func() { mutex1.Lock("test", 10); defer mutex1.Unlock("test") }()
	go func() { mutex2.Lock("test", 10); defer mutex2.Unlock("test") }()
	go func() { mutex3.Lock("test", 10); defer mutex3.Unlock("test") }()
	select {}
}
```

## 大规模分布式锁测试
对于需要创建大量分布式锁的情况，可以使用以下代码进行大规模测试：

```go
func Test_multi_lock(t *testing.T) {
	for i := 0; i < 1<<15; i++ { // 每个节点创建32768个并发任务
		go func(id int) {
			mutex1.Lock(fmt.Sprintf("lock%d", id), 10)
			defer mutex1.Unlock(fmt.Sprintf("lock%d", id))
		}(i)
	}
	select {}
}
```

Lockx 利用了 raftx 的高效特性和易失性数据存储能力，提供了一种简洁而强大的分布式锁解决方案，非常适合高并发环境下的应用。开发者可以根据自身需求调整或扩展 Lockx 的功能以适应不同的应用场景。
