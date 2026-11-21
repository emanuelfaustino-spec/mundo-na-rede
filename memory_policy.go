package main

import "fmt"

// Leave memory for the operating system and applications already running.
// This is a conservative start-time budget, not a prediction of future usage.
func safeServerMemoryLimit(total, available int) int {
	if total <= 0 || available <= 0 {
		return 0
	}
	reserve := memoryReserveMiB(total)
	limit := min(available-reserve, total*60/100, 32768)
	if limit < 512 {
		return 0
	}
	return limit
}

// memoryReserveMiB protects Windows and other active applications from the
// short memory spikes that happen while a world is loading.
func memoryReserveMiB(total int) int {
	return max(1024, total/4)
}

// serverMemoryTiers exposes only calm, easy-to-read half-GiB choices to the
// primary UI: 512 MiB, 1 GiB, 1.5 GiB, 2 GiB and so on.
func serverMemoryTiers(safe int) []int {
	tiers := []int{}
	for value := 512; value <= safe; value += 512 {
		tiers = append(tiers, value)
	}
	return tiers
}

// manualServerMemoryLimit is a deliberately higher ceiling for a host that is
// otherwise idle. It still leaves 512 MiB for Windows and never exceeds 75% of
// installed RAM. Reaching this range requires a visible acknowledgement in UI.
func manualServerMemoryLimit(total, available int) int {
	if total <= 0 || available <= 0 {
		return 0
	}
	limit := min(available-512, total*75/100, 32768)
	if limit < 512 {
		return 0
	}
	return limit
}

func recommendedServerMemory(total, available int) int {
	limit := safeServerMemoryLimit(total, available)
	if limit == 0 {
		return 0
	}
	return min(limit, 2048) / 512 * 512
}

func validateServerMemory(requested int, override bool) error {
	total, available := memoryMiB()
	safe := safeServerMemoryLimit(total, available)
	if safe == 0 {
		return fmt.Errorf("há pouca RAM livre para iniciar o servidor; feche outros programas")
	}
	if requested > safe {
		limit := manualServerMemoryLimit(total, available)
		if !override {
			return fmt.Errorf("limite acima do recomendado: confirme o uso avançado de RAM ou escolha até %d MiB", safe/512*512)
		}
		if requested > limit {
			return fmt.Errorf("limite de RAM acima do disponível agora: escolha até %d MiB ou libere memória", limit/512*512)
		}
	}
	if requested < 512 {
		return fmt.Errorf("o servidor precisa de pelo menos 512 MiB")
	}
	return nil
}

// Java needs native memory beyond the managed heap. Keep some of the
// configured per-process allowance outside -Xmx.
func javaHeapLimit(processLimitMiB int) int {
	return max(256, processLimitMiB-384)
}
