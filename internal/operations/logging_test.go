package operations

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

type sensitiveLogValue struct{}

func (sensitiveLogValue) LogValue() slog.Value {
	return slog.GroupValue(slog.Any("error", errors.New("private-lazy-value")))
}

func TestRedactionCoversErrorsBoundAttributesGroupsAndLogValues(t *testing.T) {
	var out bytes.Buffer
	logger := slog.New(RedactingHandler{Next: slog.NewJSONHandler(&out, nil)})
	logger = logger.With("error", errors.New("private-bound-value")).WithGroup("request")
	logger.Info("fixed_event", "error", errors.New("private-error-value"),
		"detail", map[string]string{"data": "private-map-value"},
		"nested", slog.GroupValue(slog.Any("error", errors.New("private-group-value"))),
		"lazy", sensitiveLogValue{}, "status", 503)
	logged := out.String()
	if strings.Contains(logged, "private-") {
		t.Fatalf("sensitive value leaked: %s", logged)
	}
	if !strings.Contains(logged, `"status":503`) || !strings.Contains(logged, "fixed_event") {
		t.Fatalf("safe event data lost: %s", logged)
	}
}
