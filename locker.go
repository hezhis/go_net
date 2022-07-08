package go_net

import "sync/atomic"

const (
	UnLock = iota + 1
	Lock
)

type Locker struct {
	locker int32
}

func NewLocker() *Locker {
	return &Locker{
		locker: UnLock,
	}
}

func (l *Locker) Lock() {
	atomic.CompareAndSwapInt32(&l.locker, UnLock, Lock)
}

func (l *Locker) Unlock() {
	atomic.CompareAndSwapInt32(&l.locker, Lock, UnLock)
}
