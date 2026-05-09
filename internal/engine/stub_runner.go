package engine

import "github.com/justinush/maestro/internal/stub"

type stubRunner struct{}

func NewStubRunner() ActionRunner {
	return stubRunner{}
}

func (stubRunner) Run(ctx ActionContext) error {
	if len(ctx.Action.Params) == 0 {
		return nil
	}
	p, err := stub.DecodeParams(ctx.Action.Params)
	if err != nil {
		return err
	}
	return applyStubSet(ctx.Variables, p)
}
