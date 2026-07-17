package service

import (
	"bytes"
	"encoding/json"
	"time"
)

type OptionalTime struct {
	Set   bool
	Value *time.Time
}

func (s *OptionalString) UnmarshalJSON(data []byte) error {
	s.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		s.Value = nil
		return nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	s.Value = &value
	return nil
}

func (t *OptionalTime) UnmarshalJSON(data []byte) error {
	t.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		t.Value = nil
		return nil
	}
	var value time.Time
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	t.Value = &value
	return nil
}

type OptionalString struct {
	Set   bool
	Value *string
}
