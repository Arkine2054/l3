package service

import (
	"context"
	"log"
)

type SimulatedSender struct{}

func NewSimulatedSender() *SimulatedSender { return &SimulatedSender{} }
func (s *SimulatedSender) Name() string    { return "simulated" }
func (s *SimulatedSender) Send(ctx context.Context, recipient, message string) error {
	log.Printf("[SIMULATED] to=%s message=%s\n", recipient, message)
	return nil
}
