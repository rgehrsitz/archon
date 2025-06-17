// Package storage handles file attachments for Archon projects.
package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Attachment represents metadata about a file in the attachments directory.
type Attachment struct {
	Name      string    `json:"name"`      // Filename
	Size      int64     `json:"size"`      // Size in bytes
	CreatedAt time.Time `json:"createdAt"` // Time of addition
}

// attachmentsMetadataFile is the metadata file name in the attachments dir.
const attachmentsMetadataFile = "attachments.json"

// LoadAttachments loads attachment metadata from the project's attachments directory.
func (v *ConfigVault) LoadAttachments() ([]Attachment, error) {
	metaPath := filepath.Join(v.GetAttachmentsPath(), attachmentsMetadataFile)
	f, err := os.Open(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Attachment{}, nil // No attachments yet
		}
		return nil, fmt.Errorf("failed to open attachments metadata: %w", err)
	}
	defer f.Close()
	var attachments []Attachment
	if err := json.NewDecoder(f).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("failed to decode attachments metadata: %w", err)
	}
	return attachments, nil
}

// SaveAttachments writes attachment metadata to the project's attachments directory.
func (v *ConfigVault) SaveAttachments(attachments []Attachment) error {
	metaPath := filepath.Join(v.GetAttachmentsPath(), attachmentsMetadataFile)
	f, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to create attachments metadata: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(attachments)
	if err == nil {
		v.changeCount++
		v.autoSnapshot("")
	}
	return err
}

// AddAttachment copies a file into the attachments directory and updates metadata.
func (v *ConfigVault) AddAttachment(srcPath string) error {
	attachmentsDir := v.GetAttachmentsPath()
	base := filepath.Base(srcPath)
	dstPath := filepath.Join(attachmentsDir, base)

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()
	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()
	size, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	// Update metadata
	attachments, err := v.LoadAttachments()
	if err != nil {
		return err
	}
	attachments = append(attachments, Attachment{
		Name:      base,
		Size:      size,
		CreatedAt: time.Now().UTC(),
	})
	return v.SaveAttachments(attachments)
}

// RemoveAttachment deletes a file from the attachments directory and updates metadata.
func (v *ConfigVault) RemoveAttachment(name string) error {
	attachmentsDir := v.GetAttachmentsPath()
	filePath := filepath.Join(attachmentsDir, name)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove attachment: %w", err)
	}
	attachments, err := v.LoadAttachments()
	if err != nil {
		return err
	}
	newAttachments := make([]Attachment, 0, len(attachments))
	for _, att := range attachments {
		if att.Name != name {
			newAttachments = append(newAttachments, att)
		}
	}
	return v.SaveAttachments(newAttachments)
}
