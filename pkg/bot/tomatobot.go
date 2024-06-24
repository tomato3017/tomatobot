package bot

import "github.com/tomato3017/tomatobot/pkg/config"

type Tomatobot struct {
	cfg config.Config
}

func (t *Tomatobot) Run() error {
	panic("not implemented")
}

func NewTomatobot(cfg config.Config) *Tomatobot {
	return &Tomatobot{
		cfg: cfg,
	}
}
