package notification

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/email"
)

// Service handles notification persistence and delivery using raw SQL.
// Raw SQL is used (rather than go-jet) so the notifications table does not
// require regenerating go-jet types.
type Service struct {
	db      *sql.DB
	emailer email.Sender
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// SetEmailer attaches an email sender so in-app notifications also reach the
// user's inbox. Optional — when nil, notifications stay in-app only.
func (s *Service) SetEmailer(e email.Sender) {
	s.emailer = e
}

// sendEmail delivers msg to the user's email address, best-effort: lookup or
// delivery failures are logged and never propagate to the caller.
func (s *Service) sendEmail(ctx context.Context, userID uuid.UUID, msg email.Message) {
	if s.emailer == nil {
		return
	}
	var to string
	if err := s.db.QueryRowContext(ctx,
		`SELECT email FROM users WHERE id = $1`, userID,
	).Scan(&to); err != nil {
		slog.Warn("email lookup failed for notification", "userID", userID, "error", err)
		return
	}
	if err := s.emailer.Send(ctx, to, msg.Subject, msg.HTML, msg.Text); err != nil {
		slog.Warn("notification email delivery failed", "to", to, "error", err)
	}
}

type Notification struct {
	ID        string  `json:"id"`
	UserID    string  `json:"userId"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Link      string  `json:"link"`
	ReadAt    *string `json:"readAt"`
	CreatedAt string  `json:"createdAt"`
}

type ListResult struct {
	Notifications []Notification `json:"notifications"`
	UnreadCount   int            `json:"unreadCount"`
}

// Create inserts a notification for a user. Best-effort callers may ignore the error.
func (s *Service) Create(ctx context.Context, userID, notifType, title, body, link string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, type, title, body, link)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, notifType, title, body, link,
	)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

// NotifyCompanyOfApplication notifies the company owner that a candidate applied.
// It looks up the company owner, job title, and candidate name from the given IDs.
func (s *Service) NotifyCompanyOfApplication(ctx context.Context, jobPostingID, jobseekerProfileID string) error {
	var companyUserID uuid.UUID
	var jobTitle string
	err := s.db.QueryRowContext(ctx,
		`SELECT c.user_id, j.title
		 FROM job_postings j
		 JOIN companies c ON c.id = j.company_id
		 WHERE j.id = $1`,
		jobPostingID,
	).Scan(&companyUserID, &jobTitle)
	if err != nil {
		return fmt.Errorf("lookup company user: %w", err)
	}

	var candidateName string
	err = s.db.QueryRowContext(ctx,
		`SELECT u.name
		 FROM jobseeker_profiles jp
		 JOIN users u ON u.id = jp.user_id
		 WHERE jp.id = $1`,
		jobseekerProfileID,
	).Scan(&candidateName)
	if err != nil {
		return fmt.Errorf("lookup candidate name: %w", err)
	}

	title := "New application"
	body := fmt.Sprintf("%s applied to %q.", candidateName, jobTitle)
	if err := s.Create(ctx, companyUserID.String(), "application_received", title, body, "/company/applications"); err != nil {
		return err
	}
	s.sendEmail(ctx, companyUserID,
		email.ApplicationReceivedEmail(jobTitle, candidateName, email.AppBaseURL()+"/company/applications"))
	return nil
}

// NotifyJobseekerOfStatus notifies the jobseeker that their application status changed.
func (s *Service) NotifyJobseekerOfStatus(ctx context.Context, applicationID, status string) error {
	var jobseekerUserID uuid.UUID
	var jobTitle string
	err := s.db.QueryRowContext(ctx,
		`SELECT jp.user_id, j.title
		 FROM applications a
		 JOIN jobseeker_profiles jp ON jp.id = a.jobseeker_id
		 JOIN job_postings j ON j.id = a.job_posting_id
		 WHERE a.id = $1`,
		applicationID,
	).Scan(&jobseekerUserID, &jobTitle)
	if err != nil {
		return fmt.Errorf("lookup jobseeker user: %w", err)
	}

	title := "Application update"
	body := fmt.Sprintf("Your application for %q is now %q.", jobTitle, status)
	if err := s.Create(ctx, jobseekerUserID.String(), "application_status", title, body, "/jobseeker/applications"); err != nil {
		return err
	}
	s.sendEmail(ctx, jobseekerUserID,
		email.StatusUpdateEmail(jobTitle, status, email.AppBaseURL()+"/jobseeker/applications"))
	return nil
}

// NotifyJobseekerOfNote notifies the jobseeker that the company left a note on their application.
func (s *Service) NotifyJobseekerOfNote(ctx context.Context, applicationID string) error {
	var jobseekerUserID uuid.UUID
	var jobTitle string
	err := s.db.QueryRowContext(ctx,
		`SELECT jp.user_id, j.title
		 FROM applications a
		 JOIN jobseeker_profiles jp ON jp.id = a.jobseeker_id
		 JOIN job_postings j ON j.id = a.job_posting_id
		 WHERE a.id = $1`,
		applicationID,
	).Scan(&jobseekerUserID, &jobTitle)
	if err != nil {
		return fmt.Errorf("lookup jobseeker user: %w", err)
	}

	title := "New message"
	body := fmt.Sprintf("The company left a note on your application for %q.", jobTitle)
	if err := s.Create(ctx, jobseekerUserID.String(), "application_note", title, body, "/jobseeker/applications"); err != nil {
		return err
	}
	s.sendEmail(ctx, jobseekerUserID,
		email.NoteEmail(jobTitle, email.AppBaseURL()+"/jobseeker/applications"))
	return nil
}

// ListForUser returns recent notifications plus the unread count.
func (s *Service) ListForUser(ctx context.Context, userID string, limit int) (*ListResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, type, title, body, link, read_at, created_at
		 FROM notifications
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		var id, uid uuid.UUID
		var readAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&id, &uid, &n.Type, &n.Title, &n.Body, &n.Link, &readAt, &createdAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		n.ID = id.String()
		n.UserID = uid.String()
		n.CreatedAt = createdAt.Format(time.RFC3339)
		if readAt.Valid {
			v := readAt.Time.Format(time.RFC3339)
			n.ReadAt = &v
		}
		notifications = append(notifications, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	unread, err := s.CountUnread(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ListResult{Notifications: notifications, UnreadCount: unread}, nil
}

func (s *Service) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}

// MarkRead marks a single notification read, scoped to the owner. Returns false if not found.
func (s *Service) MarkRead(ctx context.Context, notificationID, userID string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = now()
		 WHERE id = $1 AND user_id = $2 AND read_at IS NULL`,
		notificationID, userID,
	)
	if err != nil {
		return false, fmt.Errorf("mark read: %w", err)
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = now()
		 WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}
