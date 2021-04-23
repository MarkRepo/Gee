// Package rpc/service 通过反射实现结构体与服务的映射关系
package rpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

// 对 net/rpc 而言，一个函数需要能够被远程调用，需要满足如下五个条件：
// 1. the method’s type is exported. – 方法所属类型是导出的。
// 2. the method is exported. – 方法是导出的。
// 3. the method has two arguments, both exported (or builtin) types. – 两个入参，均为导出或内置类型。
// 4. the method’s second argument is a pointer. – 第二个入参必须是一个指针。
// 5. the method has return type error. – 返回值为 error 类型。

// 更直观一些：
// func (t *T) MethodName(argType T1, replyType *T2) error

// methodType  描述一个方法
type methodType struct {
	method    reflect.Method // method 方法的发射值
	ArgType   reflect.Type   // ArgType 方法请求参数类型
	ReplyType reflect.Type   // ReplyType 方法响应类型
	numCalls  uint64         // numCalls 统计方法调用次数
}

// NumCalls 返回调用次数
func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// newArgv 根据方法请求参数类型创建值，考虑是否为指针
func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	// argv may be a pointer type, or a value type
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

// newReply 根据方法响应类型创建值，reply必须是指针类型
func (m *methodType) newReply() reflect.Value {
	// reply must be a pointer type
	reply := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		reply.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		reply.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return reply
}

type service struct {
	name   string                 // name 映射的结构体的名称
	typ    reflect.Type           // typ 结构体的类型
	svr    reflect.Value          // svr 结构体实例, 第0个参数, 代表service实例自身
	method map[string]*methodType // method 存储结构体所有符合条件的方法
}

// newService 根据svr实例，创建service
func newService(svr interface{}) *service {
	s := new(service)
	s.svr = reflect.ValueOf(svr)
	s.name = reflect.Indirect(s.svr).Type().Name()
	s.typ = reflect.TypeOf(svr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

// registerMethods 注册结构体符合条件的方法
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuildInType(argType) || !isExportedOrBuildInType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuildInType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

// Call 根据 methodType 和 args，reply调用方法
func (s *service) Call(m *methodType, args, reply reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.svr, args, reply})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
