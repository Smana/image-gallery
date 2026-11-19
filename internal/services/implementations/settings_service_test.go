package implementations

import (
	"context"
	"testing"

	"image-gallery/internal/domain/settings"
)

// upsertRecorder is a settings repository that only records upserts.
type upsertRecorder struct {
	settings.Repository
	got *settings.UserSettings
}

func (r *upsertRecorder) Upsert(_ context.Context, s *settings.UserSettings) (*settings.UserSettings, error) {
	r.got = s
	return s, nil
}

// GetSettings without a user reads the 'default' row, so resetting without a
// user must write that row rather than insert a NULL-user one.
func TestResetSettingsWithoutAUserResetsTheDefaultRow(t *testing.T) {
	repo := &upsertRecorder{}
	if _, err := NewSettingsService(repo, nil).ResetSettings(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if repo.got.UserID == nil || *repo.got.UserID != defaultUserID {
		t.Errorf("upserted user_id = %v, want %q", repo.got.UserID, defaultUserID)
	}
}

func TestResetSettingsKeepsAnExplicitUser(t *testing.T) {
	repo := &upsertRecorder{}
	user := "alice"
	if _, err := NewSettingsService(repo, nil).ResetSettings(context.Background(), &user); err != nil {
		t.Fatal(err)
	}
	if repo.got.UserID == nil || *repo.got.UserID != user {
		t.Errorf("upserted user_id = %v, want %q", repo.got.UserID, user)
	}
}
