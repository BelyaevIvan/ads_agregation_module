package service

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"

	"brandhunt/api-service/internal/middleware"
	"brandhunt/api-service/internal/repository"

	"github.com/minio/minio-go/v7"
)

type PhotoService struct {
	photos   *repository.PhotoRepo
	listings *repository.ListingRepo
	minio    *minio.Client
	bucket   string
}

func NewPhotoService(photos *repository.PhotoRepo, listings *repository.ListingRepo, minioClient *minio.Client, bucket string) *PhotoService {
	return &PhotoService{photos: photos, listings: listings, minio: minioClient, bucket: bucket}
}

func (s *PhotoService) DeletePhoto(listingID, photoID string) (*string, error) {
	if !IsValidUUID(listingID) || !IsValidUUID(photoID) {
		return nil, middleware.NewAppError(http.StatusBadRequest, "некорректный формат UUID")
	}

	exists, err := s.listings.Exists(listingID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, middleware.NewAppError(http.StatusNotFound, "объявление не найдено")
	}

	photo, err := s.photos.GetByIDAndListing(photoID, listingID)
	if err != nil {
		return nil, err
	}
	if photo == nil {
		return nil, middleware.NewAppError(http.StatusNotFound, "фотография не найдена")
	}

	wasCover := photo.IsCover
	photoURL := photo.URL

	if err := s.photos.Delete(photoID); err != nil {
		return nil, err
	}

	var newCoverID *string
	if wasCover {
		newCoverID, err = s.photos.PromoteNextCover(listingID)
		if err != nil {
			return nil, err
		}
	}

	// Delete from MinIO (best-effort)
	s.deleteFromMinio(photoURL)

	return newCoverID, nil
}

func (s *PhotoService) deleteFromMinio(photoURL string) {
	if s.minio == nil {
		return
	}
	// Extract object name from URL: http://minio:9000/bucket/object/path.jpg -> object/path.jpg
	parsed, err := url.Parse(photoURL)
	if err != nil {
		log.Printf("WARN: не удалось разобрать URL фото для удаления из MinIO: %s", photoURL)
		return
	}
	// Path: /bucket/object_name
	path := strings.TrimPrefix(parsed.Path, "/"+s.bucket+"/")
	if path == "" || path == parsed.Path {
		log.Printf("WARN: не удалось извлечь object name из URL: %s", photoURL)
		return
	}
	if err := s.minio.RemoveObject(context.Background(), s.bucket, path, minio.RemoveObjectOptions{}); err != nil {
		log.Printf("WARN: ошибка удаления фото из MinIO (%s): %v", path, err)
	}
}
