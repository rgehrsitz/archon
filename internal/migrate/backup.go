package migrate

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
)

// Backup placeholder: create /backups/<timestamp>/ before mutation.
func Backup(dir string) error {
    return os.MkdirAll(dir, 0o755)
}

// CreateBackup creates a timestamped backup directory under
//   <basePath>/backups/<ISO8601>/
// and copies core project data (project.json, nodes/, attachments/).
// Returns the created backup directory path.
func CreateBackup(basePath string) (string, error) {
    ts := time.Now().UTC().Format("20060102T150405Z")
    backupDir := filepath.Join(basePath, "backups", ts)
    if err := os.MkdirAll(backupDir, 0o755); err != nil {
        return "", err
    }

    // Copy project.json if present
    if err := copyIfExists(filepath.Join(basePath, "project.json"), filepath.Join(backupDir, "project.json")); err != nil {
        return "", err
    }
    // Copy nodes directory
    if err := copyDirIfExists(filepath.Join(basePath, "nodes"), filepath.Join(backupDir, "nodes")); err != nil {
        return "", err
    }
    // Copy attachments directory
    if err := copyDirIfExists(filepath.Join(basePath, "attachments"), filepath.Join(backupDir, "attachments")); err != nil {
        return "", err
    }

    return backupDir, nil
}

func copyIfExists(src, dst string) error {
    info, err := os.Stat(src)
    if err != nil {
        if os.IsNotExist(err) { return nil }
        return err
    }
    if !info.Mode().IsRegular() {
        return fmt.Errorf("not a regular file: %s", src)
    }
    if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil { return err }
    in, err := os.Open(src)
    if err != nil { return err }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil { return err }
    defer func() { _ = out.Close() }()
    if _, err := io.Copy(out, in); err != nil { return err }
    return out.Sync()
}

func copyDirIfExists(src, dst string) error {
    info, err := os.Stat(src)
    if err != nil {
        if os.IsNotExist(err) { return nil }
        return err
    }
    if !info.IsDir() { return nil }
    return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
        if err != nil { return err }
        rel, err := filepath.Rel(src, path)
        if err != nil { return err }
        target := filepath.Join(dst, rel)
        if fi.IsDir() {
            return os.MkdirAll(target, 0o755)
        }
        return copyIfExists(path, target)
    })
}
