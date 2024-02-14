package main

import (
	"fmt"
	"math"
	"net/http"

	"golang.org/x/time/rate"
)

func limit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Limit(*args.LimitRPS), *args.LimitBurst)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			delay := limiter.Reserve().Delay().Seconds()
			w.Header().Add("Retry-After", fmt.Sprintf("%.0f", math.Ceil(delay)))
			http.Error(
				w,
				http.StatusText(http.StatusTooManyRequests),
				http.StatusTooManyRequests,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
