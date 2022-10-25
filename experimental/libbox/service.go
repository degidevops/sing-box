package libbox

import (
	"context"

	"github.com/sagernet/sing-box"
	E "github.com/sagernet/sing/common/exceptions"
)

type Service struct {
	ctx      context.Context
	cancel   context.CancelFunc
	instance *box.Box
}

func NewService(configContent string) (*Service, error) {
	options, err := parseConfig(configContent)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	instance, err := box.New(ctx, options)
	if err != nil {
		cancel()
		return nil, E.Cause(err, "create service")
	}
	return &Service{
		ctx:      ctx,
		cancel:   cancel,
		instance: instance,
	}, nil
}

func (s *Service) Start() error {
	return s.instance.Start()
}

func (s *Service) Close() error {
	s.cancel()
	return s.instance.Close()
}
