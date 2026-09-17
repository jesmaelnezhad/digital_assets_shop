package handlers

import (
	"database/sql"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pawradise/shared/database"
	"github.com/pawradise/shared/middleware"
)

const (
	// uploadBasePath is the root directory for file storage (PersistentVolume)
	uploadBasePath = "/data/media"
	// maxUploadSize limits file uploads to 50MB
	maxUploadSize = 50 << 20
	// thumbnailWidth is the default thumbnail width
	thumbnailWidth = 200
	// previewWidth is the default preview width
	previewWidth = 800
)

// allowedExtensions defines permitted file types
var allowedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
	".pdf":  "application/pdf",
	".zip":  "application/zip",
	".svg":  "image/svg+xml",
}

// UploadFile handles file uploads to the PersistentVolume.
// Only admins can upload files.
func UploadFile(c *gin.Context) {
	// Limit upload size
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file: " + err.Error()})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mimeType, allowed := allowedExtensions[ext]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File type '%s' not allowed. Allowed types: %s",
				ext, getAllowedExtensionsList()),
		})
		return
	}

	// Generate safe filename with timestamp to avoid collisions
	timestamp := time.Now().UnixNano()
	safeName := fmt.Sprintf("%d_%s", timestamp, sanitizeFilename(header.Filename))
	destPath := filepath.Join(uploadBasePath, safeName)

	// Ensure directory exists
	if err := os.MkdirAll(uploadBasePath, 0755); err != nil {
		log.Printf("[media-service] Failed to create upload directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		log.Printf("[media-service] Failed to create file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file content
	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(destPath)
		log.Printf("[media-service] Failed to write file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Build file URL
	fileURL := fmt.Sprintf("/api/v1/media/files/%s", safeName)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "File uploaded successfully",
		"filename":    header.Filename,
		"stored_as":   safeName,
		"size":        written,
		"mime_type":   mimeType,
		"url":         fileURL,
		"uploaded_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// DownloadFile serves files after auth verification.
// Access is controlled by the JwtAuthMiddleware and RequireAuth.
func DownloadFile(c *gin.Context) {
	filePath := c.Param("path")

	// Security: prevent directory traversal
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(uploadBasePath, cleanPath)

	// Check file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot download a directory"})
		return
	}

	// Log download with user info
	if userID, ok := middleware.GetUserIDFromContext(c); ok {
		log.Printf("[media-service] User %d downloading file: %s", userID, cleanPath)
	}

	// Serve file
	c.File(fullPath)
}

// DeleteFile removes a file from storage (admin only).
func DeleteFile(c *gin.Context) {
	filePath := c.Param("path")

	// Security: prevent directory traversal
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(uploadBasePath, cleanPath)

	// Check file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	// Delete the file
	if err := os.Remove(fullPath); err != nil {
		log.Printf("[media-service] Failed to delete file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	log.Printf("[media-service] File deleted: %s", cleanPath)
	c.JSON(http.StatusOK, gin.H{
		"message":    "File deleted successfully",
		"path":       cleanPath,
		"deleted_at": time.Now().UTC().Format(time.RFC3339),
	})
}

// VerifyAccess checks if the authenticated user has access to media resources.
func VerifyAccess(c *gin.Context) {
	userID, ok := middleware.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_access": true,
		"user_id":    userID,
		"message":    "Access verified",
	})
}

// GetFileMetadata returns metadata about a file without downloading it.
func GetFileMetadata(c *gin.Context) {
	filePath := c.Param("path")

	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid file path"})
		return
	}

	fullPath := filepath.Join(uploadBasePath, cleanPath)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to access file"})
		return
	}

	ext := strings.ToLower(filepath.Ext(cleanPath))
	mimeType, _ := allowedExtensions[ext]

	c.JSON(http.StatusOK, gin.H{
		"filename":    info.Name(),
		"size":        info.Size(),
		"mime_type":   mimeType,
		"extension":   ext,
		"modified_at": info.ModTime().UTC().Format(time.RFC3339),
		"is_dir":      info.IsDir(),
	})
}

