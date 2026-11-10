package main

import (
	"context"
	"testing"

	"light-launcher/core/internal/api"
)

func TestRegisterServicesRunsEachRegistrar(t *testing.T) {
	previous := serviceRegistrars
	t.Cleanup(func() { serviceRegistrars = previous })
	serviceRegistrars = []func(*api.Router, api.Events) error{
		func(r *api.Router, _ api.Events) error {
			r.Handle("Ping", api.NoInput(func(context.Context) (string, error) { return "pong", nil }))
			return nil
		},
	}

	router := api.NewRouter()
	if err := registerServices(router, api.NewHub()); err != nil {
		t.Fatal(err)
	}
	if methods := router.Methods(); len(methods) != 1 || methods[0] != "Ping" {
		t.Fatalf("methods = %v", methods)
	}
}
