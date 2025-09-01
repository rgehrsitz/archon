package api

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// DialogService provides file dialog functionality
type DialogService struct {
	ctx context.Context
}

// NewDialogService creates a new DialogService
func NewDialogService() *DialogService {
	return &DialogService{}
}

// SetContext sets the Wails context for runtime calls
func (s *DialogService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// OpenDirectoryDialog opens a native directory selection dialog
func (s *DialogService) OpenDirectoryDialog() (string, error) {
	if s.ctx == nil {
		return "", nil
	}

	// Use Wails runtime to show native directory dialog
	selectedPath, err := runtime.OpenDirectoryDialog(s.ctx, runtime.OpenDialogOptions{
		Title: "Select Archon Project Directory",
		DefaultDirectory: "", // Let OS choose default
		CanCreateDirectories: false,
	})

	if err != nil {
		return "", err
	}

	return selectedPath, nil
}