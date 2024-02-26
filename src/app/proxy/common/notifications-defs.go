package common

type CallbackOnBegin interface {
	Notify(finder PathFinder, scheme, profile string)
}

type CallbackOnBeginFunc func(finder PathFinder, scheme, profile string)

func (f CallbackOnBeginFunc) Notify(finder PathFinder, scheme, profile string) {
	f(finder, scheme, profile)
}

type LifecycleNotifications struct {
	OnBegin CallbackOnBeginFunc
}
