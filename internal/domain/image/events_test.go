package image

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for repeated string literals
const (
	testUserID    = "user123"
	testIPAddress = "192.168.1.1"
	testFilename  = "test.jpg"
	testTagNature = "nature"
)

func TestNewImageCreatedEvent(t *testing.T) {
	// Create test image
	testTime := time.Now()
	image := &Image{
		ID:               1,
		Filename:         testFilename,
		OriginalFilename: "original_test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/test.jpg",
		Width:            intPtr(800),
		Height:           intPtr(600),
		Tags: []Tag{
			{Name: testTagNature},
			{Name: "landscape"},
		},
		Metadata:   json.RawMessage(`{"camera": "Canon"}`),
		UploadedAt: testTime,
		CreatedAt:  testTime,
		UpdatedAt:  testTime,
	}

	userID := testUserID
	event := NewImageCreatedEvent(image, userID)

	assert.Equal(t, image.ID, event.ImageID)
	assert.Equal(t, image.Filename, event.Filename)
	assert.Equal(t, image.OriginalFilename, event.OriginalFilename)
	assert.Equal(t, image.ContentType, event.ContentType)
	assert.Equal(t, image.FileSize, event.FileSize)
	assert.Equal(t, image.StoragePath, event.StoragePath)
	assert.Equal(t, image.Width, event.Width)
	assert.Equal(t, image.Height, event.Height)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, []string{testTagNature, "landscape"}, event.Tags)
	assert.NotNil(t, event.Metadata)
	assert.False(t, event.Timestamp.IsZero())

	// Check metadata was parsed correctly
	camera, exists := event.Metadata["camera"]
	assert.True(t, exists)
	assert.Equal(t, "Canon", camera)
}

func TestNewImageCreatedEvent_EmptyMetadata(t *testing.T) {
	image := &Image{
		ID:               1,
		Filename:         testFilename,
		OriginalFilename: testFilename,
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/test.jpg",
	}

	event := NewImageCreatedEvent(image, testUserID)
	assert.Nil(t, event.Metadata)
}

func TestNewImageUpdatedEvent(t *testing.T) {
	imageID := 1
	updatedBy := testUserID
	changes := map[string]interface{}{
		"tags": []string{"new-tag"},
	}
	previousData := map[string]interface{}{
		"tags": []string{"old-tag"},
	}
	newData := map[string]interface{}{
		"tags": []string{"new-tag"},
	}

	event := NewImageUpdatedEvent(imageID, updatedBy, changes, previousData, newData)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, updatedBy, event.UpdatedBy)
	assert.Equal(t, changes, event.Changes)
	assert.Equal(t, previousData, event.PreviousData)
	assert.Equal(t, newData, event.NewData)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewImageDeletedEvent(t *testing.T) {
	testTime := time.Now()
	image := &Image{
		ID:               1,
		Filename:         testFilename,
		OriginalFilename: "original_test.jpg",
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/test.jpg",
		Width:            intPtr(800),
		Height:           intPtr(600),
		Tags: []Tag{
			{Name: testTagNature},
		},
		UploadedAt: testTime,
		CreatedAt:  testTime,
	}

	deletedBy := testUserID
	reason := "User requested deletion"
	event := NewImageDeletedEvent(image, deletedBy, reason)

	assert.Equal(t, image.ID, event.ImageID)
	assert.Equal(t, image.Filename, event.Filename)
	assert.Equal(t, image.StoragePath, event.StoragePath)
	assert.Equal(t, deletedBy, event.DeletedBy)
	assert.Equal(t, reason, event.Reason)
	assert.NotNil(t, event.ImageData)
	assert.False(t, event.Timestamp.IsZero())

	// Verify image data contains expected fields
	assert.Equal(t, image.OriginalFilename, event.ImageData["original_filename"])
	assert.Equal(t, image.ContentType, event.ImageData["content_type"])
	assert.Equal(t, image.FileSize, event.ImageData["file_size"])
	assert.Equal(t, []string{testTagNature}, event.ImageData["tags"])
}

