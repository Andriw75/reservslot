package storage

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct {
	imagesDir string
}

func NewStorage(imagesDir string) *Storage {
	return &Storage{imagesDir: imagesDir}
}

func (s *Storage) EnsureDirs() error {
	return os.MkdirAll(s.imagesDir, 0o755)
}

// SaveFiles guarda varios archivos en disco. Si alguno falla, elimina los ya
// escritos (rollback de esta operación) y devuelve el error.
// allowedExts restringe extensiones (ej. map[".jpg"]true); si es nil acepta todo.
func (s *Storage) SaveFiles(files []*multipart.FileHeader, prefix string, allowedExts map[string]bool) ([]string, error) {
	var saved []string
	for i, fh := range files {
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if allowedExts != nil && !allowedExts[ext] {
			s.removeSaved(saved)
			return nil, fmt.Errorf("extensión %q no permitida", ext)
		}

		name := fmt.Sprintf("%s_%d_%d%s", prefix, time.Now().UnixMilli(), i, ext)
		if err := s.saveFile(fh, name); err != nil {
			s.removeSaved(saved)
			return nil, err
		}
		saved = append(saved, name)
	}
	return saved, nil
}

func (s *Storage) saveFile(fh *multipart.FileHeader, name string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join(s.imagesDir, name))
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		os.Remove(filepath.Join(s.imagesDir, name))
		return err
	}
	return dst.Close()
}

func (s *Storage) removeSaved(saved []string) {
	for _, name := range saved {
		_ = os.Remove(filepath.Join(s.imagesDir, name))
	}
}

// RemoveImages borra archivos del disco (best-effort, tras éxito de la BD).
func (s *Storage) RemoveImages(filenames ...string) {
	for _, name := range filenames {
		if name == "" {
			continue
		}
		if err := os.Remove(filepath.Join(s.imagesDir, name)); err != nil && !os.IsNotExist(err) {
			slog.Warn("no se pudo eliminar imagen", "file", name, "error", err)
		}
	}
}

func (s *Storage) ImageURL(filename string) string {
	return "/imagenes/" + filename
}
