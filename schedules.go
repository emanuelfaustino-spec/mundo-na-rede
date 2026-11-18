package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type schedule struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Minutes int    `json:"minutes"`
}

func (m *manager) schedules(id string) ([]schedule, error) {
	items := []schedule{}
	data, err := os.ReadFile(filepath.Join(m.dir(id), "schedules.json"))
	if errors.Is(err, os.ErrNotExist) {
		return items, nil
	}
	if err != nil {
		return nil, err
	}
	return items, json.Unmarshal(data, &items)
}
func (m *manager) configureSchedule(w http.ResponseWriter, r *http.Request, id string) {
	items, err := m.schedules(id)
	if err != nil {
		fail(w, err)
		return
	}
	if r.Method == "GET" {
		writeJSON(w, items)
		return
	}
	var item schedule
	if err = json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048)).Decode(&item); err != nil {
		fail(w, err)
		return
	}
	if item.ID != "" {
		for i, old := range items {
			if old.ID == item.ID {
				items = append(items[:i], items[i+1:]...)
				break
			}
		}
		delete(m.lastRun, item.ID)
	} else {
		item.Command = strings.TrimSpace(item.Command)
		if item.Minutes < 1 || item.Minutes > 10080 || item.Command == "" || len(item.Command) > 240 ||
			strings.ContainsAny(item.Command, "\r\n\x00") || strings.EqualFold(item.Command, "stop") || len(items) >= 25 {
			fail(w, errors.New("use um comando por tarefa, sem stop, e um intervalo de 1 a 10080 minutos"))
			return
		}
		item.ID = randomHex(8)
		items = append(items, item)
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fail(w, err)
		return
	}
	path := filepath.Join(m.dir(id), "schedules.json")
	if err = os.WriteFile(path, data, 0600); err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, items)
}
func (m *manager) scheduleLoop(done <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case now := <-ticker.C:
			m.runSchedules(now)
		}
	}
}
func (m *manager) runSchedules(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, p := range m.running {
		if p.state != "online" {
			continue
		}
		items, err := m.schedules(id)
		if err != nil {
			continue
		}
		for _, item := range items {
			previous := m.lastRun[item.ID]
			if previous.Before(p.started) {
				previous = p.started
			}
			if now.Sub(previous) < time.Duration(item.Minutes)*time.Minute {
				continue
			}
			m.lastRun[item.ID] = now
			sendCommand(p, item.Command)
		}
	}
}
