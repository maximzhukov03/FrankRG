package handlers

import (
	"browserfiles/test/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Handler struct{
	service service.FileServicer
}

func NewHandler(serv service.FileServicer) *Handler{ 
	return &Handler{service: serv}
}

func writeJSON(w http.ResponseWriter, status int, v any){
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request){
	path := r.URL.Query().Get("path")
	objs, err := h.service.List(path)
	if err != nil{
		mapErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": objs,
	})
}

type dirReq struct{
	Path string `json:"path"`
	Name string `json:"name"`
}

func (h *Handler) Mkdir(w http.ResponseWriter, r *http.Request){
	var req dirReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid json"})
		return
	}
	if err := h.service.MakeDir(req.Path, req.Name); err != nil{
		mapErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status":"ok"})
}

type renameReq struct{
	Path    string `json:"path"`
	Name    string `json:"name"`
	NewName string `json:"newName"`
}

func (h *Handler) Rename(w http.ResponseWriter, r *http.Request){
	var req renameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid json"})
		return
	}
	if err := h.service.Rename(req.Path, req.Name, req.NewName); err != nil{
		mapErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
}

type deleteReq struct{
	Path string `json:"path"`
	Name string `json:"name"`
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request){
	var req deleteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil{
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"invalid json"})
		return
	}
	if err := h.service.Delete(req.Path, req.Name); err != nil{
		mapErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status":"ok"})
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request){
	path := r.URL.Query().Get("path")

	if err := r.ParseMultipartForm(32 << 20); err != nil{
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"multipart parse error"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil{
		writeJSON(w, http.StatusBadRequest, map[string]string{"error":"missing file"})
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	if strings.TrimSpace(name) == ""{
		name = header.Filename
	}

	if _, err := h.service.Save(path, name, file); err != nil{
		mapErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status":"ok"})
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request){
	path := r.URL.Query().Get("path")
	f, mime, filename, err := h.service.Download(path)
	if err != nil {
		mapErr(w, err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, _ = io.Copy(w, f)
}

func mapErr(w http.ResponseWriter, err error){
	switch{
	case errors.Is(err, service.ErrOutRoot):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case errors.Is(err, service.ErrEmptyName),
		errors.Is(err, service.ErrInvalidName):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case os.IsNotExist(err):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
}