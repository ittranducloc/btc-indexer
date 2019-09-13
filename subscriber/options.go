package subscriber

import "time"

type Options struct {
	Url               string
	TimeoutInSecond   time.Duration
	RetryTimeInSecond time.Duration
}

type Option func(*Options)

func Url(url string) Option {
	return func(options *Options) {
		options.Url = url
	}
}

func TimeoutDuration(timeoutInSecond time.Duration) Option {
	return func(options *Options) {
		options.TimeoutInSecond = timeoutInSecond
	}
}

func RetryDuration(retryTimeInSecond time.Duration) Option {
	return func(options *Options) {
		options.RetryTimeInSecond = retryTimeInSecond
	}
}
