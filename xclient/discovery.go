package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refersh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MutiServerDiscovery struct {
	r       *rand.Rand
	mu      sync.Mutex
	servers []string
	index   int //record the select position for robin algorithm
}

func NewMutiServerDiscovery(servers []string) *MutiServerDiscovery {
	d := &MutiServerDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MutiServerDiscovery)(nil)

func (d *MutiServerDiscovery) Refersh() error {
	return nil
}

func (d *MutiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

func (d *MutiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return " ", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return " ", errors.New("rpc discovery: not supported select mode")
	}
}

func (d *MutiServerDiscovery) GetAll() (servers []string, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers = make([]string, len(servers))
	copy(servers, d.servers)
	return servers, nil
}
