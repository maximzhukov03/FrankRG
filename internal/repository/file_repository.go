package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type TypeObject struct{
	Indx int
}

type Object struct{
	Name string
	Type TypeObject
	Size int
	TimeModification time.Time 
}

type BrowseFiles struct{
	Root string
}


func NewBrowseFiles(root string) (*BrowseFiles, error){
	if root == ""{
		return nil, errors.New("empty root")
	}

	abs, err := filepath.Abs(root)
	if err != nil{
		return nil, fmt.Errorf("abs filepath error: %w", err)
	}

	isDir, err := os.Stat(abs)
	if err != nil{
		return nil, fmt.Errorf("stat info error: %w", err)
	}
	if !isDir.IsDir(){
		return nil, fmt.Errorf("root isnt dir: %s", abs)
	}
	return &BrowseFiles{
		Root: filepath.Clean(abs),
	}, nil

}

// func (bf *BrowseFiles) NormalizerPath(path string)(string, error){
// 	var normalPath string 
	
// 	filePath 

// 	return normalPath, nil 
// }

func (bf BrowseFiles) List(path string) ([]Object, error){
	allObjects, err := os.ReadDir(path)
	if err != nil{
		return nil, fmt.Errorf("try to read dir: %w", err)
	}
	
	objectSlice := make([]Object, 0, len(allObjects))
	for _, obj := range allObjects{
		info, err := obj.Info()
		if err != nil{
			return nil, fmt.Errorf("error info object: %w", err)
		}
		if info.IsDir(){
			object := Object{
				Name: info.Name(),
				Size: 0,
				Type: TypeObject{Indx: 0},
				TimeModification: info.ModTime(),
			}
			objectSlice = append(objectSlice, object)
		}else{
			object := Object{
				Name: info.Name(),
				Size: int(info.Size()),
				Type: TypeObject{Indx: 1},
				TimeModification: info.ModTime(),
			}
			objectSlice = append(objectSlice, object)
		}
	}
	sort.SliceStable(objectSlice, func(i, j int) bool {
		if objectSlice[i].Type != objectSlice[j].Type{
			return objectSlice[i].Type == TypeObject{Indx: 0}
		}
		return strings.ToLower(objectSlice[i].Name) < strings.ToLower(objectSlice[j].Name)
	})
	return objectSlice, nil
}

func (bf *BrowseFiles) MakeDir(path, name string) error{
	return os.Mkdir(filepath.Join(path, name), 0o755)
}

func (bf *BrowseFiles) Rename(path, name, newName string) error{
	
	oldPath := filepath.Join(path, name)
	newPath := filepath.Join(path, newName)

	return os.Rename(oldPath, newPath)
}

func (bf *BrowseFiles) Delete(path, name string) error{
	fullPath := filepath.Join(path, name)
	err := os.Remove(fullPath)
	if err != nil{
		return fmt.Errorf("error remove: %w", err)
	}
	return nil
}