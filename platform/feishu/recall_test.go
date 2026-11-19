package feishu

import (
	"context"
	"testing"
)

func TestPlatform_OutboundTracking(t *testing.T) {
	p := &Platform{}
	p.recordOutbound("chat-1", "om_1")
	p.recordOutbound("chat-1", "om_2")
	p.recordOutbound("chat-2", "om_3")

	if got := p.takeOutbound("chat-1"); got != "om_2" {
		t.Fatalf("takeOutbound = %q, want newest (om_2)", got)
	}
	if got := p.takeOutbound("chat-1"); got != "" {
		t.Fatalf("second take = %q, want empty", got)
	}
	if got := p.takeOutbound("chat-2"); got != "om_3" {
		t.Fatalf("takeOutbound chat-2 = %q", got)
	}
}

func TestPlatform_RecallLastMessage_NoContext(t *testing.T) {
	p := &Platform{}
	if _, err := p.RecallLastMessage(context.Background(), "not-a-reply-context"); err == nil {
		t.Fatal("invalid reply context must fail")
	}
	if _, err := p.RecallLastMessage(context.Background(), replyContext{chatID: "chat-1"}); err == nil {
		t.Fatal("empty outbox must fail")
	}
}
