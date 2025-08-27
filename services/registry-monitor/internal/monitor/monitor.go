// Package monitor is a package
package monitor

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
)

const (
	dockerAuthURL     = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"
	dockerRegistryURL = "https://registry-1.docker.io/v2/%s/manifests/%s"
	cacheTTL          = 10 * time.Minute // Cache results for 5 minutes.
)

// authResponse stores the token from the Docker auth server.
type authResponse struct {
	Token string `json:"token"`
}

// cachedDigest holds the cached digest for a repository's 'latest' tag.
type cachedDigest struct {
	digest    string
	expiresAt time.Time
}

// Global cache and a mutex to prevent race conditions.
var (
	digestCache = make(map[string]cachedDigest)
	cacheMutex  = &sync.Mutex{}
	// Use a shared, configured client for all HTTP requests.
	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

// getAuthToken fetches a read-only token for a Docker Hub repository.
func getAuthToken(repository string) (string, error) {
	url := fmt.Sprintf(dockerAuthURL, repository)
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed with status: %s", resp.Status)
	}

	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal auth token: %w", err)
	}
	return authResp.Token, nil
}

// getLatestDigest fetches the digest for the "latest" tag of a repository.
// It correctly handles multi-arch manifests and falls back to calculating the digest if the header is missing.
func getLatestDigest(repository, token string) (string, error) {
	// --- Step 1: Check the cache first ---
	cacheMutex.Lock()
	cached, found := digestCache[repository]
	cacheMutex.Unlock()

	// if found but expired then just delete it
	if found && !time.Now().Before(cached.expiresAt) {
		// ...delete it and pretend it wasn't found.
		delete(digestCache, repository)
		found = false
	}
	cacheMutex.Unlock()

	if found && time.Now().Before(cached.expiresAt) {
		return cached.digest, nil
	}

	// --- Step 2: Fetch the manifest for the "latest" tag ---
	manifestURL := fmt.Sprintf(dockerRegistryURL, repository, "latest")
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create manifest request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	// ✅ Correctly accept both single-arch and multi-arch manifest lists.
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute manifest request: %w", err)
	}
	// ✅ Correctly defer Body.Close() immediately after checking for a nil response.
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("manifest request for '%s' failed with status: %s", repository, resp.Status)
	}

	// --- Step 3: Get the digest ---
	// Prioritize the header, as it's the most reliable source.
	latestDigest := resp.Header.Get("Docker-Content-Digest")
	if latestDigest == "" {
		// ✅ Fallback: If the header is missing, calculate the SHA256 digest of the body.
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read manifest body for digest calculation: %w", err)
		}
		latestDigest = fmt.Sprintf("sha256:%x", sha256.Sum256(bodyBytes))
	}

	// --- Step 4: Update the cache ---
	cacheMutex.Lock()
	digestCache[repository] = cachedDigest{
		digest:    latestDigest,
		expiresAt: time.Now().Add(cacheTTL),
	}
	cacheMutex.Unlock()

	return latestDigest, nil
}

// Monitor checks a list of Docker images concurrently to find available updates.
func Monitor(checkforupdates *registry_monitor.CheckUpdatesRequest) (*registry_monitor.CheckUpdatesResponse, error) {
	var (
		wg sync.WaitGroup
		// Use a buffered channel to prevent goroutines from blocking.
		updates  = make(chan *registry_monitor.ImagetoUpdate, len(checkforupdates.Images))
		response = &registry_monitor.CheckUpdatesResponse{}
	)

	for _, image := range checkforupdates.Images {
		wg.Add(1)
		// Pass the image by value to the goroutine to avoid loop variable capture issues.
		go func(img *registry_monitor.ImageInfo) {
			defer wg.Done()
			repoName := img.Repository
			if !strings.Contains(repoName, "/") {
				repoName = "library/" + repoName
			}

			token, err := getAuthToken(repoName)
			if err != nil {
				log.Printf("ERROR: Could not get auth token for %s: %v", repoName, err)
				return
			}

			latestDigest, err := getLatestDigest(repoName, token)
			if err != nil {
				log.Printf("ERROR: Could not get latest digest for %s: %v", repoName, err)
				return
			}

			if img.Digest != latestDigest {
				log.Printf("INFO: Update found for %s. Current: %s, Latest: %s", repoName, img.Digest, latestDigest)
				updates <- &registry_monitor.ImagetoUpdate{
					ContainerUid: img.ContainerUid,
					NewTag:       "latest", // The new tag is simply "latest".
					Description:  fmt.Sprintf("Update available for %s. New digest: %s", img.Repository, latestDigest),
					Timestamp:    time.Now().Unix(),
				}
			}
		}(image)
	}

	// Wait for all checks to complete.
	wg.Wait()
	// Close the channel to signal that no more updates will be sent.
	close(updates)

	for update := range updates {
		response.ImagestoUpdate = append(response.ImagestoUpdate, update)
	}

	return response, nil
}
