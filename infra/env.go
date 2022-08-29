package infra

import (
	"context"
	"errors"
	"fmt"
)

var errEnv = errors.New("environment error")

type EnvCtxKey string

const ProjectRootDir EnvCtxKey = "_projectRootDir"

func root(ctx context.Context) (string, error) {
	if v, ok := ctx.Value(ProjectRootDir).(string); ok {
		return v, nil
	}
	return "", fmt.Errorf("%w: %s", errEnv, "can't find project root dir")
}
