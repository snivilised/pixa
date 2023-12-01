package proxy

// the strategies look like they don't do much, so all this
// abstraction feels like overkill. instead the path finder
// could make a one calculation of destination path depending
// on strategy, using s simple func closure, eg we could
// funcs such as inlineDestination and ejectDestination() of
// the form func(source string) string. (rename this file
// strategy-funcs)

type outputStrategy interface {
	// Destination fills in the gap between the root and the destination
	Destination(source string) string
}

type inlineOutputStrategy struct {
}

func (s *inlineOutputStrategy) Destination(source string) string {
	_ = source
	// ./<item.Parent>/TRASH/<scheme>/<profile>/destination/<.item.Name>.<LEGACY>.ext
	return ""
}

type ejectOutputStrategy struct {
}

func (s *ejectOutputStrategy) Destination(source string) string {
	_ = source
	// ./<output>/TRASH/<scheme>/<profile>/destination
	return ""
}
