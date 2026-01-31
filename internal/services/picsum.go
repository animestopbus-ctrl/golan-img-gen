package services

import (
	"fmt"
	"io"
	"net/http"

	"github.com/yourusername/image-generator-bot/internal/utils"
)

// FetchPicsumImage gets random 1920x1080 image
func FetchPicsumImage() ([]byte, error) {
	url := "https://picsum.photos/1920/1080"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Picsum error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	utils.LogInfo("Picsum image fetched")
	return body, nil
}