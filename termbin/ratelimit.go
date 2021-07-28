package termbin

import (
	cache "github.com/patrickmn/go-cache"
	"strings"
	"time"
)

var (
	// Users contains all remote addresses currently being ratelimited
	Users *cache.Cache
	// Rate is the amount of seconds between each post from an IP address
	Rate time.Duration = 30
)

func init() {
	Users = cache.New(Rate*time.Second, 30*time.Second)
}

func isThrottled(addr string) bool {
	addr = strings.Split(addr, ":")[0]
	if _, ok := Users.Get(addr); !ok {
		Users.Set(addr, 1, 0)
		return false
	} else {
		return true
	}
}
