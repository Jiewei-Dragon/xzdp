package Order

import (
	"sync"
)

// import "sync"

// 通过 NewUserLock 函数和全局变量 UserLockMap 实现
// 确保整个程序中只存在一个 UserLock 实例。所有对用户锁的操作都通过这个唯一实例进行，避免了多实例导致的锁管理混乱。

// 结构体（struct）仅用于定义数据结构（字段），而方法（函数）是通过 “接收器” 与结构体关联的
type UserLock struct {
	//并发安全的键值对存储结构（类似 map）
	locks sync.Map
}

func NewUserLock() *UserLock {
	return &UserLock{}
}

var UserLockMap = NewUserLock()

// // (ul *UserLock) 是方法接收器（Method Receiver），用于将这个函数绑定到 UserLock 类型上，使其成为 UserLock 的成员方法。
func (ul *UserLock) Lock(userID int) {
	//基础互斥锁
	var lock *sync.Mutex
	value, ok := ul.locks.Load(userID)
	if ok {
		lock = value.(*sync.Mutex) //类型断言，用于将一个接口类型的值转换为具体的类型
	} else {
		lock = &sync.Mutex{} //没有的话，就需要创建一个新锁
		ul.locks.Store(userID, lock)
	}
	lock.Lock()
}

func (ul *UserLock) Unlock(userID int) {
	value, ok := ul.locks.Load(userID)
	if ok {
		lock := value.(*sync.Mutex)
		lock.Unlock()
	}
}
