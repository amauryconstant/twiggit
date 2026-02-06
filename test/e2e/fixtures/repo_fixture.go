//go:build e2e
// +build e2e

package fixtures

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
)

func findFixtureArchive(name string) string {
	wd, _ := os.Getwd()

	candidates := []string{
		filepath.Join(wd, "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join(wd, "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("fixtures", "repos", name+".tar.gz"),
		filepath.Join("..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("..", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("..", "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
		filepath.Join("..", "..", "..", "..", "test", "e2e", "fixtures", "repos", name+".tar.gz"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	return ""
}

type RepoFixture struct {
	name     string
	repoPath string
	tempDir  string
}

func ExtractRepoFixture(name string) (*RepoFixture, error) {
	archivePath := findFixtureArchive(name)

	if archivePath == "" {
		wd, _ := os.Getwd()
		return nil, fmt.Errorf("repo fixture '%s' not found (working dir: %s, tried: test/e2e/fixtures/repos/%s.tar.gz, fixtures/repos/%s.tar.gz)",
			name, wd, name, name)
	}

	tempDir := GinkgoT().TempDir()
	repoPath := filepath.Join(tempDir, name)

	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create repo directory: %w", err)
	}

	if err := extractTarGz(archivePath, repoPath); err != nil {
		return nil, fmt.Errorf("failed to extract repo fixture: %w", err)
	}

	return &RepoFixture{
		name:     name,
		repoPath: repoPath,
		tempDir:  tempDir,
	}, nil
}

func extractTarGz(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dst, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func (rf *RepoFixture) Path() string {
	return rf.repoPath
}

func (rf *RepoFixture) TempDir() string {
	return rf.tempDir
}

func (rf *RepoFixture) Name() string {
	return rf.name
}

// extractRepoFixtureToDir extracts a fixture archive to a specific destination directory
// This allows extracting fixtures directly into f.tempDir without creating additional temp directories
func extractRepoFixtureToDir(fixtureName, destDir string) error {
	archivePath := findFixtureArchive(fixtureName)
	if archivePath == "" {
		wd, _ := os.Getwd()
		return fmt.Errorf("repo fixture '%s' not found (working dir: %s, tried: test/e2e/fixtures/repos/%s.tar.gz, fixtures/repos/%s.tar.gz)",
			fixtureName, wd, fixtureName, fixtureName)
	}
	return extractTarGz(archivePath, destDir)
}
