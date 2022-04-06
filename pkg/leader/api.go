package leader

import (
	"encoding/json"
	"github.com/broswen/leader/pkg/worker"
	"net/http"
	"time"
)

func stateHandler(leader *Leader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(leader.State)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func updateHandler(leader *Leader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var update worker.HealthUpdate
		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if update.ID == "" {
			http.Error(w, "id must be specified", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		status := WorkerStatus{
			ID:         update.ID,
			Address:    r.RemoteAddr,
			Health:     Healthy,
			LastUpdate: time.Now(),
		}
		leader.ReceiveUpdate(status)

		err = json.NewEncoder(w).Encode(leader.Selected())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
