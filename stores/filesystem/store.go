package filesystem

import (
	"bytes"
	"context"
	"excalidraw-complete/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/sirupsen/logrus"
)

type fsStore struct {
	basePath string
}

// NewStore creates a new filesystem-based store.
func NewStore(basePath string) *fsStore {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		log.Fatalf("failed to create base directory: %v", err)
	}
	return &fsStore{basePath: basePath}
}

// DocumentStore implementation for anonymous sharing
func (s *fsStore) FindID(ctx context.Context, id string) (*core.Document, error) {
	filePath := filepath.Join(s.basePath, id)
	log := logrus.WithField("document_id", id)

	log.WithField("file_path", filePath).Info("Retrieving document by ID")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithField("error", "document not found").Warn("Document with specified ID not found")
			return nil, fmt.Errorf("document with id %s not found", id)
		}
		log.WithError(err).Error("Failed to retrieve document")
		return nil, err
	}

	document := core.Document{
		Data: *bytes.NewBuffer(data),
	}

	log.Info("Document retrieved successfully")
	return &document, nil
}

func (s *fsStore) Create(ctx context.Context, document *core.Document) (string, error) {
	id := ulid.Make().String()
	filePath := filepath.Join(s.basePath, id)
	log := logrus.WithFields(logrus.Fields{
		"document_id": id,
		"file_path":   filePath,
	})
	log.Info("Creating new document")

	if err := os.WriteFile(filePath, document.Data.Bytes(), 0644); err != nil {
		log.WithError(err).Error("Failed to create document")
		return "", err
	}

	log.Info("Document created successfully")
	return id, nil
}

// CanvasStore implementation for user-owned canvases
func (s *fsStore) getUserCanvasPath(userID string) string {
	return filepath.Join(s.basePath, userID)
}

func (s *fsStore) List(ctx context.Context, userID string) ([]*core.Canvas, error) {
	userPath := s.getUserCanvasPath(userID)
	log := logrus.WithField("user_id", userID).WithField("path", userPath)

	files, err := os.ReadDir(userPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("User directory does not exist, returning empty list.")
			return []*core.Canvas{}, nil
		}
		log.WithError(err).Error("Failed to read user directory")
		return nil, err
	}

	canvases := make([]*core.Canvas, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				log.WithError(err).Warn("Failed to get file info, skipping file")
				continue
			}
			canvas := &core.Canvas{
				ID:        file.Name(),
				UserID:    userID,
				Name:      file.Name(),
				UpdatedAt: info.ModTime(),
			}
			canvases = append(canvases, canvas)
		}
	}

	log.Infof("Listed %d canvases", len(canvases))
	return canvases, nil
}

func (s *fsStore) Get(ctx context.Context, userID, id string) (*core.Canvas, error) {
	userPath := s.getUserCanvasPath(userID)
	filePath := filepath.Join(userPath, id)
	log := logrus.WithFields(logrus.Fields{"user_id": userID, "canvas_id": id, "path": filePath})

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Canvas file not found")
			return nil, fmt.Errorf("canvas %s not found", id)
		}
		log.WithError(err).Error("Failed to read canvas file")
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		log.WithError(err).Error("Failed to get file stats")
		return nil, err
	}

	canvas := &core.Canvas{
		ID:        id,
		UserID:    userID,
		Name:      id,
		Data:      data,
		UpdatedAt: info.ModTime(),
	}

	log.Info("Canvas retrieved successfully")
	return canvas, nil
}

func (s *fsStore) Save(ctx context.Context, canvas *core.Canvas) error {
	userPath := s.getUserCanvasPath(canvas.UserID)
	filePath := filepath.Join(userPath, canvas.ID)
	log := logrus.WithFields(logrus.Fields{"user_id": canvas.UserID, "canvas_id": canvas.ID, "path": filePath})

	if err := os.MkdirAll(userPath, 0755); err != nil {
		log.WithError(err).Error("Failed to create user directory")
		return err
	}

	log.Info("Saving canvas")
	err := os.WriteFile(filePath, canvas.Data, 0644)
	if err != nil {
		log.WithError(err).Error("Failed to write canvas file")
		return err
	}

	// Set modification time for consistency, though WriteFile usually does this.
	// We preserve created time logic in the storage layer if needed.
	now := time.Now()
	canvas.UpdatedAt = now

	// A full implementation would handle CreatedAt by checking if the file exists first.
	// For this KV-like store, we'll just update ModTime via WriteFile.

	return nil
}

func (s *fsStore) Delete(ctx context.Context, userID, id string) error {
	userPath := s.getUserCanvasPath(userID)
	filePath := filepath.Join(userPath, id)
	log := logrus.WithFields(logrus.Fields{"user_id": userID, "canvas_id": id, "path": filePath})

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Canvas file not found for deletion, considered successful.")
			return nil // If it doesn't exist, the goal is achieved.
		}
		log.WithError(err).Error("Failed to delete canvas file")
		return err
	}

	log.Info("Canvas deleted successfully")
	return nil
}
