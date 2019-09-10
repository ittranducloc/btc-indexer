package subscriber

import (
	"context"
	"github.com/darkknightbk52/btc-indexer/common/log"
	"github.com/lightninglabs/gozmq"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type Subscriber interface {
	SubscribeNotification(ctx context.Context, wg *sync.WaitGroup, ch chan<- interface{})
}

const (
	DefaultTimeoutInSecond   = time.Second * 5
	DefaultRetryTimeInSecond = time.Second * 5
)

type subscriber struct {
	opts SubscriberOptions
}

func NewSubscriber(opts ...SubscriberOption) Subscriber {
	options := SubscriberOptions{
		TimeoutInSecond:   DefaultRetryTimeInSecond,
		RetryTimeInSecond: DefaultRetryTimeInSecond,
	}
	for _, o := range opts {
		o(&options)
	}

	return &subscriber{
		opts: options,
	}
}

func (s *subscriber) SubscribeNotification(ctx context.Context, wg *sync.WaitGroup, ch chan<- interface{}) {
	go s.subscribe(ctx, wg, "rawblock", ch)
	go s.subscribe(ctx, wg, "rawtx", ch)
}

func (s *subscriber) subscribe(ctx context.Context, wg *sync.WaitGroup, topic string, ch chan<- interface{}) {
	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		conn, err := gozmq.Subscribe(s.opts.Url, []string{topic}, s.opts.TimeoutInSecond)
		if err == nil {
			s.receive(ctx, topic, ch, conn)
			if ctx.Err() == context.Canceled {
				return
			}
		}

		log.L().Warn("Failed to Subscribe ZMQ message", zap.String("topic", topic), zap.Error(err))
		log.L().Warn("Retrying ... to Subscribe message", zap.String("topic", topic))
		time.Sleep(s.opts.RetryTimeInSecond)
	}
}

func (s *subscriber) receive(ctx context.Context, topic string, ch chan<- interface{}, conn *gozmq.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg, err := conn.Receive()
		if err == nil {
			select {
			case <-ctx.Done():
				err := conn.Close()
				if err != nil {
					log.L().Warn("Failed to Close ZMQ connection", zap.String("topic", topic), zap.Error(err))
				}
				return
			case ch <- msg:
			}
		}

		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			// Ignore Timeout error to prevent spamming the logs
			continue
		}

		log.L().Warn("Failed to Receive ZMQ message", zap.String("topic", topic), zap.Error(err))
		log.L().Warn("Retrying ... to Receive message", zap.String("topic", topic))
		time.Sleep(s.opts.RetryTimeInSecond)
	}
}
