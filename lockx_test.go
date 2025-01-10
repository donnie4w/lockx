// Copyright (c) 2023, donnie <donnie4w@gmail.com>
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// https://github.com/donnie4w/raftx
// https://tlnet.top/wiki/raftx
// github.com/donnie4w/lockx

package lockx

import (
	"fmt"
	"log"
	"testing"
	"time"
)

var mutex1 *Mutex
var mutex2 *Mutex
var mutex3 *Mutex

func init() {
	mutex1 = NewMutex(":20001", []string{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003"})
	mutex2 = NewMutex(":20002", []string{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003"})
	mutex3 = NewMutex(":20003", []string{"127.0.0.1:20001", "127.0.0.1:20002", "127.0.0.1:20003"})
	fmt.Println("wait for rafxt init...")
	time.Sleep(3 * time.Second) //模拟服务启动，等待集群完成初始化
}

func Test_lock(t *testing.T) {
	go lock1(1)
	go lock2(2)
	go lock3(3)
	select {}
}

func lock1(i int) {
	log.Printf("mutex1 lock%d lock.....\n", i)
	mutex1.Lock("test", 10)
	log.Printf("mutex1 lock%d get lock successful\n", i)
	time.Sleep(2 * time.Second)
	mutex1.Unlock("test")
	log.Printf("mutex1 lock%d unlock\n", i)
}

func lock2(i int) {
	log.Printf("mutex2 lock%d lock.....\n", i)
	mutex2.Lock("test", 10)
	log.Printf("mutex2 lock%d get lock successful\n", i)
	//time.Sleep(2 * time.Second)
	//mutex2.Unlock("test")
	//log.Printf("mutex2 lock%d unlock\n", i)
}

func lock3(i int) {
	log.Printf("mutex3 lock%d lock.....\n", i)
	mutex3.Lock("test", 10)
	log.Printf("mutex3 lock%d get lock successful\n", i)
	time.Sleep(2 * time.Second)
	mutex3.Unlock("test")
	log.Printf("mutex3 lock%d unlock\n", i)
}

func Test_multi_lock(t *testing.T) {
	for i := 1; i < 1<<15; i++ {
		go lock1(i)
	}
	for i := 1; i < 1<<15; i++ {
		go lock2(i)
	}
	for i := 1; i < 1<<15; i++ {
		go lock3(i)
	}
	select {}
}

func Test_tryLock(t *testing.T) {
	go trylock1(1)
	go trylock2(2)
	go trylock3(3)
	select {}
}

func trylock1(i int) {
	log.Printf("mutex1 tryLock%d lock.....\n", i)
	if mutex1.TryLock("test", 10) {
		log.Printf("mutex1 tryLock%d get lock successful\n", i)
		time.Sleep(2 * time.Second)
		mutex1.Unlock("test")
		log.Printf("mutex1 lock%d unlock\n", i)
	} else {
		log.Printf("mutex1 tryLock%d failed\n", i)
	}
}

func trylock2(i int) {
	log.Printf("mutex2 tryLock%d lock.....\n", i)
	if mutex2.TryLock("test", 10) {
		log.Printf("mutex2 tryLock%d get lock successful\n", i)
	} else {
		log.Printf("mutex2 tryLock%d failed\n", i)
	}
}

func trylock3(i int) {
	log.Printf("mutex3 tryLock%d lock.....\n", i)
	if mutex3.TryLock("test", 10) {
		log.Printf("mutex3 tryLock%d get lock successful\n", i)
		time.Sleep(2 * time.Second)
		mutex3.Unlock("test")
		log.Printf("mutex3 lock%d unlock\n", i)
	} else {
		log.Printf("mutex3 tryLock%d failed\n", i)
	}
}
