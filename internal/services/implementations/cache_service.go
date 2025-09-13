package implementations

import (
	"context"
	"log"

	"image-gallery/internal/domain/image"
	"image-gallery/internal/platform/cache"
)

// CacheService implements the domain CacheService interface using Redis/Valkey
type CacheService struct {
	client *cache.RedisClient
}

// NewCacheService creates a new cache service
func NewCacheService(client *cache.RedisClient) *CacheService {
	return &CacheService{
		client: client,
	}
}

// GetImage retrieves a cached image
func (c *CacheService) GetImage(ctx context.Context, id int) (*image.Image, error) {
	if c.client == nil {
		return nil, image.ErrCacheUnavailable
	}
	
	return c.client.GetImage(ctx, id)
}

// SetImage caches an image
func (c *CacheService) SetImage(ctx context.Context, img *image.Image, expiry int64) error {
	if c.client == nil {
		log.Printf("Cache unavailable, skipping image cache for ID %d", img.ID)
		return nil // Don't fail if cache is unavailable
	}
	
	return c.client.SetImage(ctx, img, expiry)
}

// DeleteImage removes an image from cache
func (c *CacheService) DeleteImage(ctx context.Context, id int) error {
	if c.client == nil {
		return nil // Don't fail if cache is unavailable
	}
	
	return c.client.DeleteImage(ctx, id)
}

// GetImageList retrieves a cached image list
func (c *CacheService) GetImageList(ctx context.Context, key string) (*image.ListImagesResponse, error) {
	if c.client == nil {
		return nil, image.ErrCacheUnavailable
	}
	
	return c.client.GetImageList(ctx, key)
}

// SetImageList caches an image list
func (c *CacheService) SetImageList(ctx context.Context, key string, response *image.ListImagesResponse, expiry int64) error {
	if c.client == nil {
		log.Printf("Cache unavailable, skipping image list cache for key %s", key)
		return nil // Don't fail if cache is unavailable
	}
	
	return c.client.SetImageList(ctx, key, response, expiry)
}

// InvalidateImageLists clears cached image lists
func (c *CacheService) InvalidateImageLists(ctx context.Context) error {
	if c.client == nil {
		return nil // Don't fail if cache is unavailable
	}
	
	return c.client.InvalidateImageLists(ctx)
}

// GetStats retrieves cached statistics
func (c *CacheService) GetStats(ctx context.Context, key string) (interface{}, error) {
	if c.client == nil {
		return nil, image.ErrCacheUnavailable
	}
	
	return c.client.GetStats(ctx, key)
}

// SetStats caches statistics
func (c *CacheService) SetStats(ctx context.Context, key string, stats interface{}, expiry int64) error {
	if c.client == nil {
		log.Printf("Cache unavailable, skipping stats cache for key %s", key)
		return nil // Don't fail if cache is unavailable
	}
	
	return c.client.SetStats(ctx, key, stats, expiry)
}

// Health checks if the cache service is healthy
func (c *CacheService) Health(ctx context.Context) error {
	if c.client == nil {
		return image.ErrCacheUnavailable
	}
	
	return c.client.Health(ctx)
}