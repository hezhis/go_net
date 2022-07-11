package go_net

import "sync/atomic"

const (
	UnLock = iota + 1
	Lock
)

type Locker struct {
	nLocker int32
}

func NewLocker() *Locker {
	return &Locker{
		nLocker: UnLock,
	}
}

func (l *Locker) Lock() {
	atomic.CompareAndSwapInt32(&l.nLocker, UnLock, Lock)
}

func (l *Locker) Unlock() {
	atomic.CompareAndSwapInt32(&l.nLocker, Lock, UnLock)
}
