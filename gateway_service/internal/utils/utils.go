package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func TransformMinioURL(internalURL string) string {
	if internalURL == "" {
		return ""
	}

	internalEndpoint := viper.GetString("MINIO_ENDPOINT")
	publicEndpoint := viper.GetString("MINIO_PUBLIC_ENDPOINT")

	if internalEndpoint == "" && publicEndpoint == "" {
		return internalURL
	}

	url := internalURL

	normalizedInternal := strings.Replace(internalEndpoint, "http://", "", 1)
	normalizedInternal = strings.Replace(normalizedInternal, "https://", "", 1)

	normalizedPublic := strings.Replace(publicEndpoint, "http://", "", 1)
	normalizedPublic = strings.Replace(normalizedPublic, "https://", "", 1)

	url = strings.Replace(url, normalizedInternal, normalizedPublic, 1)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(publicEndpoint, "https://") {
			url = "https://" + url
		} else {
			url = "http://" + url
		}
	}

	return url
}

func TransformShareURL(shareUUID string) string {
	if shareUUID == "" {
		return ""
	}

	baseURL := viper.GetString("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	baseURL = strings.TrimSuffix(baseURL, "/")

	return fmt.Sprintf("%s/share/%s", baseURL, shareUUID)
}
