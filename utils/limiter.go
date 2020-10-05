package utils

import (
	"log"
	"time"
)

type LimitController struct {
	Requests []time.Time
}

const SlidingWindow = 660 * time.Second
const RequestsPerSlidingWindow = 75

func (l *LimitController) clean(now time.Time) {
	var newList []time.Time

	oneHourAgo := now.Add(-time.Hour)
	for _, v := range l.Requests {
		if v.After(oneHourAgo) {
			newList = append(newList, v)
		}
	}

	l.Requests = newList
}

func (l *LimitController) calculateWaitTime(now time.Time) time.Duration {
	reqOldest := now
	slidingWindow := now.Add(-SlidingWindow)

	var reqCount int
	for _, v := range l.Requests {
		if v.After(slidingWindow) {
			if v.Before(reqOldest) {
				reqOldest = v
			}
			reqCount++
		}
	}

	if reqCount > RequestsPerSlidingWindow {
		log.Printf("Hit limit, oldest request: %s", reqOldest.String())
		return now.Sub(reqOldest) + SlidingWindow + 6*time.Second
	} else {
		return time.Duration(0)
	}
}

func (l *LimitController) WaitBeforeRequest() {
	now := time.Now()

	l.clean(now)

	waitTime := l.calculateWaitTime(now)
	if waitTime > 0 {
		log.Printf("Waiting for %s to avoid running into rate limit", waitTime.Round(time.Second).String())
		time.Sleep(waitTime)
	}

	l.Requests = append(l.Requests, time.Now())
}
