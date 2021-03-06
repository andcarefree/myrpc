package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//RPCRegistry as a medium between client and server
type RPCRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

//ServerItem is a server with starttime
type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "_myrpc_ registry"
	defaultTimeout = time.Minute * 5
)

// New return an RPCRegistry instanse with timeout
func New(timeout time.Duration) *RPCRegistry {
	return &RPCRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

// DefaultRegistry provide a RPCRegistry instance
// the Registry'timeout == 5 minute
var DefaultRegistry = New(defaultTimeout)

// PutServer registered server with addr
func (r *RPCRegistry) PutServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if server := r.servers[addr]; server == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		server.start = time.Now()
	}
}

// AliveServers return all keep-alive server
func (r *RPCRegistry) AliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

//ServerHTTP is RPCRegistry's way to communicate with client and server
//HTTP GET means AliveServers()
//HTTP POST means PutServer()
func (r *RPCRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-rpc-Servers", strings.Join(r.AliveServers(), ", "))
	case "POST":
		addr := req.Header.Get("X-rpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.PutServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//HandleHTTP means privide HTTP server
func (r *RPCRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

//HandleHTTP privide default server in "_myrpc_ registry"
func HandleHTTP() {
	DefaultRegistry.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-rpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err: ", err)
		return err
	}
	return nil
}
