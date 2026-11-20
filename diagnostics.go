package main

import (
	"fmt"
	"net"
	"os"
)

type preflightCheck struct {
	Code  string `json:"code"`
	Level string `json:"level"`
	Text  string `json:"text"`
	Help  string `json:"help,omitempty"`
}

type preflightReport struct {
	Ready  bool             `json:"ready"`
	Checks []preflightCheck `json:"checks"`
}

// preflight is local-only and never starts a server or changes the firewall.
// The caller holds manager.mu; this code reads manager.running.
func (m *manager) preflight(p Profile, game string) preflightReport {
	report := preflightReport{Ready: true, Checks: []preflightCheck{}}
	add := func(code, level, message, help string) {
		report.Checks = append(report.Checks, preflightCheck{code, level, message, help})
		if level == "error" {
			report.Ready = false
		}
	}

	if stat, err := os.Stat(game); err != nil || !stat.IsDir() {
		add("world", "error", "Pasta do mundo não encontrada.", "Abra Mundo e importe ou prepare os arquivos.")
	} else {
		add("world", "ok", "Pasta do mundo encontrada.", "")
	}

	if _, err := m.command(p, game); err != nil && !m.cachedEngineAvailable(p) {
		add("engine", "error", err.Error(), "Prepare os motores na biblioteca do PC.")
	} else {
		add("engine", "ok", "Motor disponível na biblioteca.", "")
	}

	active := m.running[p.ID] != nil && m.running[p.ID].state != "parado" && m.running[p.ID].state != "falhou"
	if active {
		add("running", "ok", "Este servidor já está em execução.", "")
	} else {
		ports := []struct {
			network string
			port    int
		}{}
		if p.Engine == "java" || p.Engine == "both" {
			ports = append(ports, struct {
				network string
				port    int
			}{"tcp4", p.Port})
		} else {
			ports = append(ports, struct {
				network string
				port    int
			}{"udp4", p.Port})
		}
		if p.Engine == "bedrock" {
			ports = append(ports, struct {
				network string
				port    int
			}{"udp4", p.Port + 1})
		}
		if p.Engine == "both" {
			ports = append(ports, struct {
				network string
				port    int
			}{"udp4", crossplayPort})
		}
		for _, check := range ports {
			if err := probeLocalPort(check.network, check.port); err != nil {
				add("port", "error", fmt.Sprintf("Porta %d ocupada neste PC.", check.port), "Pare outro servidor ou mude a porta em Mundo.")
			} else {
				add("port", "ok", fmt.Sprintf("Porta %d livre neste PC.", check.port), "")
			}
		}
	}

	if err := validateServerMemory(p.Memory); err != nil {
		add("memory", "error", err.Error(), "Diminua o limite em Mundo, ou feche outros programas.")
	} else {
		add("memory", "ok", fmt.Sprintf("Limite de RAM: %d MiB.", p.Memory), "")
	}
	if p.Engine != "bedrock" {
		_, available := memoryMiB()
		if available > 0 && p.Memory > available-256 {
			add("memory", "warn", "Pouca RAM livre para o limite escolhido.", "Feche programas ou diminua a RAM do servidor em Mundo.")
		}
	}
	if len(localIPs()) == 0 {
		add("lan", "warn", "Nenhuma rede local disponível.", "Conecte o computador ao Wi-Fi ou à rede cabeada.")
	} else {
		add("lan", "ok", "Rede local identificada.", "")
	}
	return report
}

func probeLocalPort(network string, port int) error {
	address := fmt.Sprintf(":%d", port)
	if network == "tcp4" {
		listener, err := net.Listen("tcp4", address)
		if err != nil {
			return err
		}
		return listener.Close()
	}
	socket, err := net.ListenPacket("udp4", address)
	if err != nil {
		return err
	}
	return socket.Close()
}
