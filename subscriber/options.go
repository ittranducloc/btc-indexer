package subscriber

import "time"

type SubscriberOptions struct {
	Url               string
	TimeoutInSecond   time.Duration
	RetryTimeInSecond time.Duration
}

type SubscriberOption func(*SubscriberOptions)

func BtcUrl(url string) SubscriberOption {
	return func(options *SubscriberOptions) {
		options.Url = url
	}
}

func TimeoutDuration(timeoutInSecond time.Duration) SubscriberOption {
	return func(options *SubscriberOptions) {
		options.TimeoutInSecond = timeoutInSecond
	}
}

func RetryDuration(retryTimeInSecond time.Duration) SubscriberOption {
	return func(options *SubscriberOptions) {
		options.RetryTimeInSecond = retryTimeInSecond
	}
}
