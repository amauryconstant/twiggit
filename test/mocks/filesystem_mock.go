// Package mocks contains test mocks for twiggit
package mocks

import (
	"io/fs"
	"os"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockFileInfo implements fs.FileInfo for testing
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// NewMockFileInfo creates a new MockFileInfo with sensible defaults
func NewMockFileInfo(name string, isDir bool) *MockFileInfo {
	return &MockFileInfo{
		name:    name,
		size:    0,
		mode:    0644,
		modTime: time.Now(),
		isDir:   isDir,
	}
}

// NewMockFileInfoWithDetails creates a new MockFileInfo with custom details
func NewMockFileInfoWithDetails(name string, size int64, mode os.FileMode, modTime time.Time, isDir bool) *MockFileInfo {
	return &MockFileInfo{
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
	}
}

func (m *MockFileInfo) Name() string       { return m.name }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// MockDirEntry implements fs.DirEntry for testing
type MockDirEntry struct {
	name string
	mode fs.FileMode
	info *MockFileInfo
}

// NewMockDirEntry creates a new MockDirEntry with sensible defaults
func NewMockDirEntry(name string, isDir bool) *MockDirEntry {
	var mode fs.FileMode = 0644
	if isDir {
		mode = fs.ModeDir | 0755
	}
	return &MockDirEntry{
		name: name,
		mode: mode,
		info: NewMockFileInfo(name, isDir),
	}
}

// NewMockDirEntryWithDetails creates a new MockDirEntry with custom details
func NewMockDirEntryWithDetails(name string, mode fs.FileMode, info *MockFileInfo) *MockDirEntry {
	return &MockDirEntry{name, mode, info}
}

func (m *MockDirEntry) Name() string               { return m.name }
func (m *MockDirEntry) IsDir() bool                { return m.mode&fs.ModeDir != 0 }
func (m *MockDirEntry) Type() fs.FileMode          { return m.mode.Type() }
func (m *MockDirEntry) Info() (fs.FileInfo, error) { return m.info, nil }

// FileSystemMock is a centralized mock for filesystem operations
type FileSystemMock struct {
	mock.Mock
}

// NewFileSystemMock creates a new FileSystemMock
func NewFileSystemMock() *FileSystemMock {
	return &FileSystemMock{}
}

// Stat mocks the Stat operation
func (m *FileSystemMock) Stat(name string) (fs.FileInfo, error) {
	args := m.Called(name)

	// Handle the case where no expectation is set up
	if args.Get(0) == nil {
		return (fs.FileInfo)(nil), args.Error(1)
	}

	return args.Get(0).(fs.FileInfo), args.Error(1)
}

// ReadDir mocks the ReadDir operation
func (m *FileSystemMock) ReadDir(name string) ([]fs.DirEntry, error) {
	args := m.Called(name)
	return args.Get(0).([]fs.DirEntry), args.Error(1)
}

// MkdirAll mocks the MkdirAll operation
func (m *FileSystemMock) MkdirAll(path string, perm os.FileMode) error {
	args := m.Called(path, perm)
	return args.Error(0)
}

// ReadFile mocks the ReadFile operation
func (m *FileSystemMock) ReadFile(filename string) ([]byte, error) {
	args := m.Called(filename)
	return args.Get(0).([]byte), args.Error(1)
}

// WriteFile mocks the WriteFile operation
func (m *FileSystemMock) WriteFile(filename string, data []byte, perm os.FileMode) error {
	args := m.Called(filename, data, perm)
	return args.Error(0)
}

// Remove mocks the Remove operation
func (m *FileSystemMock) Remove(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// Chmod mocks the Chmod operation (test-specific)
func (m *FileSystemMock) Chmod(name string, perm os.FileMode) error {
	args := m.Called(name, perm)
	return args.Error(0)
}

// Create mocks the Create operation (test-specific)
func (m *FileSystemMock) Create(name string) (*os.File, error) {
	args := m.Called(name)
	return args.Get(0).(*os.File), args.Error(1)
}

// RemoveAll mocks the RemoveAll operation (test-specific)
func (m *FileSystemMock) RemoveAll(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// Open mocks the Open operation (required by fs.FS interface)
func (m *FileSystemMock) Open(name string) (fs.File, error) {
	args := m.Called(name)
	return args.Get(0).(fs.File), args.Error(1)
}

// Exists mocks the Exists operation
func (m *FileSystemMock) Exists(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

// SetupPermissionError sets up the mock to return permission errors for all operations
func (m *FileSystemMock) SetupPermissionError() *FileSystemMock {
	m.On("Stat", mock.Anything).Return((fs.FileInfo)(nil), &os.PathError{Op: "stat", Path: "test", Err: os.ErrPermission})
	m.On("ReadDir", mock.Anything).Return([]fs.DirEntry{}, &os.PathError{Op: "readdir", Path: "test", Err: os.ErrPermission})
	m.On("MkdirAll", mock.Anything, mock.Anything).Return(&os.PathError{Op: "mkdir", Path: "test", Err: os.ErrPermission})
	m.On("ReadFile", mock.Anything).Return([]byte{}, &os.PathError{Op: "read", Path: "test", Err: os.ErrPermission})
	m.On("WriteFile", mock.Anything, mock.Anything, mock.Anything).Return(&os.PathError{Op: "write", Path: "test", Err: os.ErrPermission})
	m.On("Remove", mock.Anything).Return(&os.PathError{Op: "remove", Path: "test", Err: os.ErrPermission})
	m.On("Chmod", mock.Anything, mock.Anything).Return(&os.PathError{Op: "chmod", Path: "test", Err: os.ErrPermission})
	m.On("Create", mock.Anything).Return((*os.File)(nil), &os.PathError{Op: "create", Path: "test", Err: os.ErrPermission})
	m.On("RemoveAll", mock.Anything).Return(&os.PathError{Op: "removeall", Path: "test", Err: os.ErrPermission})
	m.On("Open", mock.Anything).Return((fs.File)(nil), &os.PathError{Op: "open", Path: "test", Err: os.ErrPermission})
	return m
}

// SetupReadDirError sets up the mock to return a specific error for ReadDir operations
func (m *FileSystemMock) SetupReadDirError(path string, err error) *FileSystemMock {
	m.On("ReadDir", path).Return([]fs.DirEntry{}, err)
	return m
}

// SetupWriteFileError sets up the mock to return a specific error for WriteFile operations
func (m *FileSystemMock) SetupWriteFileError(filename string, err error) *FileSystemMock {
	m.On("WriteFile", filename, mock.Anything, mock.Anything).Return(err)
	return m
}

// SetupMkdirAllError sets up the mock to return a specific error for MkdirAll operations
func (m *FileSystemMock) SetupMkdirAllError(path string, err error) *FileSystemMock {
	m.On("MkdirAll", path, mock.Anything).Return(err)
	return m
}

// SetupFileSystem sets up the mock to delegate to real filesystem operations
func (m *FileSystemMock) SetupFileSystem() *FileSystemMock {
	// This method is deprecated - use specific mock expectations instead
	return m
}