func TestNewImageViewedEvent(t *testing.T) {
	imageID := 1
	userID := testUserID
	ipAddress := testIPAddress
	userAgent := "Mozilla/5.0"
	referrer := "https://example.com"

	event := NewImageViewedEvent(imageID, userID, ipAddress, userAgent, referrer)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, userAgent, event.UserAgent)
	assert.Equal(t, referrer, event.Referrer)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewImageDownloadedEvent(t *testing.T) {
	imageID := 1
	filename := testFilename
	userID := testUserID
	ipAddress := testIPAddress
	downloadSize := int64(1024)

	event := NewImageDownloadedEvent(imageID, filename, userID, ipAddress, downloadSize)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, filename, event.Filename)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, ipAddress, event.IPAddress)
	assert.Equal(t, downloadSize, event.DownloadSize)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewTagCreatedEvent(t *testing.T) {
	tag := &Tag{
		ID:        1,
		Name:      testTagNature,
		CreatedAt: time.Now(),
	}
	createdBy := testUserID

	event := NewTagCreatedEvent(tag, createdBy)

	assert.Equal(t, tag.ID, event.TagID)
	assert.Equal(t, tag.Name, event.Name)
	assert.Equal(t, createdBy, event.CreatedBy)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewTagDeletedEvent(t *testing.T) {
	tag := &Tag{
		ID:   1,
		Name: testTagNature,
	}
	deletedBy := testUserID
	reason := "Unused tag"
	imageCount := 5

	event := NewTagDeletedEvent(tag, deletedBy, reason, imageCount)

	assert.Equal(t, tag.ID, event.TagID)
	assert.Equal(t, tag.Name, event.Name)
	assert.Equal(t, deletedBy, event.DeletedBy)
	assert.Equal(t, reason, event.Reason)
	assert.Equal(t, imageCount, event.ImageCount)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewTagAttachedEvent(t *testing.T) {
	imageID := 1
	tagID := 2
	tagName := testTagNature
	attachedBy := testUserID

	event := NewTagAttachedEvent(imageID, tagID, tagName, attachedBy)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, tagID, event.TagID)
	assert.Equal(t, tagName, event.TagName)
	assert.Equal(t, attachedBy, event.AttachedBy)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewTagDetachedEvent(t *testing.T) {
	imageID := 1
	tagID := 2
	tagName := testTagNature
	detachedBy := testUserID

	event := NewTagDetachedEvent(imageID, tagID, tagName, detachedBy)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, tagID, event.TagID)
	assert.Equal(t, tagName, event.TagName)
	assert.Equal(t, detachedBy, event.DetachedBy)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewStorageQuotaReachedEvent(t *testing.T) {
	userID := testUserID
	currentUsage := int64(900 * 1024 * 1024) // 900MB
	quotaLimit := int64(1024 * 1024 * 1024)  // 1GB

	event := NewStorageQuotaReachedEvent(userID, currentUsage, quotaLimit)

	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, currentUsage, event.CurrentUsage)
	assert.Equal(t, quotaLimit, event.QuotaLimit)
	assert.InDelta(t, 87.89, event.PercentageUsed, 0.1) // ~87.89%
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewImageProcessingFailedEvent(t *testing.T) {
	imageID := 1
	filename := testFilename
	processType := "thumbnail"
	errorMsg := "Failed to generate thumbnail"
	context := map[string]interface{}{
		"width":  150,
		"height": 150,
	}

	event := NewImageProcessingFailedEvent(imageID, filename, processType, errorMsg, context)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, filename, event.Filename)
	assert.Equal(t, processType, event.ProcessType)
	assert.Equal(t, errorMsg, event.Error)
	assert.Equal(t, context, event.Context)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewThumbnailGeneratedEvent(t *testing.T) {
	imageID := 1
	thumbnailPath := "/thumbnails/test_thumb.jpg"
	thumbnailSize := int64(512)
	width := 150
	height := 150
	processingTime := int64(250)

	event := NewThumbnailGeneratedEvent(imageID, thumbnailPath, thumbnailSize, width, height, processingTime)

	assert.Equal(t, imageID, event.ImageID)
	assert.Equal(t, thumbnailPath, event.ThumbnailPath)
	assert.Equal(t, thumbnailSize, event.ThumbnailSize)
	assert.Equal(t, width, event.Width)
	assert.Equal(t, height, event.Height)
	assert.Equal(t, processingTime, event.ProcessingTime)
	assert.False(t, event.Timestamp.IsZero())
}

func TestNewDomainEvent(t *testing.T) {
	eventType := EventImageCreated
	aggregateID := "image_1"
	userID := testUserID
	data := map[string]interface{}{
		"filename": testFilename,
		"size":     1024,
	}

	event := NewDomainEvent(eventType, aggregateID, userID, data)

	assert.NotEmpty(t, event.ID)
	assert.Equal(t, eventType, event.Type)
	assert.Equal(t, aggregateID, event.AggregateID)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, 1, event.Version)
	assert.Equal(t, data, event.Data)
	assert.NotNil(t, event.Metadata)
	assert.False(t, event.Timestamp.IsZero())
}

func TestDomainEvent_AddMetadata(t *testing.T) {
	event := NewDomainEvent(EventImageCreated, "image_1", testUserID, nil)

	event.AddMetadata("source", "web")
	event.AddMetadata("client_version", "1.0.0")

	assert.Equal(t, "web", event.Metadata["source"])
	assert.Equal(t, "1.0.0", event.Metadata["client_version"])
}

func TestDomainEvent_ToJSON_FromJSON(t *testing.T) {
	originalEvent := NewDomainEvent(EventImageCreated, "image_1", testUserID, map[string]interface{}{
		"filename": testFilename,
	})
	originalEvent.AddMetadata("source", "web")

	// Serialize to JSON
	jsonData, err := originalEvent.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize from JSON
	var deserializedEvent DomainEvent
	err = deserializedEvent.FromJSON(jsonData)
	require.NoError(t, err)

	assert.Equal(t, originalEvent.ID, deserializedEvent.ID)
	assert.Equal(t, originalEvent.Type, deserializedEvent.Type)
	assert.Equal(t, originalEvent.AggregateID, deserializedEvent.AggregateID)
	assert.Equal(t, originalEvent.UserID, deserializedEvent.UserID)
	assert.Equal(t, originalEvent.Version, deserializedEvent.Version)
	assert.Equal(t, originalEvent.Data, deserializedEvent.Data)
	assert.Equal(t, originalEvent.Metadata, deserializedEvent.Metadata)
}

func TestDomainEvent_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		event    *DomainEvent
		expected bool
	}{
		{
			name: "valid event",
			event: &DomainEvent{
				ID:          "123",
				Type:        EventImageCreated,
				AggregateID: "image_1",
				Timestamp:   time.Now(),
			},
			expected: true,
		},
		{
			name: "empty ID",
			event: &DomainEvent{
				Type:        EventImageCreated,
				AggregateID: "image_1",
				Timestamp:   time.Now(),
			},
			expected: false,
		},
		{
			name: "empty type",
			event: &DomainEvent{
				ID:          "123",
				AggregateID: "image_1",
				Timestamp:   time.Now(),
			},
			expected: false,
		},
		{
			name: "empty aggregate ID",
			event: &DomainEvent{
				ID:        "123",
				Type:      EventImageCreated,
				Timestamp: time.Now(),
			},
			expected: false,
		},
		{
			name: "zero timestamp",
			event: &DomainEvent{
				ID:          "123",
				Type:        EventImageCreated,
				AggregateID: "image_1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.event.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDomainEvent_GetAge(t *testing.T) {
	// Create an event with a timestamp 1 second ago
	event := &DomainEvent{
		Timestamp: time.Now().Add(-1 * time.Second),
	}

	age := event.GetAge()
	assert.GreaterOrEqual(t, age, int64(1000)) // At least 1000ms
	assert.LessOrEqual(t, age, int64(2000))    // Less than 2000ms
}

func TestEventStream_NewEventStream(t *testing.T) {
	stream := NewEventStream()

	assert.NotNil(t, stream)
	assert.Empty(t, stream.Events)
	assert.Equal(t, 0, stream.Total)
	assert.True(t, stream.From.IsZero())
	assert.True(t, stream.To.IsZero())
}

func TestEventStream_AddEvent(t *testing.T) {
	stream := NewEventStream()

	// Add first event
	event1 := DomainEvent{
		ID:          "1",
		Type:        EventImageCreated,
		AggregateID: "image_1",
		Timestamp:   time.Now().Add(-2 * time.Hour),
	}
	stream.AddEvent(event1)

	assert.Len(t, stream.Events, 1)
	assert.Equal(t, 1, stream.Total)
	assert.Equal(t, event1.Timestamp, stream.From)
	assert.Equal(t, event1.Timestamp, stream.To)

	// Add second event (newer)
	event2 := DomainEvent{
		ID:          "2",
		Type:        EventImageUpdated,
		AggregateID: "image_1",
		Timestamp:   time.Now().Add(-1 * time.Hour),
	}
	stream.AddEvent(event2)

	assert.Len(t, stream.Events, 2)
	assert.Equal(t, 2, stream.Total)
	assert.Equal(t, event1.Timestamp, stream.From) // Should remain the earlier time
	assert.Equal(t, event2.Timestamp, stream.To)   // Should be updated to the later time
}

func TestEventStream_FilterByType(t *testing.T) {
	stream := NewEventStream()

	events := []DomainEvent{
		{Type: EventImageCreated, AggregateID: "image_1"},
		{Type: EventImageUpdated, AggregateID: "image_1"},
		{Type: EventImageCreated, AggregateID: "image_2"},
		{Type: EventTagCreated, AggregateID: "tag_1"},
	}

	for _, event := range events {
		stream.AddEvent(event)
	}

	createdEvents := stream.FilterByType(EventImageCreated)
	assert.Len(t, createdEvents, 2)
	for _, event := range createdEvents {
		assert.Equal(t, EventImageCreated, event.Type)
	}

	updatedEvents := stream.FilterByType(EventImageUpdated)
	assert.Len(t, updatedEvents, 1)

	tagEvents := stream.FilterByType(EventTagCreated)
	assert.Len(t, tagEvents, 1)
}

func TestEventStream_FilterByAggregateID(t *testing.T) {
	stream := NewEventStream()

	events := []DomainEvent{
		{Type: EventImageCreated, AggregateID: "image_1"},
		{Type: EventImageUpdated, AggregateID: "image_1"},
		{Type: EventImageCreated, AggregateID: "image_2"},
		{Type: EventTagCreated, AggregateID: "tag_1"},
	}

	for _, event := range events {
		stream.AddEvent(event)
	}

	image1Events := stream.FilterByAggregateID("image_1")
	assert.Len(t, image1Events, 2)
	for _, event := range image1Events {
		assert.Equal(t, "image_1", event.AggregateID)
	}

	image2Events := stream.FilterByAggregateID("image_2")
	assert.Len(t, image2Events, 1)

	tagEvents := stream.FilterByAggregateID("tag_1")
	assert.Len(t, tagEvents, 1)
}

func TestEventStream_FilterByTimeRange(t *testing.T) {
	stream := NewEventStream()

	now := time.Now()
	events := []DomainEvent{
		{Type: EventImageCreated, Timestamp: now.Add(-3 * time.Hour)},
		{Type: EventImageUpdated, Timestamp: now.Add(-2 * time.Hour)},
		{Type: EventImageDeleted, Timestamp: now.Add(-1 * time.Hour)},
		{Type: EventTagCreated, Timestamp: now},
	}

	for _, event := range events {
		stream.AddEvent(event)
	}

	// Filter events from 2.5 hours ago to 30 minutes ago
	from := now.Add(-150 * time.Minute) // 2.5 hours ago
	to := now.Add(-30 * time.Minute)    // 30 minutes ago

	filteredEvents := stream.FilterByTimeRange(from, to)
	assert.Len(t, filteredEvents, 2) // Should include the -2h and -1h events
}

func TestEventTypes(t *testing.T) {
	// Test that event types are defined correctly
	eventTypes := []EventType{
		EventImageCreated,
		EventImageUpdated,
		EventImageDeleted,
		EventImageViewed,
		EventImageDownloaded,
		EventTagCreated,
		EventTagDeleted,
		EventTagAttached,
		EventTagDetached,
		EventStorageQuotaReached,
		EventImageProcessingFailed,
		EventThumbnailGenerated,
	}

	for _, eventType := range eventTypes {
		assert.NotEmpty(t, string(eventType))
	}
}

// Benchmark tests

func BenchmarkNewImageCreatedEvent(b *testing.B) {
	image := &Image{
		ID:               1,
		Filename:         testFilename,
		OriginalFilename: testFilename,
		ContentType:      "image/jpeg",
		FileSize:         1024,
		StoragePath:      "/storage/test.jpg",
		Tags:             []Tag{{Name: testTagNature}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewImageCreatedEvent(image, testUserID)
	}
}

func BenchmarkDomainEvent_ToJSON(b *testing.B) {
	event := NewDomainEvent(EventImageCreated, "image_1", testUserID, map[string]interface{}{
		"filename": testFilename,
		"size":     1024,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = event.ToJSON()
	}
}

func BenchmarkEventStream_FilterByType(b *testing.B) {
	stream := NewEventStream()

	// Add many events
	for i := 0; i < 1000; i++ {
		event := DomainEvent{
			Type:        EventImageCreated,
			AggregateID: "image_" + string(rune(i)),
		}
		stream.AddEvent(event)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stream.FilterByType(EventImageCreated)
	}
}
