package znet

import "zinx/ziface"

type BaseRouter struct{}

// 这里之所以BaseRouter的方法都为空，
// 是因为有的Router不希望有PreHandle或PostHandle
// 所以Router全部继承BaseRouter的好处是，不需要实现PreHandle和PostHandle也可以实例化
func (br *BaseRouter) PreHandle(req ziface.IRequset)  {}
func (br *BaseRouter) Handle(req ziface.IRequset)     {}
func (br *BaseRouter) PostHandle(req ziface.IRequset) {}
