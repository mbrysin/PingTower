package service

import (
	"context"
	"errors"
	"testing"

	"github.com/mbrysin/PingTower/services/sites/model"
)

type fakeSiteRepository struct {
	createFn     func(context.Context, *model.Site) (*model.Site, error)
	getFn        func(context.Context, int) (*model.Site, error)
	listFn       func(context.Context, int, int) ([]model.Site, error)
	updateNameFn func(context.Context, int, string) (*model.Site, error)
	setEnabledFn func(context.Context, int, bool) (*model.Site, error)
	deleteFn     func(context.Context, int) error
}

func (f *fakeSiteRepository) Create(ctx context.Context, site *model.Site) (*model.Site, error) {
	if f.createFn != nil {
		return f.createFn(ctx, site)
	}
	return site, nil
}

func (f *fakeSiteRepository) Get(ctx context.Context, id int) (*model.Site, error) {
	if f.getFn != nil {
		return f.getFn(ctx, id)
	}
	return &model.Site{ID: id}, nil
}

func (f *fakeSiteRepository) List(ctx context.Context, limit, offset int) ([]model.Site, error) {
	if f.listFn != nil {
		return f.listFn(ctx, limit, offset)
	}
	return nil, nil
}

func (f *fakeSiteRepository) UpdateName(
	ctx context.Context,
	id int,
	name string,
) (*model.Site, error) {
	if f.updateNameFn != nil {
		return f.updateNameFn(ctx, id, name)
	}
	return &model.Site{ID: id, Name: name}, nil
}

func (f *fakeSiteRepository) SetEnabled(
	ctx context.Context,
	id int,
	enabled bool,
) (*model.Site, error) {
	if f.setEnabledFn != nil {
		return f.setEnabledFn(ctx, id, enabled)
	}
	return &model.Site{ID: id, Enabled: enabled}, nil
}

func (f *fakeSiteRepository) Delete(ctx context.Context, id int) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}

func TestSiteService_Create(t *testing.T) {
	tests := []struct {
		name     string
		siteName string
		rawURL   string
		wantErr  bool
		wantName string
		wantURL  string
	}{
		{
			name:     "valid site trims input",
			siteName: "  Example  ",
			rawURL:   "  https://example.com  ",
			wantName: "Example",
			wantURL:  "https://example.com",
		},
		{
			name:     "empty name",
			siteName: "   ",
			rawURL:   "https://example.com",
			wantErr:  true,
		},
		{
			name:     "invalid URL",
			siteName: "Example",
			rawURL:   "not-a-url",
			wantErr:  true,
		},
		{
			name:     "unsupported scheme",
			siteName: "Example",
			rawURL:   "ftp://example.com",
			wantErr:  true,
		},
		{
			name:     "URL without hostname",
			siteName: "Example",
			rawURL:   "https:///path",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *model.Site
			repo := &fakeSiteRepository{
				createFn: func(_ context.Context, site *model.Site) (*model.Site, error) {
					got = site
					return site, nil
				},
			}
			service := NewSiteService(repo)

			created, err := service.Create(context.Background(), tt.siteName, tt.rawURL)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidSite) {
					t.Fatalf("Create() error = %v, want ErrInvalidSite", err)
				}
				if got != nil {
					t.Fatal("repository Create() вызван для некорректного ввода")
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}
			if got == nil {
				t.Fatal("repository Create() не был вызван")
			}
			if created.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", created.Name, tt.wantName)
			}
			if created.URL != tt.wantURL {
				t.Errorf("URL = %q, want %q", created.URL, tt.wantURL)
			}
		})
	}
}

func TestSiteService_Create_RepositoryError(t *testing.T) {
	repoErr := errors.New("database error")
	service := NewSiteService(&fakeSiteRepository{
		createFn: func(context.Context, *model.Site) (*model.Site, error) {
			return nil, repoErr
		},
	})

	_, err := service.Create(context.Background(), "Example", "https://example.com")
	if !errors.Is(err, repoErr) {
		t.Fatalf("Create() error = %v, want %v", err, repoErr)
	}
}

