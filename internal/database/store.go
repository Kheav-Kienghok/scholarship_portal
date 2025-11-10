package database

import (
	"context"

	importDB "github.com/Kheav-Kienghok/scholarship_portal/internal/database/db"
	"github.com/Kheav-Kienghok/scholarship_portal/internal/utils"
)

// SqlcReminderStore is exported and wraps queries.
type SqlcReminderStore struct {
	queries *importDB.Queries
}

func NewReminderStore(db *Database) *SqlcReminderStore {
	return &SqlcReminderStore{queries: db.Queries}
}

func (s *SqlcReminderStore) GetPendingReminders(ctx context.Context) ([]utils.ReminderRequest, error) {
	rows, err := s.queries.GetRemindersForToday(ctx)
	if err != nil {
		return nil, err
	}

	reminders := make([]utils.ReminderRequest, 0, len(rows))
	for _, r := range rows {
		reminders = append(reminders, utils.ReminderRequest{
			FullName:        r.Fullname.String,
			Email:           r.Email.String,
			ScholarshipName: r.Title.String,
			Description:     r.Description.String,
			Deadline:        r.DeadlineEnd.Time.Format("2006-01-02"),
			ApplyLink:       r.OfficialLink.String,
		})
	}
	return reminders, nil
}
