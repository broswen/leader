package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type HealthUpdate struct {
	ID string `json:"id"`
}

type Selected struct {
	Selected []string `json:"selected"`
}

type Worker struct {
	ID           string        `json:"id"`
	Working      bool          `json:"working"`
	LeaderAddr   string        `json:"leaderAddr"`
	UpdatePeriod time.Duration `json:"updatePeriod"`
	Selected     Selected      `json:"selected"`
	LastUpdate   time.Time     `json:"lastUpdate"`
}

func New(id, leaderAddr string) (*Worker, error) {
	return &Worker{
		ID:           id,
		Working:      false,
		LeaderAddr:   leaderAddr,
		UpdatePeriod: time.Second,
	}, nil
}

func (w *Worker) Router() chi.Router {
	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	return r
}

func (w *Worker) Work() {
	go func() {
		for _ = range time.Tick(time.Second) {
			if w.Working {
				log.Printf("%s: working...", w.ID)
				time.Sleep(3 * time.Second)
			}
		}
	}()
}

func (w *Worker) SendUpdates() <-chan error {
	errChan := make(chan error, 1)
	go func() {
		for _ = range time.Tick(w.UpdatePeriod) {
			log.Printf("%s: sending update to %s", w.ID, w.LeaderAddr)
			err := w.sendUpdate()
			if err != nil {
				errChan <- err
			}
		}
	}()
	return errChan
}

func (w *Worker) sendUpdate() error {
	var err error
	defer func() {
		if err != nil {
			FailedUpdateCount.WithLabelValues(w.ID).Inc()
		}
	}()

	b := bytes.NewBuffer([]byte{})
	err = json.NewEncoder(b).Encode(HealthUpdate{ID: w.ID})
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.LeaderAddr+"/update", b)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	var selected Selected
	err = json.NewDecoder(resp.Body).Decode(&selected)
	if err != nil {
		return err
	}
	w.updateSelected(selected.Selected)
	SuccessfulUpdateCount.WithLabelValues(w.ID).Inc()
	w.LastUpdate = time.Now()
	return nil
}

func (w *Worker) updateSelected(ids []string) {
	working := false
	for _, id := range ids {
		if id == w.ID {
			working = true
			break
		}
	}
	if working != w.Working {
		w.Working = working
		log.Printf("working: %v", w.Working)
	}
}
