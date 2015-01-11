package main

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hoisie/mustache"
)

func FileTemplate(f File) map[string]string {
	// Compute a human-readable file size
	var sizeString string
	if f.Size < 1024 {
		sizeString = strconv.FormatInt(f.Size, 10) + " bytes"
	} else if f.Size < 1048576 {
		sizeString = strconv.FormatInt(f.Size/1024, 10) + " KiB"
	} else if f.Size < 1073741824 {
		sizeString = strconv.FormatInt(f.Size/1048576, 10) + " MiB"
	} else {
		sizeString = strconv.FormatInt(f.Size/1073741824, 10) + " GiB"
	}

	// Create a human-readable date
	t := time.Unix(f.Uploaded, 0)
	dateString := strconv.Itoa(int(t.Month())) + "/" + strconv.Itoa(t.Day()) +
		"/" + strconv.Itoa(t.Year())

	// Find the appropriate image
	icon := "unknown"
	icons := map[string]string{
		"png":  "image",
		"jpg":  "image",
		"jpeg": "image",
		"gif":  "image",
	}
	lowerName := strings.ToLower(f.Name)
	for ext, img := range icons {
		if strings.HasSuffix(lowerName, "."+ext) {
			icon = img
			break
		}
	}

	return map[string]string{
		"name":     f.Name,
		"id":       f.Id,
		"uploaded": dateString,
		"size":     sizeString,
		"icon":     "images/icons/" + icon + ".png",
	}
}

func HandleFiles(w http.ResponseWriter, r *http.Request) {
	if !IsAuthenticated(w, r) {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// We don't want their browser to cache the file list.
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	DbLock.RLock()
	fileMaps := make([]map[string]string, len(Database.Files))
	for i, file := range Database.Files {
		fileMaps[i] = FileTemplate(file)
	}
	DbLock.RUnlock()

	template := map[string]interface{}{"files": fileMaps}
	templatePath := filepath.Join(AssetsPath, "files.mustache")
	body := mustache.RenderFile(templatePath, template)
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(body))
}
