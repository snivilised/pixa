package proxy

type deletionStrategy interface {
}

type inlineDeletionStrategy struct {
}

type ejectDeletionStrategy struct {
}