func TestSiteService_Get(t *testing.T) {
	t.Run("valid id delegates to repository", func(t *testing.T) {
		const wantID = 42
		var gotID int

		service := NewSiteService(&fakeSiteRepository{
			getFn: func(_ context.Context, id int) (*model.Site, error) {
				gotID = id
				return &model.Site{ID: id}, nil
			},
		})

		site, err := service.Get(context.Background(), wantID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if gotID != wantID || site.ID != wantID {
			t.Errorf("Get() id = %d, repository received %d; want %d", site.ID, gotID, wantID)
		}
	})

	for _, id := range []int{0, -1} {
		t.Run("invalid id", func(t *testing.T) {
			called := false
			service := NewSiteService(&fakeSiteRepository{
				getFn: func(context.Context, int) (*model.Site, error) {
					called = true
					return nil, nil
				},
			})

			_, err := service.Get(context.Background(), id)
			if !errors.Is(err, ErrInvalidSite) {
				t.Errorf("Get() error = %v, want ErrInvalidSite", err)
			}
			if called {
				t.Error("repository Get() не должен вызываться")
			}
		})
	}
}

func TestSiteService_List(t *testing.T) {
	var gotLimit, gotOffset int
	wantSites := []model.Site{{ID: 1, Name: "Example"}}

	service := NewSiteService(&fakeSiteRepository{
		listFn: func(_ context.Context, limit, offset int) ([]model.Site, error) {
			gotLimit, gotOffset = limit, offset
			return wantSites, nil
		},
	})

	sites, err := service.List(context.Background(), 20, 5)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if gotLimit != 20 || gotOffset != 5 {
		t.Errorf("repository received limit=%d offset=%d, want limit=20 offset=5", gotLimit, gotOffset)
	}
	if len(sites) != 1 || sites[0].ID != wantSites[0].ID {
		t.Errorf("List() = %#v, want %#v", sites, wantSites)
	}
}

func TestSiteService_UpdateName(t *testing.T) {
	t.Run("trims name and delegates", func(t *testing.T) {
		var gotID int
		var gotName string

		service := NewSiteService(&fakeSiteRepository{
			updateNameFn: func(_ context.Context, id int, name string) (*model.Site, error) {
				gotID, gotName = id, name
				return &model.Site{ID: id, Name: name}, nil
			},
		})

		site, err := service.UpdateName(context.Background(), 12, "  New name  ")
		if err != nil {
			t.Fatalf("UpdateName() error = %v", err)
		}
		if gotID != 12 || gotName != "New name" {
			t.Errorf("repository received id=%d name=%q", gotID, gotName)
		}
		if site.Name != "New name" {
			t.Errorf("Name = %q, want %q", site.Name, "New name")
		}
	})

	tests := []struct {
		name string
		id   int
		text string
	}{
		{name: "zero id", id: 0, text: "Name"},
		{name: "negative id", id: -1, text: "Name"},
		{name: "empty name", id: 1, text: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			service := NewSiteService(&fakeSiteRepository{
				updateNameFn: func(context.Context, int, string) (*model.Site, error) {
					called = true
					return nil, nil
				},
			})

			_, err := service.UpdateName(context.Background(), tt.id, tt.text)
			if !errors.Is(err, ErrInvalidSite) {
				t.Errorf("UpdateName() error = %v, want ErrInvalidSite", err)
			}
			if called {
				t.Error("repository UpdateName() не должен вызываться")
			}
		})
	}
}

func TestSiteService_SetEnabled(t *testing.T) {
	var gotID int
	var gotEnabled bool

	service := NewSiteService(&fakeSiteRepository{
		setEnabledFn: func(_ context.Context, id int, enabled bool) (*model.Site, error) {
			gotID, gotEnabled = id, enabled
			return &model.Site{ID: id, Enabled: enabled}, nil
		},
	})

	site, err := service.SetEnabled(context.Background(), 7, true)
	if err != nil {
		t.Fatalf("SetEnabled() error = %v", err)
	}
	if gotID != 7 || !gotEnabled {
		t.Errorf("repository received id=%d enabled=%t, want id=7 enabled=true", gotID, gotEnabled)
	}
	if !site.Enabled {
		t.Error("SetEnabled() returned a disabled site")
	}
}

func TestSiteService_SetEnabled_InvalidID(t *testing.T) {
	called := false
	service := NewSiteService(&fakeSiteRepository{
		setEnabledFn: func(context.Context, int, bool) (*model.Site, error) {
			called = true
			return nil, nil
		},
	})

	_, err := service.SetEnabled(context.Background(), 0, true)
	if !errors.Is(err, ErrInvalidSite) {
		t.Errorf("SetEnabled() error = %v, want ErrInvalidSite", err)
	}
	if called {
		t.Error("repository SetEnabled() не должен вызываться")
	}
}

func TestSiteService_Delete(t *testing.T) {
	var gotID int
	service := NewSiteService(&fakeSiteRepository{
		deleteFn: func(_ context.Context, id int) error {
			gotID = id
			return nil
		},
	})

	if err := service.Delete(context.Background(), 9); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if gotID != 9 {
		t.Errorf("repository received id=%d, want 9", gotID)
	}
}

func TestSiteService_Delete_InvalidID(t *testing.T) {
	called := false
	service := NewSiteService(&fakeSiteRepository{
		deleteFn: func(context.Context, int) error {
			called = true
			return nil
		},
	})

	err := service.Delete(context.Background(), 0)
	if !errors.Is(err, ErrInvalidSite) {
		t.Errorf("Delete() error = %v, want ErrInvalidSite", err)
	}
	if called {
		t.Error("repository Delete() не должен вызываться")
	}
}
