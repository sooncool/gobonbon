package router

import "gobonbon/iface"

type BaseRouter struct{}

// PreHandle -
func (br *BaseRouter) PreHandle(req iface.IRequest) {}

// Handle -
func (br *BaseRouter) Handle(req iface.IRequest) {}

// PostHandle -
func (br *BaseRouter) PostHandle(req iface.IRequest) {}
