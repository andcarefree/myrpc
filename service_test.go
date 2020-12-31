package myrpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Nums1, Nums2 int
}

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Nums1 + args.Nums2
	return nil
}

func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Nums2 + args.Nums1
	return nil
}

func My_assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestNewService(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	My_assert(len(s.method) == 1, "wrong service Method, expect 1, but got %d", len(s.method))
	mType := s.method["Sum"]
	My_assert(mType != nil, "wrong method, Sum should not nil")
}

func TestMethodType_Call(t *testing.T) {
	var foo Foo
	s := newService(&foo)
	mType := s.method["Sum"]
	argv := mType.newArgv()
	replyv := mType.newReplyv()
	argv.Set(reflect.ValueOf(Args{Nums1: 1, Nums2: 3}))
	err := s.call(mType, argv, replyv)
	My_assert(err == nil && *replyv.Interface().(*int) == 4 && mType.NumCalls() == 1, "failed to call Foo.Sum")
}
