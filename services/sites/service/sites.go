package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/mbrysin/PingTower/services/sites/model"
)

const (
	maxSiteNameLength = 100
	maxSiteURLBytes   = 2048
)

var ErrInvalidSite = errors.New("invalid site")

type SiteRepository interface {
	Create(ctx context.Context, site *model.Site) (*model.Site, error)
	Get(ctx context.Context, id int) (*model.Site, error)
	List(ctx context.Context, limit, offset int) ([]model.Site, error)
	UpdateName(ctx context.Context, id int, name string) (*model.Site, error)
	SetEnabled(ctx context.Context, id int, enabled bool) (*model.Site, error)
	Delete(ctx context.Context, id int) error
}

type SiteService struct {
	sites SiteRepository
}

func NewSiteService(sites SiteRepository) *SiteService {
	return &SiteService{sites: sites}
}

func (s *SiteService) Create(ctx context.Context, name string, rawURL string) (*model.Site, error) {
	name = strings.TrimSpace(name)
	rawURL = strings.TrimSpace(rawURL)

	if name == "" || utf8.RuneCountInString(name) > maxSiteNameLength ||
		!validURL(rawURL) || len(rawURL) > maxSiteURLBytes {
		return nil, ErrInvalidSite
	}

	return s.sites.Create(ctx, &model.Site{
		Name: name,
		URL:  rawURL,
	})
}

func validURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return (u.Scheme == "http" || u.Scheme == "https") && u.Hostname() != ""
}

func (s *SiteService) Get(ctx context.Context, id int) (*model.Site, error) {
	if id <= 0 {
		return nil, ErrInvalidSite
	}
	return s.sites.Get(ctx, id)
}

func (s *SiteService) List(ctx context.Context, limit, offset int) ([]model.Site, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("list sites: limit must be greater than 0")
	}
	if offset < 0 {
		return nil, fmt.Errorf("list sites: offset must not be negative")
	}
	return s.sites.List(ctx, limit, offset)
}

func (s *SiteService) UpdateName(ctx context.Context, id int, name string) (*model.Site, error) {
	name = strings.TrimSpace(name)
	if id <= 0 || name == "" || utf8.RuneCountInString(name) > maxSiteNameLength {
		return nil, ErrInvalidSite
	}

	return s.sites.UpdateName(ctx, id, name)
}

func (s *SiteService) SetEnabled(ctx context.Context, id int, enabled bool) (*model.Site, error) {
	if id <= 0 {
		return nil, ErrInvalidSite
	}

	return s.sites.SetEnabled(ctx, id, enabled)
}

func (s *SiteService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return ErrInvalidSite
	}

	return s.sites.Delete(ctx, id)
}
