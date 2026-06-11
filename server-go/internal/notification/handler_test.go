package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/testutil"
)

func TestNotificationService(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	uID, _, _ := testutil.CreateJobseeker(db, "notif@ex.com", "notif", "pass123", "Notif User")
	svc := NewService(db)

	t.Run("create and list", func(t *testing.T) {
		if err := svc.Create(ctx, uID.String(), "test", "Title 1", "Body 1", "/link1"); err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := svc.Create(ctx, uID.String(), "test", "Title 2", "Body 2", "/link2"); err != nil {
			t.Fatalf("create: %v", err)
		}

		res, err := svc.ListForUser(ctx, uID.String(), 50)
		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(res.Notifications) != 2 {
			t.Fatalf("expected 2 notifications, got %d", len(res.Notifications))
		}
		if res.UnreadCount != 2 {
			t.Fatalf("expected 2 unread, got %d", res.UnreadCount)
		}
		// Newest first
		if res.Notifications[0].Title != "Title 2" {
			t.Fatalf("expected newest first, got %q", res.Notifications[0].Title)
		}
	})

	t.Run("mark read", func(t *testing.T) {
		res, _ := svc.ListForUser(ctx, uID.String(), 50)
		id := res.Notifications[0].ID
		found, err := svc.MarkRead(ctx, id, uID.String())
		if err != nil || !found {
			t.Fatalf("mark read: found=%v err=%v", found, err)
		}
		count, _ := svc.CountUnread(ctx, uID.String())
		if count != 1 {
			t.Fatalf("expected 1 unread after marking one, got %d", count)
		}
	})

	t.Run("mark all read", func(t *testing.T) {
		if err := svc.MarkAllRead(ctx, uID.String()); err != nil {
			t.Fatalf("mark all: %v", err)
		}
		count, _ := svc.CountUnread(ctx, uID.String())
		if count != 0 {
			t.Fatalf("expected 0 unread, got %d", count)
		}
	})
}

func TestNotifyHelpers(t *testing.T) {
	db := testutil.SetupTestDB()
	ctx := context.Background()

	cu, cID, _ := testutil.CreateCompanyUser(db, "ncomp@ex.com", "ncomp", "pass123", "Notif Co", true)
	jID, _ := testutil.CreateJob(db, cID, "Backend Engineer", "Technology", true)
	jsUID, pID, _ := testutil.CreateJobseeker(db, "napp@ex.com", "napp", "pass123", "Applicant Name")
	appID, _ := testutil.CreateApplication(db, pID, jID, "applied")

	svc := NewService(db)

	t.Run("notify company of application", func(t *testing.T) {
		if err := svc.NotifyCompanyOfApplication(ctx, jID.String(), pID.String()); err != nil {
			t.Fatalf("notify company: %v", err)
		}
		res, _ := svc.ListForUser(ctx, cu.String(), 50)
		if len(res.Notifications) != 1 {
			t.Fatalf("expected 1 company notification, got %d", len(res.Notifications))
		}
		if res.Notifications[0].Type != "application_received" {
			t.Fatalf("unexpected type %q", res.Notifications[0].Type)
		}
	})

	t.Run("notify jobseeker of status", func(t *testing.T) {
		if err := svc.NotifyJobseekerOfStatus(ctx, appID.String(), "reviewed"); err != nil {
			t.Fatalf("notify jobseeker: %v", err)
		}
		res, _ := svc.ListForUser(ctx, jsUID.String(), 50)
		if len(res.Notifications) != 1 {
			t.Fatalf("expected 1 jobseeker notification, got %d", len(res.Notifications))
		}
		if res.Notifications[0].Type != "application_status" {
			t.Fatalf("unexpected type %q", res.Notifications[0].Type)
		}
	})
}

func TestNotificationHandler(t *testing.T) {
	db := testutil.SetupTestDB()
	uID, _, _ := testutil.CreateJobseeker(db, "nh@ex.com", "nh", "pass123", "NH User")
	tok := testutil.GenerateToken(uID.String(), "jobseeker", 15*time.Minute)

	svc := NewService(db)
	_ = svc.Create(context.Background(), uID.String(), "test", "Hello", "World", "/x")

	h := NewHandler(svc)
	router := gin.New()
	g := router.Group("/api/v1/notifications")
	g.Use(middleware.AuthRequired(testutil.TestJWTSecret))
	g.GET("/me", h.ListMine)
	g.PUT("/read-all", h.MarkAllRead)
	g.PUT("/:id/read", h.MarkRead)

	t.Run("list mine", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/notifications/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var res ListResult
		json.Unmarshal(w.Body.Bytes(), &res)
		if res.UnreadCount != 1 {
			t.Fatalf("expected 1 unread, got %d", res.UnreadCount)
		}
	})

	t.Run("mark all read", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/notifications/read-all", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("requires auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/notifications/me", nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}
