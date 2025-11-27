package controllers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	// "time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// UploadResponse
type UploadResponse struct {
	URL string `json:"url"`
}

// UploadImage godoc
//
//	@Summary		Upload an image
//	@Description	Upload an image file (jpg, png, jpeg) and get a local URL. Max size 5MB.
//	@Tags			Utility
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Image file"
//	@Success		201		{object}	UploadResponse
//	@Failure		400		{object}	map[string]string	"Invalid file"
//	@Failure		500		{object}	map[string]string	"Server error"
//	@Router			/upload [post]
func UploadImage(c echo.Context) error {
	// 1. Read form file
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No file uploaded"})
	}

	// 2. Validate File Size (e.g., Max 5MB)
	if file.Size > 5*1024*1024 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "File too large (Max 5MB)"})
	}

	// 3. Open the file
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not open file"})
	}
	defer src.Close()

	// 4. Create unique filename
	// Extract extension
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Only JPG, JPEG, and PNG allowed"})
	}

	// Generate random name: "upload-<uuid><ext>"
	newFileName := "upload-" + uuid.New().String() + ext

	// 5. Ensure upload directory exists
	uploadPath := "./uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.Mkdir(uploadPath, 0755)
	}

	// 6. Create destination file
	dstPath := filepath.Join(uploadPath, newFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Could not create destination file"})
	}
	defer dst.Close()

	// 7. Copy data
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save file"})
	}

	// 8. Return the relative URL
	// Ideally, construct full URL based on env var, but relative path works for simple setups
	// The URL will look like: /uploads/upload-xyz-123.jpg
	fileURL := "/uploads/" + newFileName

	return c.JSON(http.StatusCreated, UploadResponse{URL: fileURL})
}
