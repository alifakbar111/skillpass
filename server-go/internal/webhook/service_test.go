package webhook

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"skillpass-server-go/internal/testutil"
)

func TestWebhookCRUD(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	_, cID, _ := testutil.CreateCompanyUser(db, "wh@ex.com", "whco", "pass123", "Webhook Co", true)
	svc := NewService(db)

	t.Run("create", func(t *testing.T) {
		created, err := svc.Create(ctx, cID.String(), "https://example.com/hook")
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if created.Secret == "" {
			t.Fatal("expected secret in create response")
		}
		if len(created.Secret) != 64 {
			t.Fatalf("expected 64-char hex secret, got %d chars", len(created.Secret))
		}
	})

	t.Run("invalid URL rejected", func(t *testing.T) {
		if _, err := svc.Create(ctx, cID.String(), "not-a-url"); err == nil {
			t.Fatal("expected error for invalid URL")
		}
		if _, err := svc.Create(ctx, cID.String(), "ftp://example.com"); err == nil {
			t.Fatal("expected error for non-http scheme")
		}
	})

	t.Run("list omits secret", func(t *testing.T) {
		hooks, err := svc.List(ctx, cID.String())
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(hooks) != 1 {
			t.Fatalf("expected 1 webhook, got %d", len(hooks))
		}
		if hooks[0].Secret != "" {
			t.Fatal("list must not expose secrets")
		}
	})

	t.Run("delete", func(t *testing.T) {
		hooks, _ := svc.List(ctx, cID.String())
		found, err := svc.Delete(ctx, hooks[0].ID, cID.String())
		if err != nil || !found {
			t.Fatalf("delete: found=%v err=%v", found, err)
		}
		remaining, _ := svc.List(ctx, cID.String())
		if len(remaining) != 0 {
			t.Fatalf("expected 0 webhooks after delete, got %d", len(remaining))
		}
	})

	t.Run("delete other company's webhook fails", func(t *testing.T) {
		_, c2, _ := testutil.CreateCompanyUser(db, "wh2@ex.com", "whco2", "pass123", "Webhook Co 2", true)
		created, _ := svc.Create(ctx, cID.String(), "https://example.com/hook2")
		found, err := svc.Delete(ctx, created.ID, c2.String())
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		if found {
			t.Fatal("expected not-found when deleting another company's webhook")
		}
	})
}

func TestWebhookDispatch(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	_, cID, _ := testutil.CreateCompanyUser(db, "whd@ex.com", "whdco", "pass123", "Dispatch Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Platform Engineer", "Technology", true)
	_, pID, _ := testutil.CreateJobseeker(db, "whdjs@ex.com", "whdjs", "pass123", "Dispatch JS")

	var mu sync.Mutex
	var receivedBody []byte
	var receivedSig string
	done := make(chan struct{}, 1)

	// Create a test receiver on loopback. Since validateURL now rejects private
	// addresses, we insert the webhook directly into the DB to test dispatch.
	receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		defer mu.Unlock()
		receivedBody = body
		receivedSig = r.Header.Get("X-SkillPass-Signature")
		done <- struct{}{}
	}))
	defer receiver.Close()

	secret := generateSecretForTest()
	svc := NewService(db)
	_, err := db.ExecContext(ctx,
		`INSERT INTO company_webhooks (company_id, url, secret) VALUES ($1, $2, $3)`,
		cID, receiver.URL, secret)
	if err != nil {
		t.Fatalf("insert webhook: %v", err)
	}

	if err := svc.DispatchApplicationReceived(ctx, jID.String(), pID.String()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("webhook was not delivered within 5s")
	}

	mu.Lock()
	defer mu.Unlock()

	// Verify the HMAC signature matches the body.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(receivedBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	if receivedSig != expected {
		t.Fatalf("signature mismatch: got %s want %s", receivedSig, expected)
	}
}

// generateSecretForTest creates a deterministic secret to avoid relying on
// the non-deterministic rand.Read in production.
func generateSecretForTest() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	return hex.EncodeToString(b)
}
