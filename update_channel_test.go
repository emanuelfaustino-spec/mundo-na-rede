package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStoreChannelDoesNotOfferDirectUpdates(t *testing.T) {
	previous := distributionChannel
	distributionChannel = "store"
	defer func() { distributionChannel = previous }()
	m := &manager{}
	get := httptest.NewRecorder()
	if !m.updateAPI(get, httptest.NewRequest(http.MethodGet, "/api/update", nil)) {
		t.Fatal("rota de atualização não atendida")
	}
	if get.Code != http.StatusOK {
		t.Fatalf("status: %d", get.Code)
	}
	var body struct {
		Available      bool `json:"available"`
		ManagedByStore bool `json:"managedByStore"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Available || !body.ManagedByStore {
		t.Fatalf("estado de atualização incorreto: %+v", body)
	}
	apply := httptest.NewRecorder()
	if !m.updateAPI(apply, httptest.NewRequest(http.MethodPost, "/api/update/apply", nil)) {
		t.Fatal("rota de aplicar atualização não atendida")
	}
	if apply.Code == http.StatusOK {
		t.Fatal("o canal Store não pode aplicar atualizações próprias")
	}
}
