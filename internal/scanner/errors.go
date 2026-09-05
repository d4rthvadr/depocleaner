package scanner

import (
	"errors"
	"fmt"

	"github.com/d4rthvadr/node-cleaner/pkg/models"
)

func (s *Scanner) handleError(err error, path string) {
	var appErr *models.ApplicationError

	if errors.As(err, &appErr) {

		switch appErr.Type {
		case models.ErrTypePermission:
			// TODO: log warning about permission denied
			// s.logger.Warn("Skipping inaccessible path", "path", path)
			fmt.Printf("Skipping inaccessible path: %s\n", path)
		case models.ErrTypeNotFound:
			// s.logger.Warn("Path not found", "path", path)
			fmt.Printf("Path not found: %s\n", path)
		default:
			// s.logger.Error("Error scanning path", "path", path, "error", err)
			fmt.Printf("Error scanning path: %s, error: %v\n", path, err)
		}
	} else {
		// s.logger.Error("Unexpected error scanning path", "path", path, "error", err)
		fmt.Printf("Unexpected error scanning path: %s, error: %v\n", path, err)
	}

	select {
	case s.errorChan <- err:
	default:
		// If errorChan is full, drop the error to avoid blocking
	}

}
