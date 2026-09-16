package likes

import (
    "context"
    "errors"
    "testing"

    "github.com/google/uuid"

    "dating-platform/backend/internal/profiles"
)

type profileRepoStub struct {
    own    *profiles.Profile
    target *profiles.Profile
    err    error
}

func (r *profileRepoStub) Create(context.Context, *profiles.Profile) error { return nil }
func (r *profileRepoStub) GetByUserID(context.Context, uuid.UUID) (*profiles.Profile, error) { return r.own, r.err }
func (r *profileRepoStub) GetPublicByID(context.Context, uuid.UUID, uuid.UUID) (*profiles.Profile, error) { return r.target, r.err }
func (r *profileRepoStub) GetByIDAny(context.Context, uuid.UUID) (*profiles.Profile, error) { return r.target, r.err }
func (r *profileRepoStub) Update(context.Context, uuid.UUID, profiles.ProfilePatch) (*profiles.Profile, error) { return nil, nil }
func (r *profileRepoStub) AddPhoto(context.Context, uuid.UUID, *profiles.Photo) error { return nil }
func (r *profileRepoStub) ListPhotos(context.Context, uuid.UUID) ([]profiles.Photo, error) { return nil, nil }
func (r *profileRepoStub) CountPhotos(context.Context, uuid.UUID) (int, error) { return 0, nil }
func (r *profileRepoStub) GetPhoto(context.Context, uuid.UUID, uuid.UUID) (*profiles.Photo, error) { return nil, nil }
func (r *profileRepoStub) DeletePhoto(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type repoStub struct {
    matched bool
    removed bool
    sentPage int
    sentSize int
}

func (r *repoStub) Add(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return r.matched, nil }
func (r *repoStub) Remove(context.Context, uuid.UUID, uuid.UUID) error { r.removed = true; return nil }
func (r *repoStub) IsLiked(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return false, nil }
func (r *repoStub) HasMatch(context.Context, uuid.UUID, uuid.UUID) (bool, error) { return false, nil }
func (r *repoStub) ListSent(context.Context, uuid.UUID, int, int) (*ListResult, error) { return nil, nil }
func (r *repoStub) ListReceived(context.Context, uuid.UUID, int, int) (*ListResult, error) { return nil, nil }
func (r *repoStub) ListMatches(context.Context, uuid.UUID, int, int) (*MatchResult, error) { return nil, nil }

func TestServiceAddRejectsSelfLike(t *testing.T) {
    userID := uuid.New()
    own := &profiles.Profile{ID: uuid.New(), UserID: userID}
    svc := NewService(&repoStub{}, &profileRepoStub{own: own, target: own})

    if _, err := svc.Add(context.Background(), userID, own.ID); !errors.Is(err, ErrCannotLikeSelf) {
        t.Fatalf("Add self = %v, want ErrCannotLikeSelf", err)
    }
}

func TestServiceAddPropagatesHiddenProfile(t *testing.T) {
    userID := uuid.New()
    svc := NewService(&repoStub{}, &profileRepoStub{err: profiles.ErrNotFound})

    if _, err := svc.Add(context.Background(), userID, uuid.New()); !errors.Is(err, profiles.ErrNotFound) {
        t.Fatalf("Add hidden profile = %v, want profiles.ErrNotFound", err)
    }
}

func TestServiceAddReturnsMatchState(t *testing.T) {
    userID := uuid.New()
    own := &profiles.Profile{ID: uuid.New(), UserID: userID}
    target := &profiles.Profile{ID: uuid.New(), UserID: uuid.New()}
    svc := NewService(&repoStub{matched: true}, &profileRepoStub{own: own, target: target})

    matched, err := svc.Add(context.Background(), userID, target.ID)
    if err != nil || !matched {
        t.Fatalf("Add = (%v, %v), want (true, nil)", matched, err)
    }
}

func TestClampPaging(t *testing.T) {
    if page, size := clampPaging(0, MaxPageSize+1); page != 1 || size != MaxPageSize {
        t.Fatalf("clampPaging = (%d, %d)", page, size)
    }
}
