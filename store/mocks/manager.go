// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/darkknightbk52/btc-indexer/model"

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// AddBlocksData provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *Manager) AddBlocksData(_a0 []*model.Block, _a1 []*model.Tx, _a2 []*model.TxIn, _a3 []*model.TxOut) error {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*model.Block, []*model.Tx, []*model.TxIn, []*model.TxOut) error); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetBlock provides a mock function with given fields: height
func (_m *Manager) GetBlock(height int64) (*model.Block, error) {
	ret := _m.Called(height)

	var r0 *model.Block
	if rf, ok := ret.Get(0).(func(int64) *model.Block); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlocks provides a mock function with given fields: heights
func (_m *Manager) GetBlocks(heights []int64) (map[int64]*model.Block, error) {
	ret := _m.Called(heights)

	var r0 map[int64]*model.Block
	if rf, ok := ret.Get(0).(func([]int64) map[int64]*model.Block); ok {
		r0 = rf(heights)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int64]*model.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]int64) error); ok {
		r1 = rf(heights)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBlocksData provides a mock function with given fields: fromHeight, toHeight, interestedAddresses
func (_m *Manager) GetBlocksData(fromHeight int64, toHeight int64, interestedAddresses []string) (map[int64]*model.Block, map[int64][]*model.TxIn, map[int64][]*model.TxOut, error) {
	ret := _m.Called(fromHeight, toHeight, interestedAddresses)

	var r0 map[int64]*model.Block
	if rf, ok := ret.Get(0).(func(int64, int64, []string) map[int64]*model.Block); ok {
		r0 = rf(fromHeight, toHeight, interestedAddresses)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[int64]*model.Block)
		}
	}

	var r1 map[int64][]*model.TxIn
	if rf, ok := ret.Get(1).(func(int64, int64, []string) map[int64][]*model.TxIn); ok {
		r1 = rf(fromHeight, toHeight, interestedAddresses)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[int64][]*model.TxIn)
		}
	}

	var r2 map[int64][]*model.TxOut
	if rf, ok := ret.Get(2).(func(int64, int64, []string) map[int64][]*model.TxOut); ok {
		r2 = rf(fromHeight, toHeight, interestedAddresses)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(map[int64][]*model.TxOut)
		}
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(int64, int64, []string) error); ok {
		r3 = rf(fromHeight, toHeight, interestedAddresses)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// GetLatestBlock provides a mock function with given fields:
func (_m *Manager) GetLatestBlock() (*model.Block, error) {
	ret := _m.Called()

	var r0 *model.Block
	if rf, ok := ret.Get(0).(func() *model.Block); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Reorg provides a mock function with given fields: event
func (_m *Manager) Reorg(event *model.Reorg) error {
	ret := _m.Called(event)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Reorg) error); ok {
		r0 = rf(event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
