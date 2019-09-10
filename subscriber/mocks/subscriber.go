// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import mock "github.com/stretchr/testify/mock"

import sync "sync"

// Subscriber is an autogenerated mock type for the Subscriber type
type Subscriber struct {
	mock.Mock
}

// SubscribeNotification provides a mock function with given fields: ctx, wg, ch
func (_m *Subscriber) SubscribeNotification(ctx context.Context, wg *sync.WaitGroup, ch chan<- interface{}) error {
	ret := _m.Called(ctx, wg, ch)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *sync.WaitGroup, chan<- interface{}) error); ok {
		r0 = rf(ctx, wg, ch)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
