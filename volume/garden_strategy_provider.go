package volume

import (
	"crypto/sha256"
	"fmt"
	"os/exec"

	"github.com/cloudfoundry/gunk/command_runner"
)

//go:generate counterfeiter . StrategyProvider
type StrategyProvider interface {
	ProvideStrategy(rootFsPath string) (Strategy, error)
}

type gardenStrategyProvider struct {
	filesystem Filesystem
	runner     command_runner.CommandRunner
}

func NewGardenStrategyProvider(filesystem Filesystem, runner command_runner.CommandRunner) StrategyProvider {
	return &gardenStrategyProvider{
		filesystem: filesystem,
		runner:     runner,
	}
}

func (p *gardenStrategyProvider) ProvideStrategy(rootFsPath string) (Strategy, error) {
	initVolume, err := p.filesystem.NewVolume(shaID(rootFsPath))
	if err != nil {
		return nil, err
	}

	err = p.runner.Run(exec.Command("cp", "-R", rootFsPath, initVolume.DataPath()))
	if err != nil {
		initVolume.Destroy()
		return nil, err
	}

	liveVolume, err := initVolume.Initialize()
	if err != nil {
		initVolume.Destroy()
		return nil, err
	}

	return &COWStrategy{ParentHandle: liveVolume.Handle()}, nil
}

func shaID(rootFsPath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(rootFsPath)))
}
