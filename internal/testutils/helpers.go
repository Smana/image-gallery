package testutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	
	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/database"
)

// TestSuite provides common test utilities for integration tests
type TestSuite struct {
	Containers *TestContainers
	Repos      *database.Repositories
}

// SetupTestSuite initializes a complete test suite with containers and repositories
func SetupTestSuite(ctx context.Context) (*TestSuite, error) {
	containers, err := SetupTestContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to setup test containers: %w", err)
	}

	// Initialize repositories
	repos := database.NewRepositories(
		database.NewImageRepository(containers.DB),
		database.NewTagRepository(containers.DB),
		database.NewAlbumRepository(containers.DB),
	)

	return &TestSuite{
		Containers: containers,
		Repos:      repos,
	}, nil
}

// Cleanup cleans up all test resources
func (ts *TestSuite) Cleanup(ctx context.Context) error {
	return ts.Containers.Cleanup(ctx)
}

// ResetData clears all test data and resets the environment
func (ts *TestSuite) ResetData(ctx context.Context) error {
	return ts.Containers.ResetDatabase(ctx)
}

// CreateTestImage creates a test image record in the database
func (ts *TestSuite) CreateTestImage(ctx context.Context, filename string) (*database.Image, error) {
	img := &database.Image{
		Filename:         fmt.Sprintf("test_%s_%d.jpg", filename, rand.Int()),
		OriginalFilename: filename + ".jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024 * rand.Int63n(100), // Random size up to 100KB
		StoragePath:      fmt.Sprintf("images/test_%d.jpg", time.Now().UnixNano()),
		Width:            intPtr(800 + rand.Intn(400)),
		Height:           intPtr(600 + rand.Intn(400)),
		Metadata:         database.Metadata{},
	}

	if err := ts.Repos.Images.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("failed to create test image: %w", err)
	}

	return img, nil
}

// CreateTestTag creates a test tag in the database
func (ts *TestSuite) CreateTestTag(ctx context.Context, name string) (*database.Tag, error) {
	tag := &database.Tag{
		Name:        name,
		Description: stringPtr(fmt.Sprintf("Test tag: %s", name)),
		Color:       stringPtr("#" + fmt.Sprintf("%06x", rand.Intn(0xFFFFFF))),
	}

	if err := ts.Repos.Tags.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("failed to create test tag: %w", err)
	}

	return tag, nil
}

// CreateTestAlbum creates a test album in the database
func (ts *TestSuite) CreateTestAlbum(ctx context.Context, name string) (*database.Album, error) {
	album := &database.Album{
		Name:        name,
		Description: stringPtr(fmt.Sprintf("Test album: %s", name)),
		IsPublic:    rand.Intn(2) == 1,
	}

	if err := ts.Repos.Albums.Create(ctx, album); err != nil {
		return nil, fmt.Errorf("failed to create test album: %w", err)
	}

	return album, nil
}

// CreateTestImageWithTags creates a test image with associated tags
func (ts *TestSuite) CreateTestImageWithTags(ctx context.Context, filename string, tagNames []string) (*database.Image, []*database.Tag, error) {
	img, err := ts.CreateTestImage(ctx, filename)
	if err != nil {
		return nil, nil, err
	}

	var tags []*database.Tag
	for _, tagName := range tagNames {
		tag, err := ts.CreateTestTag(ctx, tagName)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create tag %s: %w", tagName, err)
		}

		// Associate tag with image
		if err := ts.Repos.Tags.AddToImage(ctx, img.ID, tag.ID); err != nil {
			return nil, nil, fmt.Errorf("failed to associate tag %s with image: %w", tagName, err)
		}

		tags = append(tags, tag)
	}

	return img, tags, nil
}

// GenerateTestImageData creates mock image data for testing
func GenerateTestImageData(width, height int) []byte {
	// Create a simple test image (just some bytes that represent an image)
	// In a real test, you might want to generate actual image data
	size := width * height * 3 // RGB
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	return data
}

// CreateMultipartFormData creates multipart form data for file upload testing
func CreateMultipartFormData(filename string, data []byte, additionalFields map[string]string) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	fileWriter, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, "", err
	}
	
	if _, err := fileWriter.Write(data); err != nil {
		return nil, "", err
	}

	// Add additional fields
	for key, value := range additionalFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), writer.FormDataContentType(), nil
}

// MakeTestRequest creates an HTTP test request with the given parameters
func MakeTestRequest(method, url string, body io.Reader, headers map[string]string) *http.Request {
	req := httptest.NewRequest(method, url, body)
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	return req
}

// MakeJSONRequest creates an HTTP test request with JSON body
func MakeJSONRequest(method, url string, payload interface{}) *http.Request {
	var body io.Reader
	if payload != nil {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewReader(jsonData)
	}
	
	req := httptest.NewRequest(method, url, body)
	req.Header.Set("Content-Type", "application/json")
	
	return req
}

// AssertHTTPStatus checks if the HTTP response has the expected status code
func AssertHTTPStatus(t TestingInterface, resp *httptest.ResponseRecorder, expectedStatus int) {
	if resp.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Body: %s", expectedStatus, resp.Code, resp.Body.String())
	}
}

// AssertJSONResponse checks if the response contains valid JSON and optionally validates structure
func AssertJSONResponse(t TestingInterface, resp *httptest.ResponseRecorder, target interface{}) error {
	if !strings.Contains(resp.Header().Get("Content-Type"), "application/json") {
		t.Errorf("Expected JSON response, got %s", resp.Header().Get("Content-Type"))
		return fmt.Errorf("not a JSON response")
	}
	
	if target != nil {
		if err := json.Unmarshal(resp.Body.Bytes(), target); err != nil {
			t.Errorf("Failed to unmarshal JSON response: %v", err)
			return err
		}
	}
	
	return nil
}

// TestingInterface defines the interface for testing frameworks (compatible with testing.T)
type TestingInterface interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Helper()
}

// Utility functions

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// WaitForContainer waits for a container to be ready with timeout
func WaitForContainer(ctx context.Context, container testcontainers.Container, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for container")
		case <-ticker.C:
			state, err := container.State(ctx)
			if err != nil {
				continue
			}
			if state.Running {
				return nil
			}
		}
	}
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomEmail generates a random email address for testing
func RandomEmail() string {
	return fmt.Sprintf("%s@test.com", RandomString(8))
}

// CreateImageRequest creates a test image creation request
func CreateImageRequest(filename, contentType string, tags []string) *image.CreateImageRequest {
	return &image.CreateImageRequest{
		OriginalFilename: filename,
		ContentType:      contentType,
		FileSize:         1024 * rand.Int63n(100),
		Width:            intPtr(800),
		Height:           intPtr(600),
		Tags:             tags,
	}
}