package ratelimit

import (
	cache "github.com/patrickmn/go-cache"
	"time"
)

const (
	DefaultWindow     = 10
	DefaultBurst      = 30
	DefaultStrictMode = true
)

var debugChannel chan string

/* Source only need to have one unique key for the rate limiter
   the rest of it is entirely arbitrary, so we can ship around
   information about the origin of the request and still use it
   for the ratelimiter safely. */
type Identity interface {
	UniqueKey() string
}

/* Source implements the Identity interface to keep track of requests
   for rate limiting and routing. */
type IRCSource struct {
	Actual interface{}
}

// Queue implements an Enforcer to create an arbitrary ratelimiter
type Queue struct {
	Source Identity
	// Patrons are the IRC users that we are rate limiting
	Patrons *cache.Cache
	// Ruleset is the actual ratelimitting model
	Ruleset Policy
	Known   map[interface{}]time.Duration
	Debug   bool
}

// Rules defines the mechanics of our ratelimiter
type Policy struct {
	// Window defines the seconds between each post from an IP address
	Window time.Duration
	// Burst is used differently based on Strict mode
	Burst int

	/* Strict bool determines how our limiter functions
	* Strict true
		* First query triggers addition to ratelimit list
		* Continued queries while ratelimited will increment Burst counter
		* Surpassing Burst during ratelimit will:
			1) Reset the Window timer, keeping the request limited
			2) Only cut the Burst counter in half, not reset it
	* Strict false
		* Queries will not trigger addition to ratelimit list until Burst is
		exceeded within Window
	*/
	Strict bool
}
