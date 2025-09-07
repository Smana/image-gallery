package image

import (
	"encoding/json"
	"time"
)

// EventType represents the type of domain event
type EventType string

const (
	// Image events
	EventImageCreated   EventType = "image.created"
	EventImageUpdated   EventType = "image.updated"
	EventImageDeleted   EventType = "image.deleted"
	EventImageViewed    EventType = "image.viewed"
	EventImageDownloaded EventType = "image.downloaded"

	// Tag events
	EventTagCreated EventType = "tag.created"
	EventTagDeleted EventType = "tag.deleted"
	EventTagAttached EventType = "tag.attached"
	EventTagDetached EventType = "tag.detached"

	// System events
	EventStorageQuotaReached EventType = "storage.quota_reached"
	EventImageProcessingFailed EventType = "image.processing_failed"
	EventThumbnailGenerated EventType = "image.thumbnail_generated"
)

// DomainEvent represents a domain event in the image gallery system
type DomainEvent struct {
	ID          string                 `json:"id"`
	Type        EventType              `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	UserID      string                 `json:"user_id,omitempty"`
	Version     int                    `json:"version"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ImageCreatedEvent is published when a new image is created
type ImageCreatedEvent struct {
	ImageID          int                    `json:"image_id"`
	Filename         string                 `json:"filename"`
	OriginalFilename string                 `json:"original_filename"`
	ContentType      string                 `json:"content_type"`
	FileSize         int64                  `json:"file_size"`
	StoragePath      string                 `json:"storage_path"`
	Width            *int                   `json:"width,omitempty"`
	Height           *int                   `json:"height,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	UserID           string                 `json:"user_id"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	Timestamp        time.Time              `json:"timestamp"`
}

// ImageUpdatedEvent is published when an image is updated
type ImageUpdatedEvent struct {
	ImageID      int                    `json:"image_id"`
	UpdatedBy    string                 `json:"updated_by"`
	Changes      map[string]interface{} `json:"changes"`
	PreviousData map[string]interface{} `json:"previous_data"`
	NewData      map[string]interface{} `json:"new_data"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ImageDeletedEvent is published when an image is deleted
type ImageDeletedEvent struct {
	ImageID     int                    `json:"image_id"`
	Filename    string                 `json:"filename"`
	StoragePath string                 `json:"storage_path"`
	DeletedBy   string                 `json:"deleted_by"`
	Reason      string                 `json:"reason,omitempty"`
	ImageData   map[string]interface{} `json:"image_data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ImageViewedEvent is published when an image is viewed
type ImageViewedEvent struct {
	ImageID   int       `json:"image_id"`
	UserID    string    `json:"user_id,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Referrer  string    `json:"referrer,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ImageDownloadedEvent is published when an image is downloaded
type ImageDownloadedEvent struct {
	ImageID      int    `json:"image_id"`
	Filename     string `json:"filename"`
	UserID       string `json:"user_id,omitempty"`
	IPAddress    string `json:"ip_address,omitempty"`
	DownloadSize int64  `json:"download_size"`
	Timestamp    time.Time `json:"timestamp"`
}

// TagCreatedEvent is published when a new tag is created
type TagCreatedEvent struct {
	TagID     int       `json:"tag_id"`
	Name      string    `json:"name"`
	CreatedBy string    `json:"created_by"`
	Timestamp time.Time `json:"timestamp"`
}

// TagDeletedEvent is published when a tag is deleted
type TagDeletedEvent struct {
	TagID     int       `json:"tag_id"`
	Name      string    `json:"name"`
	DeletedBy string    `json:"deleted_by"`
	Reason    string    `json:"reason,omitempty"`
	ImageCount int      `json:"image_count"` // Number of images that had this tag
	Timestamp time.Time `json:"timestamp"`
}

// TagAttachedEvent is published when a tag is attached to an image
type TagAttachedEvent struct {
	ImageID   int       `json:"image_id"`
	TagID     int       `json:"tag_id"`
	TagName   string    `json:"tag_name"`
	AttachedBy string   `json:"attached_by"`
	Timestamp time.Time `json:"timestamp"`
}

// TagDetachedEvent is published when a tag is detached from an image
type TagDetachedEvent struct {
	ImageID    int       `json:"image_id"`
	TagID      int       `json:"tag_id"`
	TagName    string    `json:"tag_name"`
	DetachedBy string    `json:"detached_by"`
	Timestamp  time.Time `json:"timestamp"`
}

// StorageQuotaReachedEvent is published when storage quota is reached
type StorageQuotaReachedEvent struct {
	UserID         string    `json:"user_id"`
	CurrentUsage   int64     `json:"current_usage"`
	QuotaLimit     int64     `json:"quota_limit"`
	PercentageUsed float64   `json:"percentage_used"`
	Timestamp      time.Time `json:"timestamp"`
}

// ImageProcessingFailedEvent is published when image processing fails
type ImageProcessingFailedEvent struct {
	ImageID     int                    `json:"image_id,omitempty"`
	Filename    string                 `json:"filename"`
	ProcessType string                 `json:"process_type"` // thumbnail, resize, optimize, etc.
	Error       string                 `json:"error"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ThumbnailGeneratedEvent is published when a thumbnail is successfully generated
type ThumbnailGeneratedEvent struct {
	ImageID       int       `json:"image_id"`
	ThumbnailPath string    `json:"thumbnail_path"`
	ThumbnailSize int64     `json:"thumbnail_size"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	ProcessingTime int64    `json:"processing_time_ms"`
	Timestamp     time.Time `json:"timestamp"`
}

// Event factory functions

// NewImageCreatedEvent creates a new ImageCreatedEvent
func NewImageCreatedEvent(image *Image, userID string) *ImageCreatedEvent {
	tagNames := make([]string, len(image.Tags))
	for i, tag := range image.Tags {
		tagNames[i] = tag.Name
	}

	var metadata map[string]interface{}
	if len(image.Metadata) > 0 {
		json.Unmarshal(image.Metadata, &metadata)
	}

	return &ImageCreatedEvent{
		ImageID:          image.ID,
		Filename:         image.Filename,
		OriginalFilename: image.OriginalFilename,
		ContentType:      image.ContentType,
		FileSize:         image.FileSize,
		StoragePath:      image.StoragePath,
		Width:            image.Width,
		Height:           image.Height,
		Tags:             tagNames,
		UserID:           userID,
		Metadata:         metadata,
		Timestamp:        time.Now(),
	}
}

// NewImageUpdatedEvent creates a new ImageUpdatedEvent
func NewImageUpdatedEvent(imageID int, updatedBy string, changes, previousData, newData map[string]interface{}) *ImageUpdatedEvent {
	return &ImageUpdatedEvent{
		ImageID:      imageID,
		UpdatedBy:    updatedBy,
		Changes:      changes,
		PreviousData: previousData,
		NewData:      newData,
		Timestamp:    time.Now(),
	}
}

// NewImageDeletedEvent creates a new ImageDeletedEvent
func NewImageDeletedEvent(image *Image, deletedBy, reason string) *ImageDeletedEvent {
	imageData := map[string]interface{}{
		"original_filename": image.OriginalFilename,
		"content_type":      image.ContentType,
		"file_size":         image.FileSize,
		"width":             image.Width,
		"height":            image.Height,
		"uploaded_at":       image.UploadedAt,
		"created_at":        image.CreatedAt,
	}

	if len(image.Tags) > 0 {
		tagNames := make([]string, len(image.Tags))
		for i, tag := range image.Tags {
			tagNames[i] = tag.Name
		}
		imageData["tags"] = tagNames
	}

	return &ImageDeletedEvent{
		ImageID:     image.ID,
		Filename:    image.Filename,
		StoragePath: image.StoragePath,
		DeletedBy:   deletedBy,
		Reason:      reason,
		ImageData:   imageData,
		Timestamp:   time.Now(),
	}
}

// NewImageViewedEvent creates a new ImageViewedEvent
func NewImageViewedEvent(imageID int, userID, ipAddress, userAgent, referrer string) *ImageViewedEvent {
	return &ImageViewedEvent{
		ImageID:   imageID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Referrer:  referrer,
		Timestamp: time.Now(),
	}
}

// NewImageDownloadedEvent creates a new ImageDownloadedEvent
func NewImageDownloadedEvent(imageID int, filename string, userID, ipAddress string, downloadSize int64) *ImageDownloadedEvent {
	return &ImageDownloadedEvent{
		ImageID:      imageID,
		Filename:     filename,
		UserID:       userID,
		IPAddress:    ipAddress,
		DownloadSize: downloadSize,
		Timestamp:    time.Now(),
	}
}

// NewTagCreatedEvent creates a new TagCreatedEvent
func NewTagCreatedEvent(tag *Tag, createdBy string) *TagCreatedEvent {
	return &TagCreatedEvent{
		TagID:     tag.ID,
		Name:      tag.Name,
		CreatedBy: createdBy,
		Timestamp: time.Now(),
	}
}

// NewTagDeletedEvent creates a new TagDeletedEvent
func NewTagDeletedEvent(tag *Tag, deletedBy, reason string, imageCount int) *TagDeletedEvent {
	return &TagDeletedEvent{
		TagID:      tag.ID,
		Name:       tag.Name,
		DeletedBy:  deletedBy,
		Reason:     reason,
		ImageCount: imageCount,
		Timestamp:  time.Now(),
	}
}

// NewTagAttachedEvent creates a new TagAttachedEvent
func NewTagAttachedEvent(imageID, tagID int, tagName, attachedBy string) *TagAttachedEvent {
	return &TagAttachedEvent{
		ImageID:    imageID,
		TagID:      tagID,
		TagName:    tagName,
		AttachedBy: attachedBy,
		Timestamp:  time.Now(),
	}
}

// NewTagDetachedEvent creates a new TagDetachedEvent
func NewTagDetachedEvent(imageID, tagID int, tagName, detachedBy string) *TagDetachedEvent {
	return &TagDetachedEvent{
		ImageID:    imageID,
		TagID:      tagID,
		TagName:    tagName,
		DetachedBy: detachedBy,
		Timestamp:  time.Now(),
	}
}

// NewStorageQuotaReachedEvent creates a new StorageQuotaReachedEvent
func NewStorageQuotaReachedEvent(userID string, currentUsage, quotaLimit int64) *StorageQuotaReachedEvent {
	percentageUsed := float64(currentUsage) / float64(quotaLimit) * 100

	return &StorageQuotaReachedEvent{
		UserID:         userID,
		CurrentUsage:   currentUsage,
		QuotaLimit:     quotaLimit,
		PercentageUsed: percentageUsed,
		Timestamp:      time.Now(),
	}
}

// NewImageProcessingFailedEvent creates a new ImageProcessingFailedEvent
func NewImageProcessingFailedEvent(imageID int, filename, processType, errorMsg string, context map[string]interface{}) *ImageProcessingFailedEvent {
	return &ImageProcessingFailedEvent{
		ImageID:     imageID,
		Filename:    filename,
		ProcessType: processType,
		Error:       errorMsg,
		Context:     context,
		Timestamp:   time.Now(),
	}
}

// NewThumbnailGeneratedEvent creates a new ThumbnailGeneratedEvent
func NewThumbnailGeneratedEvent(imageID int, thumbnailPath string, thumbnailSize int64, width, height int, processingTime int64) *ThumbnailGeneratedEvent {
	return &ThumbnailGeneratedEvent{
		ImageID:        imageID,
		ThumbnailPath:  thumbnailPath,
		ThumbnailSize:  thumbnailSize,
		Width:          width,
		Height:         height,
		ProcessingTime: processingTime,
		Timestamp:      time.Now(),
	}
}

// Helper methods for DomainEvent

// NewDomainEvent creates a new domain event
func NewDomainEvent(eventType EventType, aggregateID, userID string, data map[string]interface{}) *DomainEvent {
	return &DomainEvent{
		ID:          generateEventID(),
		Type:        eventType,
		AggregateID: aggregateID,
		UserID:      userID,
		Version:     1,
		Data:        data,
		Metadata:    make(map[string]interface{}),
		Timestamp:   time.Now(),
	}
}

// AddMetadata adds metadata to the event
func (e *DomainEvent) AddMetadata(key string, value interface{}) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
}

// ToJSON serializes the event to JSON
func (e *DomainEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON deserializes the event from JSON
func (e *DomainEvent) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// IsValid checks if the event is valid
func (e *DomainEvent) IsValid() bool {
	return e.ID != "" && 
		   e.Type != "" && 
		   e.AggregateID != "" && 
		   !e.Timestamp.IsZero()
}

// GetAge returns the age of the event in seconds
func (e *DomainEvent) GetAge() int64 {
	return time.Since(e.Timestamp).Milliseconds()
}

// Helper function to generate event IDs (in a real system, you'd use UUID)
func generateEventID() string {
	return time.Now().Format("20060102150405") + "_" + time.Now().Format("000000000")
}

// EventStream represents a stream of domain events
type EventStream struct {
	Events []DomainEvent `json:"events"`
	From   time.Time     `json:"from"`
	To     time.Time     `json:"to"`
	Total  int           `json:"total"`
}

// NewEventStream creates a new event stream
func NewEventStream() *EventStream {
	return &EventStream{
		Events: make([]DomainEvent, 0),
	}
}

// AddEvent adds an event to the stream
func (es *EventStream) AddEvent(event DomainEvent) {
	es.Events = append(es.Events, event)
	es.Total = len(es.Events)
	
	if es.From.IsZero() || event.Timestamp.Before(es.From) {
		es.From = event.Timestamp
	}
	
	if es.To.IsZero() || event.Timestamp.After(es.To) {
		es.To = event.Timestamp
	}
}

// FilterByType filters events by type
func (es *EventStream) FilterByType(eventType EventType) []DomainEvent {
	var filtered []DomainEvent
	for _, event := range es.Events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterByAggregateID filters events by aggregate ID
func (es *EventStream) FilterByAggregateID(aggregateID string) []DomainEvent {
	var filtered []DomainEvent
	for _, event := range es.Events {
		if event.AggregateID == aggregateID {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// FilterByTimeRange filters events by time range
func (es *EventStream) FilterByTimeRange(from, to time.Time) []DomainEvent {
	var filtered []DomainEvent
	for _, event := range es.Events {
		if event.Timestamp.After(from) && event.Timestamp.Before(to) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}