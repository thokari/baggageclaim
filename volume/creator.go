package volume

import "sync"

type Creator interface {
	Create(path string) (string, error)
}

type creator struct {
	provider          StrategyProvider
	repository        Repository
	defaultRootFsPath string
	mu                *sync.RWMutex
}

func NewCreator(provider StrategyProvider, repository Repository, defaultRootFsPath string) Creator {
	return &creator{
		provider:          provider,
		repository:        repository,
		defaultRootFsPath: defaultRootFsPath,
		mu:                &sync.RWMutex{},
	}
}

func (c *creator) Create(rootFsPath string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if rootFsPath == "" {
		rootFsPath = c.defaultRootFsPath
	}

	strategy, err := c.provider.ProvideStrategy(rootFsPath)
	if err != nil {
		return "", err
	}

	volume, err := c.repository.CreateVolume(strategy, Properties{}, 0)
	if err != nil {
		return "", err
	}

	return volume.Path, nil
}
