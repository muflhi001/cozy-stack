package lock

import (
	"sync"
	"time"

	"github.com/cozy/cozy-stack/pkg/prefixer"
	"github.com/redis/go-redis/v9"
)

// LockGetter return a lock on a resource matching the given `name`.
type Getter interface {
	// ReadWrite returns the read/write lock for the given name.
	// By convention, the name should be prefixed by the instance domain on which
	// it applies, then a slash and the package name (ie alice.example.net/vfs).
	ReadWrite(db prefixer.Prefixer, name string) ErrorRWLocker

	// LongOperation returns a lock suitable for long operations. It will refresh
	// the lock in redis to avoid its automatic expiration.
	LongOperation(db prefixer.Prefixer, name string) ErrorLocker
}

func New(client redis.UniversalClient) Getter {
	if client == nil {
		return NewInMemory()
	}

	return NewRedisLockGetter(client)
}

// An ErrorLocker is a locker which can fail (returns an error)
type ErrorLocker interface {
	Lock() error
	Unlock()
}

// ErrorRWLocker is the interface for a RWLock as inspired by RWMutex
type ErrorRWLocker interface {
	ErrorLocker
	RLock() error
	RUnlock()
}

type longOperationLocker interface {
	ErrorLocker
	Extend()
}

type longOperation struct {
	lock    longOperationLocker
	mu      sync.Mutex
	tick    *time.Ticker
	timeout time.Duration
}

func (l *longOperation) Lock() error {
	if err := l.lock.Lock(); err != nil {
		return err
	}
	l.tick = time.NewTicker(l.timeout / 3)
	go func() {
		defer l.mu.Unlock()
		for {
			l.mu.Lock()
			if l.tick == nil {
				return
			}
			ch := l.tick.C
			l.mu.Unlock()
			<-ch
			l.mu.Lock()
			if l.tick == nil {
				return
			}
			l.lock.Extend()
			l.mu.Unlock()
		}
	}()
	return nil
}

func (l *longOperation) Unlock() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tick != nil {
		l.tick.Stop()
		l.tick = nil
	}
	l.lock.Unlock()
}
