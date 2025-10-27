package database

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TypeObject int
const(
	TypeDir TypeObject = iota 
	TypeFile
)
const accessRights = 0o755

type Object struct{
	Name string
	Type TypeObject
	Size int
	TimeModification time.Time 
}

type Options struct{
	ConflictingName bool
}

type BrowseFiles struct{
	Root string
	Opt Options
}


func NewBrowseFiles(root string) (*BrowseFiles, error){
	if root == ""{
		return nil, errors.New("empty root")
	}

	abs, err := filepath.Abs(root)
	if err != nil{
		return nil, fmt.Errorf("abs filepath error: %w", err)
	}

	stat, err := os.Stat(abs)
	if err != nil{
		return nil, fmt.Errorf("stat info error: %w", err)
	}
	if !stat.IsDir(){
		return nil, fmt.Errorf("root isnt dir: %s", abs)
	}
	return &BrowseFiles{
		Root: filepath.Clean(abs),
	}, nil

}

func (bf BrowseFiles) List(path string) ([]Object, error){
	allObjects, err := os.ReadDir(path)
	if err != nil{
		return nil, fmt.Errorf("try to read dir: %w", err)
	}
	
	objectSlice := make([]Object, 0, len(allObjects))
	for _, obj := range allObjects{
		info, err := obj.Info()
		if err != nil{
			return nil, fmt.Errorf("info object: %w", err)
		}
		if info.IsDir(){
			object := Object{
				Name: info.Name(),
				Size: 0,
				Type: TypeDir,
				TimeModification: info.ModTime(),
			}
			objectSlice = append(objectSlice, object)
		}else{
			object := Object{
				Name: info.Name(),
				Size: int(info.Size()),
				Type: TypeFile,
				TimeModification: info.ModTime(),
			}
			objectSlice = append(objectSlice, object)
		}
	}
	sort.SliceStable(objectSlice, func(i, j int) bool {
		if objectSlice[i].Type != objectSlice[j].Type{
			return objectSlice[i].Type == TypeDir
		}
		return strings.ToLower(objectSlice[i].Name) < strings.ToLower(objectSlice[j].Name)
	})
	return objectSlice, nil
}

func (bf *BrowseFiles) MakeDir(path, name string) error{
	p := filepath.Clean(filepath.Join(path, name))
	return os.Mkdir(p, accessRights)
}

func (bf *BrowseFiles) Rename(path, name, newName string) error{
	
	oldPath := filepath.Join(path, name)
	newPath := filepath.Join(path, newName)

	return os.Rename(oldPath, newPath)
}

func (bf *BrowseFiles) Delete(path, name string) error{
	fullPath := filepath.Clean(filepath.Join(path, name))
	err := os.RemoveAll(fullPath)
	if err != nil{
		return fmt.Errorf("remove error: %w", err)
	}
	return nil
}

func (bf *BrowseFiles) Save(path, name string, r io.Reader) (int64, error){
	distPath := filepath.Join(path, name)
	if bf.Opt.ConflictingName == false{
		return 0, fmt.Errorf("file or dir name '%s' not unique", name)
	}
	tempFile, err := os.CreateTemp(path, ".tempFile-*")
	if err != nil{
		return 0, fmt.Errorf("create temp file: %w", err)
	}
	written, err := io.Copy(tempFile, r)
	if err != nil{
		err := tempFile.Close()
		if err != nil{
			return written, fmt.Errorf("close temporary file: %w", err)
		}
		return written, err 
	}
	err = tempFile.Sync()
	if err != nil{
		err := tempFile.Close()
		if err != nil{
			return written, fmt.Errorf("error close temporary file")
		}
		return written, fmt.Errorf("error sync commits the current contents: %w", err) 
	}
	err = tempFile.Close()
	if err != nil{
		return written, fmt.Errorf("error of close temp file: %w", err)
	}

	if bf.Opt.ConflictingName{
		err := os.Remove(distPath)
		if err != nil{
			return written, fmt.Errorf("error of deleting temporary file")
		}
	}
	err = os.Rename(tempFile.Name(), distPath)
	if err !=nil{
		_ = os.Remove(tempFile.Name())
		return written, fmt.Errorf("rename temp file: %w", err)
	}
	err = os.Remove(tempFile.Name())
	if err != nil{
		return written, fmt.Errorf("remove temp file: %w", err)
	}
	return written, nil
}

func (bf *BrowseFiles) Download(path string) (*os.File, string, error){
	stat, err := os.Stat(path)
	if err != nil{
		return nil, "" ,fmt.Errorf("stat info error: %w",err)
	}
	if stat.IsDir(){
		return nil, "", errors.New("error download, path is dir")
	}
	op, err := os.Open(path)
	if err != nil{
		return nil, "", fmt.Errorf("error open os path: %w", err)
	}
	mi := mime.TypeByExtension(filepath.Ext(path))
	if mi == ""{
		buffer := make([]byte, 512)
		n, err := op.Read(buffer)
		if err != nil && err != io.EOF{
			op.Close()
			return nil, "", fmt.Errorf("read file for mime: %w", err)
		}
		_,_ = op.Seek(0, io.SeekStart)
		mi = http.DetectContentType(buffer[:n])
	}
	if mi == ""{
		mi = "application/octet-stream"
	}
	return op, mi, nil
}

func (bf *BrowseFiles) Exist(path string) (ex bool, Dir bool, err error){
	stat, err := os.Stat(path)
	if err != nil{
		if os.IsNotExist(err){
			return false, false, nil
		}
		return false, false, err
	}
	return true, stat.IsDir(), nil
}