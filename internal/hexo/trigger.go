package hexo

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"outline-hexo-connector/internal/config"
	"time"
)

type Trigger struct {
	cfg         *config.Config
	ch          chan struct{}
	lastTrigger time.Time
}

func NewTrigger(cfg *config.Config) *Trigger {
	return &Trigger{
		cfg:         cfg,
		ch:          make(chan struct{}, 1),
		lastTrigger: time.Now(),
	}
}

func (t *Trigger) Watch(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("Stop watching for Hexo build triggers")
				return
			case <-t.ch:
				log.Printf("Trigger received - Starting Hexo build")
				t.lastTrigger = time.Now()
				err := t.build()
				if err != nil {
					log.Printf("Error building Hexo - %v", err)
				} else {
					log.Printf("Hexo build completed")
				}

				select {
				case <-ctx.Done():
					log.Printf("Stop watching for Hexo build triggers")
					return
				case <-time.After(time.Duration(t.cfg.HexoBuildInterval)):
					t.flush()
				}
			}
		}
	}()
}

func (t *Trigger) TriggerBuild() {
	select {
	case t.ch <- struct{}{}:
	default:
		log.Printf("Trigger pending - Next Hexo build will run after %v", time.Duration(t.cfg.HexoBuildInterval)-time.Since(t.lastTrigger))
	}
}

func (t *Trigger) flush() {
	select {
	case <-t.ch:
	default:
	}
}

func (t *Trigger) build() error {
	cmd := exec.Command("bash", "-c", t.cfg.HexoBuildCommand)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w - %s", err, output)
	}
	return nil
}
