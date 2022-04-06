package leader

import (
	"github.com/broswen/leader/pkg/worker"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"sync"
	"time"
)

type Health string

var Unknown Health = "UNKNOWN"
var Healthy Health = "HEALTHY"
var Unhealthy Health = "UNHEALTHY"

type Leader struct {
	Address          string `json:"address"`
	HealthyTimeout   time.Duration
	UnhealthyTimeout time.Duration
	State            State
}

type State struct {
	mu               sync.RWMutex
	WantedWorkers    int                      `json:"wantedWorkers"`
	Workers          map[string]*WorkerStatus `json:"workers"`
	UnhealthyWorkers map[string]*WorkerStatus `json:"unhealthyWorkers"`
	Selected         map[string]bool          `json:"selected"`
}

type WorkerStatus struct {
	ID         string    `json:"id"`
	Address    string    `json:"addr"`
	Health     Health    `json:"health"`
	LastUpdate time.Time `json:"lastUpdate"`
}

func New(addr string, workers int) (*Leader, error) {
	return &Leader{
		Address:          addr,
		HealthyTimeout:   5 * time.Second,
		UnhealthyTimeout: 5 * time.Minute,
		State: State{
			WantedWorkers:    workers,
			Workers:          make(map[string]*WorkerStatus, 0),
			UnhealthyWorkers: make(map[string]*WorkerStatus, 0),
			Selected:         make(map[string]bool, 0),
		},
	}, nil
}

func (l *Leader) Router() chi.Router {
	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/state", stateHandler(l))
	r.Post("/update", updateHandler(l))
	return r
}

func (l *Leader) Update() {
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				//check for unhealthy workers
				l.checkWorkers()
				//if not enough selected, elect a new healthy worker
				if len(l.State.Selected) < l.State.WantedWorkers {
					l.selectWorker()
				}
			}
		}
	}()
}

func (l *Leader) checkWorkers() {
	l.State.mu.Lock()
	defer l.State.mu.Unlock()
	for id, status := range l.State.Workers {
		if status.Health == Healthy && time.Since(status.LastUpdate) > l.HealthyTimeout {
			//if workers was previously healthy
			//move to unhealthy workers list
			delete(l.State.Workers, id)
			l.State.UnhealthyWorkers[id] = status

			//if previously selected, remove
			if l.State.Selected[id] {
				delete(l.State.Selected, id)
			}
		}
	}

	for id, status := range l.State.UnhealthyWorkers {
		if status.Health == Unhealthy && time.Since(status.LastUpdate) > l.UnhealthyTimeout {
			// worker has been unhealthy for too long, remove from state
			delete(l.State.UnhealthyWorkers, id)
		}
	}

	UnhealthyWorkers.Set(float64(len(l.State.UnhealthyWorkers)))
	HealthyWorkers.Set(float64(len(l.State.Workers)))
	SelectedWorkers.Set(float64(len(l.State.Selected)))
	WantedWorkers.Set(float64(l.State.WantedWorkers))
}

func (l *Leader) selectWorker() {
	l.State.mu.Lock()
	defer l.State.mu.Unlock()
	for id, _ := range l.State.Workers {
		if !l.State.Selected[id] {
			l.State.Selected[id] = true
			break
		}
	}
}

func (l *Leader) ReceiveUpdate(status WorkerStatus) {
	l.State.mu.Lock()
	defer l.State.mu.Unlock()
	l.State.Workers[status.ID] = &status
	delete(l.State.UnhealthyWorkers, status.ID)
	UpdateCount.WithLabelValues(status.ID).Inc()
}

func (l *Leader) Selected() worker.Selected {
	l.State.mu.Lock()
	defer l.State.mu.Unlock()
	workers := make([]string, 0)
	for k, v := range l.State.Selected {
		if v {
			workers = append(workers, k)
		}
	}
	return worker.Selected{
		workers,
	}
}
