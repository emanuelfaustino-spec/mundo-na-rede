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

func recommendedServerMemory(total, available int) int {
	limit := safeServerMemoryLimit(total, available)
	if limit == 0 {
		return 0
	}
	return min(limit, 2048) / 512 * 512
}

func validateServerMemory(requested int) error {
	total, available := memoryMiB()
	limit := safeServerMemoryLimit(total, available)
	if limit == 0 {
		return fmt.Errorf("há pouca RAM livre para iniciar o servidor; feche outros programas")
	}
	if requested > limit {
		return fmt.Errorf("limite de RAM acima do disponível: escolha até %d MiB ou libere memória", limit)
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
