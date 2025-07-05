package aws

import (
	"bytes"
	"context"
	"excalidraw-complete/core"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/oklog/ulid/v2"
)

type s3Store struct {
	s3Client *s3.Client
	bucket   string
}

// NewStore creates a new S3-based store.
func NewStore(bucketName string) *s3Store {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	return &s3Store{
		s3Client: s3Client,
		bucket:   bucketName,
	}
}

// DocumentStore implementation for anonymous sharing
func (s *s3Store) FindID(ctx context.Context, id string) (*core.Document, error) {
	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(id),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document with id %s: %v", id, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read document data: %v", err)
	}

	document := core.Document{
		Data: *bytes.NewBuffer(data),
	}

	return &document, nil
}

func (s *s3Store) Create(ctx context.Context, document *core.Document) (string, error) {
	id := ulid.Make().String()

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(id),
		Body:   bytes.NewReader(document.Data.Bytes()),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %v", err)
	}

	return id, nil
}

// CanvasStore implementation for user-owned canvases
func (s *s3Store) getCanvasKey(userID, canvasID string) string {
	return path.Join(userID, canvasID)
}

func (s *s3Store) List(ctx context.Context, userID string) ([]*core.Canvas, error) {
	prefix := userID + "/"
	output, err := s.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list canvases for user %s: %v", userID, err)
	}

	canvases := make([]*core.Canvas, 0, len(output.Contents))
	for _, object := range output.Contents {
		canvasID := path.Base(*object.Key)
		canvas := &core.Canvas{
			ID:        canvasID,
			UserID:    userID,
			Name:      canvasID, // S3 doesn't have a native 'name' field, using ID.
			UpdatedAt: *object.LastModified,
		}
		canvases = append(canvases, canvas)
	}

	return canvases, nil
}

func (s *s3Store) Get(ctx context.Context, userID, id string) (*core.Canvas, error) {
	key := s.getCanvasKey(userID, id)
	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// A specific check for NoSuchKey can be useful here.
		if bytes.Contains([]byte(err.Error()), []byte("NoSuchKey")) {
			return nil, fmt.Errorf("canvas not found")
		}
		return nil, fmt.Errorf("failed to get canvas %s: %v", id, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read canvas data: %v", err)
	}

	canvas := &core.Canvas{
		ID:        id,
		UserID:    userID,
		Name:      id,
		Data:      data,
		UpdatedAt: *resp.LastModified,
	}

	return canvas, nil
}

func (s *s3Store) Save(ctx context.Context, canvas *core.Canvas) error {
	key := s.getCanvasKey(canvas.UserID, canvas.ID)
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(canvas.Data),
	})
	if err != nil {
		return fmt.Errorf("failed to save canvas %s: %v", canvas.ID, err)
	}
	return nil
}

func (s *s3Store) Delete(ctx context.Context, userID, id string) error {
	key := s.getCanvasKey(userID, id)
	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete canvas %s: %v", id, err)
	}
	return nil
}
