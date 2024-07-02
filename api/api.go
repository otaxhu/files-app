package api

import (
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"regexp"
	"strconv"

	"github.com/elnormous/contenttype"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/otaxhu/files-app/dto"
	"github.com/otaxhu/files-app/service"
)

func Start(addr string, serv *service.FileService) error {
	mux := chi.NewMux()

	mux.Use(cors.Handler(cors.Options{
		AllowedMethods: []string{"POST", "GET"},
	}))

	mux.Get("/file/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("show_info") {
			w.Header().Set("Content-Type", "application/json")
			info, err := serv.GetFileInfo(r.Context(), r.PathValue("id"))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"filename": info.Filename,
				"size":     info.Len,
			})
			return
		}

		file, err := serv.GetFile(r.Context(), r.PathValue("id"))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
			return
		}
		defer file.Reader.Close()

		// Trigger download
		w.Header().Set("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", strconv.Itoa(int(file.Len)))

		io.CopyN(w, file.Reader, file.Len)
	})

	mux.Post("/file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		f, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
			return
		}
		defer f.Close()

		id, err := serv.SaveFile(r.Context(), dto.SaveFile{
			Filename: header.Filename,
			Reader:   f,
			Len:      header.Size,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"id": id})
	})
	return http.ListenAndServe(addr, mux)
}

var filePathMatcher = regexp.MustCompile(`^/file(/[^/]+)?/?$`)

func StartFrontend(addr string, f fs.FS) error {
	mux := chi.NewMux()
	mux.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		accepted, _, err := contenttype.GetAcceptableMediaType(r, []contenttype.MediaType{
			contenttype.NewMediaType("*/*"),
			contenttype.NewMediaType("text/html"),
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
			return
		}
		if accepted.String() == "text/html" {
			if !filePathMatcher.MatchString(r.URL.Path) {
				file, _ := f.Open("404.html")
				defer file.Close()
				info, _ := file.Stat()
				w.Header().Set("Content-Length", strconv.Itoa(int(info.Size())))
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusNotFound)
				io.CopyN(w, file, info.Size())
				return
			}
			w.Header().Set("Content-Type", "text/html")
			http.ServeFileFS(w, r, f, "index.html")
			return
		}
		http.ServeFileFS(w, r, f, r.URL.Path)
	})
	return http.ListenAndServe(addr, mux)
}
