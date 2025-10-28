package service

import (
	database "browserfiles/test/internal/repository"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrOutRoot = errors.New("path outside of root")
	ErrEmptyName = errors.New("empty name")
	ErrInvalidName = errors.New("invalid new name")
)

type FileServicer interface{
	List(path string) ([]database.Object, error)
	MakeDir(path, name string) error
	Rename(path, name, newName string) error
	Delete(path, name string) error
	Save(path, name string, r io.Reader) (int64, error)
	Download(path string) (*os.File, string, string, error)
	Exist(logicPath string) (ex bool, isDir bool, err error)
}

type FileService struct{
	repo database.Repository
	root string
}

func NewFileService(repo database.Repository) *FileService{
	return &FileService{
		repo: repo,
		root: repo.RootPath(),
	}
}

func (s *FileService) joinSafe(logicalPath string, parts ...string) (string, error){
	lp := strings.TrimSpace(logicalPath)
	if lp == "" {
		lp = "."
	}
	joined := filepath.Join(append([]string{s.root, lp}, parts...)...)
	clean := filepath.Clean(joined)

	if !strings.HasPrefix(clean+string(filepath.Separator), s.root+string(filepath.Separator)){
		return "", ErrOutRoot
	}
	return clean, nil
}

func (s *FileService) List(logicalPath string) ([]database.Object, error){
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return nil, err
	}
	return s.repo.List(p)
}

func (s *FileService) MakeDir(logicalPath, name string) error{
	name = strings.TrimSpace(name)
	if name == ""{
		return ErrEmptyName
	}
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return err
	}
	return s.repo.MakeDir(p, name)
}

func (s *FileService) Rename(logicalPath, name, newName string) error{
	name = strings.TrimSpace(name)
	newName = strings.TrimSpace(newName)
	if name == "" || newName == "" || newName == name{
		return ErrEmptyName
	}
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return err
	}
	return s.repo.Rename(p, name, newName)
}

func (s *FileService) Delete(logicalPath, name string) error{
	name = strings.TrimSpace(name)
	if name == ""{
		return ErrEmptyName
	}
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return err
	}
	return s.repo.Delete(p, name)
}

func (s *FileService) Save(logicalPath, name string, r io.Reader) (int64, error){
	name = strings.TrimSpace(name)
	if name == ""{
		return 0, ErrEmptyName
	}
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return 0, err
	}
	return s.repo.Save(p, name, r)
}

func (s *FileService) Download(logicalPath string) (*os.File, string, string, error){
	p, err := s.joinSafe(logicalPath)
	if err != nil{
		return nil, "", "", err
	}
	f, mime, err := s.repo.Download(p)
	if err != nil{
		return nil, "", "", err
	}
	filename := filepath.Base(p)
	return f, mime, filename, nil
}

func (s *FileService) Exist(logicPath string) (bool, bool, error){
	jS, err := s.joinSafe(logicPath)
	if err != nil{
		return  false, false, err
	}

	return  s.repo.Exist(jS)
}