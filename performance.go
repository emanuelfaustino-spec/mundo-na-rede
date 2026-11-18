package main

type cpuSample struct {
	total, idle, process uint64
	memory               uint64
}

func (p *process) measurements() map[string]any {
	sample := readCPU(p.cmd.Process.Pid)
	out := map[string]any{}
	if sample.memory > 0 {
		out["serverMemory"] = sample.memory
	}
	previous := p.sample
	if previous.total > 0 && sample.total > previous.total {
		delta := sample.total - previous.total
		if sample.idle >= previous.idle && sample.idle-previous.idle <= delta {
			out["hostCPU"] = 100 * float64(delta-(sample.idle-previous.idle)) / float64(delta)
		}
		if sample.process >= previous.process && sample.process-previous.process <= delta {
			out["serverCPU"] = 100 * float64(sample.process-previous.process) / float64(delta)
		}
	}
	p.sample = sample
	return out
}
