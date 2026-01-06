package main

import (
	"crypto/rand"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type SpeedTestController struct {
	SampleDataPath string
}

func main() {
	// Konfiguráció (Spring @Value megfelelője környezeti változóból)
	dataPath := os.Getenv("SAMPLE_DATA_PATH")
	if dataPath == "" {
		dataPath = "/home/pzoli/temp/test.bin" // Alapértelmezett érték
	}

	controller := &SpeedTestController{
		SampleDataPath: dataPath,
	}

	router := gin.Default()

	// CORS beállítás javítása
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Engedélyezett címek listája
		allowedOrigins := map[string]bool{
			"http://localhost:4200":             true,
			"http://lenovo.me.local:4200": true,
		}

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		// Itt adjuk hozzá a Cache-Control és Pragma fejléceket az engedélyezettek listájához
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cache-Control, Pragma")

		// A Preflight (OPTIONS) kérések azonnali megválaszolása
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Végpontok definiálása
	api := router.Group("/api/speedtest")
	{
		api.GET("/init", controller.InitSampleData)
		api.POST("/upload", controller.Upload)
		api.GET("/stream-download", controller.DownloadStream)
	}

	router.Run(":8080")
}

// InitSampleData generálja a 100MB-os fájlt
func (ctrl *SpeedTestController) InitSampleData(c *gin.Context) {
	if _, err := os.Stat(ctrl.SampleDataPath); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "File already exists"})
		return
	}

	size := int64(100 * 1024 * 1024)
	data := make([]byte, size)
	_, err := rand.Read(data) // Kriptográfiailag biztonságos véletlenszerű adat
	if err != nil {
		c.String(http.StatusInternalServerError, "Error generating data")
		return
	}

	err = os.WriteFile(ctrl.SampleDataPath, data, 0644)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error saving file")
		return
	}

	c.String(http.StatusOK, "Initialization completed")
}

// Upload fogadja a feltöltött adatokat
func (ctrl *SpeedTestController) Upload(c *gin.Context) {
	// Beolvassuk a body-t (Spring @RequestBody byte[] data megfelelője)
	_, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Error reading body")
		return
	}
	c.String(http.StatusOK, "Upload successful")
}

// DownloadStream kiszolgálja a fájlt streamként
func (ctrl *SpeedTestController) DownloadStream(c *gin.Context) {
	file, err := os.Open(ctrl.SampleDataPath)
	if err != nil {
		c.String(http.StatusNotFound, "File not found. Run /init first.")
		return
	}
	defer file.Close()

	c.Header("Content-Disposition", "attachment; filename=test.bin")
	c.Header("Content-Type", "application/octet-stream")

	// A Go io.Copy vagy a Gin FileStream rendkívül hatékony
	c.File(ctrl.SampleDataPath)
}
