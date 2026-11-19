package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/domain/settings"
)

// listImages records the request it received and answers with resp.
type listImages struct {
	stubImages
	got  *image.ListImagesRequest
	resp *image.ListImagesResponse
}

func (s *listImages) ListImages(_ context.Context, req *image.ListImagesRequest) (*image.ListImagesResponse, error) {
	s.got = req
	return s.resp, nil
}

func serveList(t *testing.T, svc *listImages, target string, hdr map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	h := &Handler{imageService: svc}
	r := chi.NewRouter()
	r.Get("/api/images", h.listImagesHandler)
	req := httptest.NewRequest(http.MethodGet, target, http.NoBody)
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestListImagesPassesPaginationThroughAndReportsTheRealTotal(t *testing.T) {
	svc := &listImages{resp: &image.ListImagesResponse{
		Images:     []image.Image{{ID: 7, OriginalFilename: "a.png", UploadedAt: time.Now(), Tags: []image.Tag{{Name: "nature"}}}},
		TotalCount: 250, Page: 3, PageSize: 24, TotalPages: 11,
	}}
	rec := serveList(t, svc, "/api/images?page=3&page_size=24&tags=nature&match_all=true", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body %q", rec.Code, rec.Body.String())
	}
	if svc.got.Page != 3 || svc.got.PageSize != 24 || !svc.got.MatchAll || len(svc.got.Tags) != 1 {
		t.Errorf("service got %+v", svc.got)
	}
	var body struct {
		TotalCount int               `json:"total_count"`
		Page       int               `json:"page"`
		PageSize   int               `json:"page_size"`
		TotalPages int               `json:"total_pages"`
		TagColors  map[string]string `json:"tag_colors"`
		Images     []ImageResponse   `json:"images"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalCount != 250 || body.Page != 3 || body.PageSize != 24 || body.TotalPages != 11 {
		t.Errorf("pagination = %+v", body)
	}
	if body.TagColors["nature"] != settings.GetTagColor("nature") || body.TagColors["nature"] == "" {
		t.Errorf("tag_colors = %v", body.TagColors)
	}
	if len(body.Images) != 1 || body.Images[0].ID != "7" {
		t.Errorf("images = %+v", body.Images)
	}
}

func TestListImagesDefaultsAndClampsPagination(t *testing.T) {
	svc := &listImages{resp: &image.ListImagesResponse{}}
	serveList(t, svc, "/api/images", nil)
	if svc.got.Page != 1 || svc.got.PageSize != image.MaxPageSize {
		t.Errorf("defaults: page %d size %d", svc.got.Page, svc.got.PageSize)
	}
	serveList(t, svc, "/api/images?page_size=5000", nil)
	if svc.got.PageSize != image.MaxPageSize {
		t.Errorf("page_size 5000 reached the service as %d", svc.got.PageSize)
	}
}

func TestListImagesRejectsANonNumericPage(t *testing.T) {
	for _, q := range []string{"page=abc", "page=0", "page_size=-1"} {
		rec := serveList(t, &listImages{resp: &image.ListImagesResponse{}}, "/api/images?"+q, nil)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", q, rec.Code)
		}
	}
}

func TestListImagesAnswersJSONToHTMXRequests(t *testing.T) {
	svc := &listImages{resp: &image.ListImagesResponse{
		Images: []image.Image{{ID: 1, OriginalFilename: `x'),alert(1),('.png`}}, TotalCount: 1, Page: 1, PageSize: 100, TotalPages: 1,
	}}
	rec := serveList(t, svc, "/api/images?filtered=true", map[string]string{"HX-Request": "true"})
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q", ct)
	}
	if strings.Contains(rec.Body.String(), "onclick") {
		t.Errorf("response still carries markup: %s", rec.Body.String())
	}
}

func TestListImagesColorsTheActiveFiltersEvenWithoutMatches(t *testing.T) {
	rec := serveList(t, &listImages{resp: &image.ListImagesResponse{}}, "/api/images?tags=coast", nil)
	var body struct {
		TagColors map[string]string `json:"tag_colors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TagColors["coast"] == "" {
		t.Errorf("tag_colors = %v, want a color for the coast filter", body.TagColors)
	}
}
