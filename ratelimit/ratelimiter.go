package ratelimit

import (
	"fmt"
	cache "github.com/patrickmn/go-cache"
	"time"
)

func NewDefaultLimiter(id Identity) *Queue {
	q := new(Queue)
	q.Source = id
	q.Ruleset = Policy{
		Window: DefaultWindow,
		Burst:  DefaultBurst,
		Strict: DefaultStrictMode,
	}
	q.Patrons = cache.New(q.Ruleset.Window*time.Second, 5*time.Second)
	q.Known = make(map[interface{}]time.Duration)
	return q
}

func (q *Queue) DebugChannel() chan string {
	q.Patrons.OnEvicted(func(src string, count interface{}) {
		q.debugPrint("ratelimit (expired): ", src, " ", count)
	})
	q.Debug = true
	debugChannel = make(chan string, 10)
	return debugChannel
}

func (q *Queue) Check(from Identity) bool {
	var (
		count int
		err   error
	)
	src := from.UniqueKey()
	if count, err = q.Patrons.IncrementInt(src, 1); err != nil {
		q.debugPrint("ratelimit (new): ", src)
		q.Patrons.Add(src, 1, q.Ruleset.Window*time.Second)
		return false
	}
	// Slow the fuck down
	if count > q.Ruleset.Burst {
		if !q.Ruleset.Strict {
			q.debugPrint("ratelimit (limited): ", count, src)
			return true
		}

		if _, ok := q.Known[src]; !ok {
			q.Known[src] = q.Ruleset.Window
		}
		q.Known[src]++
		q.Patrons.Replace(src, count, q.Known[src]*time.Second)
		q.debugPrint("ratelimit (limited): ", count, " ", src, " ", q.Known[src])
		return true
	}
	return false
}

func (q *Queue) debugPrint(a ...interface{}) {
	if q.Debug {
		debugChannel <- fmt.Sprint(a...)
	}
}
