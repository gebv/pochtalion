package pochtalion

import (
	"context"

	"github.com/inpime/sdata"
)

var Sender Pochtalion

type Pochtalion interface {
	Send(from, title, to, body string) chan error
	SendMailling(from, title, body string, emails ...string) chan error
}

type Store interface {
	New(newname string) Store

	// NewFrom(ctx context.Context, fromname, newname string) error

	Add(ctx context.Context, emails ...string) error
	Set(email string, data *sdata.Map) error
	Get(email string) (*sdata.Map, error)
	Del(email string) error

	Search(ctx context.Context, prevkey string, size int) (*sdata.Array, error)
}
