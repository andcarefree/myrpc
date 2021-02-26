package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/myrepo/myrpc/codec"

	"github.com/myrepo/myrpc"
	"github.com/myrepo/myrpc/xclient"
)

const (
	endpoints1 string = "localhost:12380"
	endpoints2 string = "localhost:22380"
	endpoints3 string = "localhost:32380"
	lease1     int64  = 1
	lease2     int64  = 2
	lease3     int64  = 3
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(lease int64) {
	var foo Foo
	l, _ := net.Listen("tcp", ":0")
	server := myrpc.NewServer()
	_ = server.Regsiter(&foo)
	rand.Seed(int64(time.Now().UnixNano()))
	a := rand.Intn(1000)
	register, err := myrpc.NewServiceRegister([]string{endpoints1, endpoints2, endpoints3}, "/Foo/"+string(a), l.Addr().String(), lease)
	if err != nil {
		log.Fatalln("rigister to endpoints error: ", err)
	}
	go register.ListenLeaseRespChan()
	server.Accept(l)
}

func foo(ctx context.Context, xc *xclient.XClient, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(addr1, addr2 string) {
	opt := &myrpc.Option{
		MagicNumber: 0x3bef5c,
		CodecType:   codec.JsonType,
		//ConnecTimeout: time.Second * 3,
		//HandleTimeout: time.Second * 3,
	}
	d := xclient.NewMutiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, opt)
	defer func() {
		err := xc.Close()
		if err != nil {
			log.Printf("client: close error: %s", err.Error())
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(context.Background(), xc, "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(addr1, addr2 string) {
	d := xclient.NewMutiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() {
		err := xc.Close()
		if err != nil {
			log.Printf("client: close error: %s", err.Error())
		}
	}()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(context.Background(), xc, "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(ctx, xc, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)

	//start three server points
	go startServer(lease1)
	go startServer(lease2)
	go startServer(lease3)

	time.Sleep(time.Minute * 5)
	// call(addr1, addr2)
	// broadcast(addr1, addr2)
}