// ValidateFileType checks if a file's extension is allowed.
func ValidateFileType(c *gin.Context) {
	var req struct {
		Filename string `json:"filename" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	ext := strings.ToLower(filepath.Ext(req.Filename))
	mimeType, allowed := allowedExtensions[ext]

	c.JSON(http.StatusOK, gin.H{
		"filename":  req.Filename,
		"extension": ext,
		"mime_type": mimeType,
		"allowed":   allowed,
	})
}

// GeneratePreview creates a preview version of an image for a product.
// It resizes the uploaded image to preview dimensions and stores it.
func GeneratePreview(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve image: " + err.Error()})
		return
	}
	defer file.Close()

	// Validate image type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG, PNG, and WebP images are supported"})
		return
	}

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image: " + err.Error()})
		return
	}

	// Create product-specific directory
	productDir := filepath.Join(uploadBasePath, "products", productID)
	if err := os.MkdirAll(productDir, 0755); err != nil {
		log.Printf("[media-service] Failed to create product directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Resize image to preview dimensions (maintaining aspect ratio)
	previewImg := resizeImage(img, previewWidth, 0)
	previewName := fmt.Sprintf("preview_%d%s", time.Now().UnixNano(), ext)
	previewPath := filepath.Join(productDir, previewName)

	previewFile, err := os.Create(previewPath)
	if err != nil {
		log.Printf("[media-service] Failed to create preview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preview"})
		return
	}
	defer previewFile.Close()

	// Encode preview based on format
	switch format {
	case "jpeg":
		err = jpeg.Encode(previewFile, previewImg, &jpeg.Options{Quality: 85})
	case "png":
		err = png.Encode(previewFile, previewImg)
	default:
		// Default to JPEG for unknown formats
		err = jpeg.Encode(previewFile, previewImg, &jpeg.Options{Quality: 85})
	}

	if err != nil {
		os.Remove(previewPath)
		log.Printf("[media-service] Failed to encode preview: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode preview"})
		return
	}

	// Get dimensions
	bounds := previewImg.Bounds()

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Preview generated successfully",
		"product_id":  productID,
		"format":      format,
		"width":       bounds.Dx(),
		"height":      bounds.Dy(),
		"filename":    previewName,
		"size":        getFileSize(previewPath),
		"preview_url": fmt.Sprintf("/api/v1/media/files/products/%s/%s", productID, previewName),
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	})
}

// CreateThumbnail creates a thumbnail version of an image.
func CreateThumbnail(c *gin.Context) {
	productID := c.Param("productId")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve image: " + err.Error()})
		return
	}
	defer file.Close()

	// Validate image type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG, PNG, and WebP images are supported"})
		return
	}

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode image: " + err.Error()})
		return
	}

	// Create product-specific directory
	productDir := filepath.Join(uploadBasePath, "products", productID)
	if err := os.MkdirAll(productDir, 0755); err != nil {
		log.Printf("[media-service] Failed to create product directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	// Resize image to thumbnail dimensions (square crop from center)
	thumbImg := resizeThumbnail(img, thumbnailWidth)
	thumbName := fmt.Sprintf("thumb_%d%s", time.Now().UnixNano(), ext)
	thumbPath := filepath.Join(productDir, thumbName)

	thumbFile, err := os.Create(thumbPath)
	if err != nil {
		log.Printf("[media-service] Failed to create thumbnail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save thumbnail"})
		return
	}
	defer thumbFile.Close()

	// Encode thumbnail based on format
	switch format {
	case "jpeg":
		err = jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(thumbFile, thumbImg)
	default:
		err = jpeg.Encode(thumbFile, thumbImg, &jpeg.Options{Quality: 80})
	}

	if err != nil {
		os.Remove(thumbPath)
		log.Printf("[media-service] Failed to encode thumbnail: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode thumbnail"})
		return
	}

	// Get dimensions
	bounds := thumbImg.Bounds()

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Thumbnail created successfully",
		"product_id":   productID,
		"format":       format,
		"width":        bounds.Dx(),
		"height":       bounds.Dy(),
		"filename":     thumbName,
		"size":         getFileSize(thumbPath),
		"thumbnail_url": fmt.Sprintf("/api/v1/media/files/products/%s/%s", productID, thumbName),
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	})
}

// =========================================================================
// image processing helpers (standard library only)
// =========================================================================

// resizeImage scales an image to the given maxWidth while maintaining aspect ratio.
// If maxHeight is 0, it's calculated proportionally.
func resizeImage(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Calculate target dimensions maintaining aspect ratio
	targetW := maxWidth
	targetH := maxHeight

	if maxHeight == 0 {
		// Scale proportionally based on width
		if srcW <= maxWidth {
			return src // No resize needed
		}
		ratio := float64(maxWidth) / float64(srcW)
		targetH = int(float64(srcH) * ratio)
	}

	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	// Simple nearest-neighbor scaling (standard library only)
	xRatio := float64(srcW) / float64(targetW)
	yRatio := float64(srcH) / float64(targetH)

	for y := 0; y < targetH; y++ {
		for x := 0; x < targetW; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			dst.Set(x, y, src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

// resizeThumbnail creates a square thumbnail by cropping the center and resizing.
func resizeThumbnail(src image.Image, size int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	// Crop to square from center
	var cropRect image.Rectangle
	if srcW > srcH {
		// Landscape: crop horizontally
		offset := (srcW - srcH) / 2
		cropRect = image.Rect(bounds.Min.X+offset, bounds.Min.Y, bounds.Min.X+offset+srcH, bounds.Max.Y)
	} else {
		// Portrait or square: crop vertically
		offset := (srcH - srcW) / 2
		cropRect = image.Rect(bounds.Min.X, bounds.Min.Y+offset, bounds.Max.X, bounds.Min.Y+offset+srcW)
	}

	// Create square cropped image
	cropped := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
	draw.Draw(cropped, cropped.Bounds(), src, cropRect.Min, draw.Src)

	// Resize to target size
	return resizeImage(cropped, size, size)
}

// =========================================================================
// helper functions
// =========================================================================

// sanitizeFilename removes path components and special chars from a filename
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	// Replace spaces and special characters
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// getAllowedExtensionsList returns a comma-separated list of allowed extensions
func getAllowedExtensionsList() string {
	exts := make([]string, 0, len(allowedExtensions))
	for ext := range allowedExtensions {
		exts = append(exts, ext)
	}
	return strings.Join(exts, ", ")
}

// getFileSize returns the size of a file at the given path
func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

// ensureDB ensures the database connection is available.
// Returns a helpful error message if the connection is not initialized.
func ensureDB() (*sql.DB, error) {
	db := database.GetDB()
	if db == nil {
		return nil, fmt.Errorf("database connection not initialized")
	}
	return db, nil
}

// parseProductID parses a product ID string to int
func parseProductID(s string) (int, error) {
	return strconv.Atoi(s)
}

// getAverageColor returns the average color of an image (useful for placeholders).
func getAverageColor(img image.Image) color.Color {
	bounds := img.Bounds()
	var r, g, b, a uint64
	pixelCount := uint64(bounds.Dx() * bounds.Dy())

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, pa := img.At(x, y).RGBA()
			r += uint64(pr)
			g += uint64(pg)
			b += uint64(pb)
			a += uint64(pa)
		}
	}

	if pixelCount > 0 {
		return color.RGBA{
			R: uint8(r / pixelCount >> 8),
			G: uint8(g / pixelCount >> 8),
			B: uint8(b / pixelCount >> 8),
			A: uint8(a / pixelCount >> 8),
		}
	}
	return color.Black
}
