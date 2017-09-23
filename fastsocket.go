package fastsocket

import (
	"github.com/mailru/easygo/netpoll"
)

const DefaultPoolSize = 256 * 1024

var poller netpoll.Poller
var workerPool *Pool

func init() {
	var err error
	poller, err = netpoll.New(nil)
	if err != nil {
		panic(err)
	}

	SetWorkerPool(NewPool(DefaultPoolSize, 1, 1))
}

func SetWorkerPool(p *Pool) {
	workerPool = p
}
