package pg

import "context"

type StaticProvider struct {
	DB Executor
}

func (p StaticProvider) Get(_ context.Context) Executor { return p.DB }
