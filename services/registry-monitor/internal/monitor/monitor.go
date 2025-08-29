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
	cacheTTL          = 30 * time.Minute // Cache results for 30 minutes
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
	cacheMutex  = &sync.RWMutex{} // Use RWMutex for better read performance
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
	cacheMutex.RLock()
	cached, found := digestCache[repository]
	cacheMutex.RUnlock()

	// Check if cache entry is valid
	if found && time.Now().Before(cached.expiresAt) {
		return cached.digest, nil
	}

	// Clean up expired entry if found
	if found && !time.Now().Before(cached.expiresAt) {
		cacheMutex.Lock()
		delete(digestCache, repository)
		cacheMutex.Unlock()
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

// getCurrentDigest fetches the current digest for a specific image tag.
// This is used when we need to determine the actual digest of a running container.
func getCurrentDigest(repository, tag, token string) (string, error) {
	if tag == "" {
		tag = "latest"
	}

	manifestURL := fmt.Sprintf(dockerRegistryURL, repository, tag)
	req, err := http.NewRequest("GET", manifestURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create manifest request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.docker.distribution.manifest.list.v2+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute manifest request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("manifest request for '%s:%s' failed with status: %s", repository, tag, resp.Status)
	}

	// Get the digest from the header (most reliable)
	currentDigest := resp.Header.Get("Docker-Content-Digest")
	if currentDigest == "" {
		// Fallback: calculate SHA256 of the manifest body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read manifest body for digest calculation: %w", err)
		}
		currentDigest = fmt.Sprintf("sha256:%x", sha256.Sum256(bodyBytes))
	}

	return currentDigest, nil
}

// Monitor checks a list of Docker images concurrently to find available updates.
func Monitor(checkforupdates *registry_monitor.CheckUpdatesRequest) (*registry_monitor.CheckUpdatesResponse, error) {
	if checkforupdates == nil || len(checkforupdates.Images) == 0 {
		return &registry_monitor.CheckUpdatesResponse{}, nil
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex // Protect the slice from concurrent writes
		response = &registry_monitor.CheckUpdatesResponse{
			ImagestoUpdate: make([]*registry_monitor.ImagetoUpdate, 0),
		}
	)

	// Process images concurrently
	for _, image := range checkforupdates.Images {
		wg.Add(1)
		// Pass the image by value to the goroutine to avoid loop variable capture issues.
		go func(img *registry_monitor.ImageInfo) {
			defer wg.Done()

			// Validate input
			if img == nil || img.Repository == "" {
				log.Printf("WARN: Skipping invalid image info: %+v", img)
				return
			}

			repoName := img.Repository
			if !strings.Contains(repoName, "/") {
				repoName = "library/" + repoName
			}

			// Get auth token
			token, err := getAuthToken(repoName)
			if err != nil {
				log.Printf("ERROR: Could not get auth token for %s: %v", repoName, err)
				return
			}

			// Get the most recent digest for the "latest" tag from the registry.
			latestDigest, err := getLatestDigest(repoName, token)
			if err != nil {
				log.Printf("ERROR: Could not get latest digest for %s: %v", repoName, err)
				return
			}

			// Determine the current digest of the container.
			var currentDigest string
			// The digest was not provided, so we resolve the provided tag to its current digest.
			log.Printf("DEBUG: Digest not provided for %s. Resolving tag '%s'.", repoName, img.Tag)
			resolvedDigest, err := getCurrentDigest(repoName, img.Tag, token)
			if err != nil {
				log.Printf("ERROR: Could not resolve current digest for %s:%s: %v", repoName, img.Tag, err)
				return // Cannot proceed without a current digest to compare against.
			}
			currentDigest = resolvedDigest

			// Compare digests - ONLY add if there's an actual update
			if currentDigest != latestDigest {
				log.Printf("INFO: Update found for %s. Current: %s, Latest: %s",
					repoName,
					truncateDigest(currentDigest),
					truncateDigest(latestDigest))

				update := &registry_monitor.ImagetoUpdate{
					ContainerUid: img.ContainerUid,
					NewTag:       fmt.Sprintf("%s:latest", img.Repository),
					Description: fmt.Sprintf("Update available for %s. Current: %s, New: %s",
						img.Repository,
						truncateDigest(currentDigest),
						truncateDigest(latestDigest)),
					Timestamp: time.Now().Unix(),
				}

				// Thread-safe append to response
				mu.Lock()
				response.ImagestoUpdate = append(response.ImagestoUpdate, update)
				mu.Unlock()
			} else {
				log.Printf("DEBUG: No update needed for %s (digests match: %s)",
					repoName,
					truncateDigest(currentDigest))
			}
		}(image)
	}

	// Wait for all checks to complete
	wg.Wait()

	log.Printf("INFO: Checked %d images, found %d updates", len(checkforupdates.Images), len(response.ImagestoUpdate))
	return response, nil
}

// truncateDigest is a helper function to safely truncate digests for logging
func truncateDigest(digest string) string {
	if len(digest) > 12 {
		return digest[:12] + "..."
	}
	return digest
}
