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
	handle := p.generateHandle(rootFsPath)

	liveVolume, exists, err := p.filesystem.LookupVolume(handle)
	if err != nil {
		return nil, err
	}

	if !exists {
		initVolume, err := p.filesystem.NewVolume(handle)
		if err != nil {
			return nil, err
		}

		err = p.runner.Run(exec.Command("sh", "-c", fmt.Sprintf("cp -R %s/* %s", rootFsPath, initVolume.DataPath())))
		if err != nil {
			initVolume.Destroy()
			return nil, err
		}

		liveVolume, err = initVolume.Initialize()
		if err != nil {
			initVolume.Destroy()
			return nil, err
		}
	}

	return &COWStrategy{ParentHandle: liveVolume.Handle()}, nil
}

func (p *gardenStrategyProvider) generateHandle(rootFsPath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(rootFsPath)))
}
