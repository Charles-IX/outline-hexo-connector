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
	cfg             *config.Config
	timer           *time.Timer
	timerCh         <-chan time.Time
	triggerCh       chan struct{}
	lastTriggerTime time.Time
	pending         bool
}

func NewTrigger(cfg *config.Config) *Trigger {
	return &Trigger{
		cfg:       cfg,
		triggerCh: make(chan struct{}, 1),
	}
}

func (t *Trigger) Watch(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				if t.timer != nil {
					t.timer.Stop()
				}
				log.Printf("Stop watching for Hexo build triggers")
				return

			case <-t.triggerCh:
				if t.timer == nil {
					log.Printf("Trigger received - Starting Hexo build")
					err := t.build()
					if err != nil {
						log.Printf("Error building Hexo - %v", err)
					} else {
						log.Printf("Hexo build completed")
					}

					t.timer = time.NewTimer(time.Duration(t.cfg.HexoBuildInterval) * time.Second)
					t.timerCh = t.timer.C
					t.lastTriggerTime = time.Now()
					t.pending = false

				} else {
					t.pending = true
					remaining := time.Until(t.lastTriggerTime.Add(time.Duration(t.cfg.HexoBuildInterval) * time.Second))
					log.Printf("Trigger pending - Will build after %v", remaining)
				}

			case <-t.timerCh:
				if t.pending {
					log.Printf("Trigger timer expired with pending tasks - Starting Hexo build")
					err := t.build()
					if err != nil {
						log.Printf("Error building Hexo - %v", err)
					} else {
						log.Printf("Hexo build completed")
					}

					t.timer.Reset(time.Duration(t.cfg.HexoBuildInterval) * time.Second)
					t.lastTriggerTime = time.Now()
					t.pending = false
				} else {
					log.Printf("Trigger timer expired with no pending tasks - Back to idle")
					t.timer = nil
					t.timerCh = nil
				}
			}
		}
	}()
}

func (t *Trigger) TriggerBuild() {
	select {
	case t.triggerCh <- struct{}{}:
	default:
		remaining := time.Until(t.lastTriggerTime.Add(time.Duration(t.cfg.HexoBuildInterval) * time.Second))
		log.Printf("Trigger pending - Will build after %v", remaining)
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
