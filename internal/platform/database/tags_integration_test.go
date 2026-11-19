package database_test

import (
	"context"
	"testing"

	"image-gallery/internal/platform/database"
	"image-gallery/internal/testutils"
)

func TestGetByTagsLoadsTagsAndCountsMatches(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	suite, err := testutils.SetupTestSuite(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = suite.Cleanup(ctx) }()
	repo := suite.Repos.Images

	tagIDs := map[string]int{}
	for _, name := range []string{"forest", "peak", "city"} {
		tag, err := suite.CreateTestTag(ctx, "it-"+name)
		if err != nil {
			t.Fatal(err)
		}
		tagIDs[name] = tag.ID
	}
	for file, tags := range map[string][]string{"a.png": {"forest", "peak"}, "b.png": {"forest"}, "c.png": {"city"}} {
		img, err := suite.CreateTestImage(ctx, file)
		if err != nil {
			t.Fatal(err)
		}
		for _, tag := range tags {
			if err := suite.Repos.Tags.AddToImage(ctx, img.ID, tagIDs[tag]); err != nil {
				t.Fatal(err)
			}
		}
	}

	page := database.PaginationParams{Limit: 10}
	matched, err := repo.GetByTags(ctx, []string{"it-forest"}, false, page)
	if err != nil || len(matched) != 2 {
		t.Fatalf("any it-forest: %d images, err %v", len(matched), err)
	}
	for _, img := range matched {
		if len(img.Tags) == 0 {
			t.Errorf("%s came back without its tags", img.OriginalFilename)
		}
	}

	for _, tc := range []struct {
		tags     []string
		matchAll bool
		want     int
	}{
		{[]string{"it-forest"}, false, 2},
		{[]string{"it-forest", "it-city"}, false, 3},
		{[]string{"it-forest", "it-peak"}, true, 1},
		{[]string{"it-forest", "it-forest"}, true, 2},
		{[]string{"it-forest", "it-city"}, true, 0},
	} {
		got, err := repo.CountByTags(ctx, tc.tags, tc.matchAll)
		if err != nil || got != tc.want {
			t.Errorf("CountByTags(%v, all=%v) = %d, %v; want %d", tc.tags, tc.matchAll, got, err, tc.want)
		}
	}
}
