package waitlist

import "testing"

func TestNormalizePhoneUSConvenience(t *testing.T) {
	got, err := normalizePhone("(717) 555-1212")
	if err != nil {
		t.Fatalf("normalizePhone returned error: %v", err)
	}
	if got != "+17175551212" {
		t.Fatalf("got %q, want %q", got, "+17175551212")
	}
}

func TestNormalizePhoneInternational(t *testing.T) {
	got, err := normalizePhone("+44 20 7946 0958")
	if err != nil {
		t.Fatalf("normalizePhone returned error: %v", err)
	}
	if got != "+442079460958" {
		t.Fatalf("got %q, want %q", got, "+442079460958")
	}
}

func TestNormalizePhoneRejectsLetters(t *testing.T) {
	if _, err := normalizePhone("717-CLEAR"); err == nil {
		t.Fatal("expected invalid phone error")
	}
}

func TestNormalizeEmail(t *testing.T) {
	got, err := normalizeEmail("Chris@Example.com")
	if err != nil {
		t.Fatalf("normalizeEmail returned error: %v", err)
	}
	if got != "chris@example.com" {
		t.Fatalf("got %q, want %q", got, "chris@example.com")
	}
}
