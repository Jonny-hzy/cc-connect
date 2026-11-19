package core

import (
	"context"
	"errors"
	"testing"
)

type recallStubPlatform struct {
	stubPlatformEngine
	id  string
	err error
}

func (p *recallStubPlatform) RecallLastMessage(context.Context, any) (string, error) {
	return p.id, p.err
}

func TestEngine_RecallLastMessage(t *testing.T) {
	p := &recallStubPlatform{stubPlatformEngine: stubPlatformEngine{n: "test"}, id: "om_recalled"}
	e := NewEngine("test", &stubAgent{}, []Platform{p}, "", LangEnglish)
	key := "test:user1"
	e.interactiveStates[key] = &interactiveState{platform: p, replyCtx: "ctx"}

	if _, err := e.RecallLastMessage(context.Background(), key); err == nil {
		t.Fatal("recall must be rejected while the project disables it")
	}

	e.SetAllowAgentRecall(true)
	id, err := e.RecallLastMessage(context.Background(), key)
	if err != nil {
		t.Fatalf("RecallLastMessage: %v", err)
	}
	if id != "om_recalled" {
		t.Fatalf("recalled id = %q", id)
	}

	p.err = errors.New("no recent outbound message to recall")
	if _, err := e.RecallLastMessage(context.Background(), key); err == nil {
		t.Fatal("platform error must surface")
	}
}

func TestEngine_RecallLastMessage_UnsupportedPlatform(t *testing.T) {
	p := &stubPlatformEngine{n: "test"}
	e := NewEngine("test", &stubAgent{}, []Platform{p}, "", LangEnglish)
	e.SetAllowAgentRecall(true)
	key := "test:user1"
	e.interactiveStates[key] = &interactiveState{platform: p, replyCtx: "ctx"}

	if _, err := e.RecallLastMessage(context.Background(), key); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("err = %v, want ErrNotSupported", err)
	}
}
