package btc_indexer

import (
	"github.com/micro/go-micro/errors"
	"github.com/micro/go-micro/server"
)

func RPCError(method string, err error) error {
	id := server.DefaultOptions().Name + "." + method
	switch err {
	default:
		return errors.InternalServerError(id, err.Error())
	}
}
