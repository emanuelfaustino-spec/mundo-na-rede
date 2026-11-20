package main

import (
	"errors"
	"time"
)

// generationJob tracks a real Bedrock ticking-area sweep.  It does not claim
// that chunks were generated until the server has had time to tick each area.
// BDS only accepts a small ticking radius, so the sweep moves that area over
// the requested square around the world spawn.
type generationJob struct {
	Radius, Total, Completed int
	State                    string
	Started                  time.Time
	Cancel                   chan struct{}
}

func generationPoints(radius int) [][2]int {
	if radius < 4 {
		radius = 4
	}
	const step = 8 // radius 4 on each side: adjacent areas touch.
	points := make([][2]int, 0)
	for z := -radius; z <= radius; z += step {
		for x := -radius; x <= radius; x += step {
			points = append(points, [2]int{x * 16, z * 16})
		}
	}
	return points
}

func (m *manager) startPregeneration(proc *process, radius int) error {
	if radius < 4 || radius > 64 {
		return errors.New("escolha entre 4 e 64 chunks para pré-gerar")
	}
	if proc.generation != nil && proc.generation.State == "running" {
		return errors.New("já existe uma pré-geração em andamento")
	}
	points := generationPoints(radius)
	job := &generationJob{Radius: radius, Total: len(points), State: "running", Started: time.Now(), Cancel: make(chan struct{})}
	proc.generation = job
	go m.runPregeneration(proc, job, points)
	return nil
}

func (m *manager) runPregeneration(proc *process, job *generationJob, points [][2]int) {
	const name = "mnr-preload"
	for _, point := range points {
		select {
		case <-job.Cancel:
			_ = sendCommand(proc, "tickingarea remove "+name)
			m.mu.Lock()
			job.State = "cancelled"
			m.mu.Unlock()
			return
		case <-proc.done:
			m.mu.Lock()
			job.State = "stopped"
			m.mu.Unlock()
			return
		default:
		}
		if err := sendCommand(proc, "tickingarea remove "+name); err != nil {
			m.mu.Lock()
			job.State = "failed"
			m.mu.Unlock()
			return
		}
		if err := sendCommand(proc, "tickingarea add circle "+itoa(point[0])+" 0 "+itoa(point[1])+" 4 "+name); err != nil {
			m.mu.Lock()
			job.State = "failed"
			m.mu.Unlock()
			return
		}
		timer := time.NewTimer(1500 * time.Millisecond)
		select {
		case <-job.Cancel:
			timer.Stop()
			_ = sendCommand(proc, "tickingarea remove "+name)
			m.mu.Lock()
			job.State = "cancelled"
			m.mu.Unlock()
			return
		case <-proc.done:
			timer.Stop()
			m.mu.Lock()
			job.State = "stopped"
			m.mu.Unlock()
			return
		case <-timer.C:
		}
		m.mu.Lock()
		job.Completed++
		m.mu.Unlock()
	}
	_ = sendCommand(proc, "tickingarea remove "+name)
	m.mu.Lock()
	job.State = "completed"
	m.mu.Unlock()
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	buf := [12]byte{}
	i := len(buf)
	for value > 0 {
		i--
		buf[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
