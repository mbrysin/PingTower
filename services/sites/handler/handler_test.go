package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mbrysin/PingTower/services/sites/model"
)

type fakeSitesService struct {
	gotLimit  int
	gotOffset int
}

func (f *fakeSitesService) List(
	_ context.Context,
	limit, offset int,
) ([]model.Site, error) {
	f.gotLimit = limit
	f.gotOffset = offset
	return []model.Site{}, nil
}

func (f *fakeSitesService) Get(
	_ context.Context,
	_ int,
) (*model.Site, error) {
	return nil, nil
}

func TestListDefaultsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &fakeSitesService{}
	h := New(svc)

	router := gin.New()
	h.RegisterRoutes(router)

	req := httptest.NewRequest(http.MethodGet, "/sites", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if svc.gotLimit != 20 || svc.gotOffset != 0 {
		t.Fatalf(
			"pagination: got limit=%d offset=%d, want limit=20 offset=0",
			svc.gotLimit, svc.gotOffset,
		)
	}
}
