// package monitor provides functionality to check for updates to Docker images
package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	registry_monitor "github.com/MadhavKrishanGoswami/Lighthouse/services/common/genproto/registry-monitor"
)

const (
	// Note: Removed non-breaking space characters for consistency.
	dockerAuthURL     = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"
	dockerRegistryURL = "https://registry-1.docker.io/v2/%s/manifests/%s"
)

// authResponse is a private struct to unmarshal the JSON response from the Docker auth server.
type authResponse struct {
	Token string `json:"token"`
}

// getAuthToken fetches an anonymous, read-only token for a given Docker Hub repository.
func getAuthToken(repository string) (string, error) {
	// Format the URL with the specific repository we want to access.
	url := fmt.Sprintf(dockerAuthURL, repository)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make auth request to Docker Hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("docker auth request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read auth response body: %w", err)
	}

	var authResp authResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal auth token: %w", err)
	}

	return authResp.Token, nil
}

// getImageDigest fetches the unique digest for a specific image tag.
func getImageDigest(repository, tag, token string) (string, error) {
	// Format the registry URL with the repository and tag.
	url := fmt.Sprintf(dockerRegistryURL, repository, tag)

	// Create a new HTTP request.
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for manifest: %w", err)
	}

	// Set the required headers for the Docker Registry API.
	req.Header.Set("Authorization", "Bearer "+token)
	// This header tells the registry we accept the standard manifest format.
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	// Execute the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch manifest for %s:%s: %w", repository, tag, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get manifest for %s:%s with status: %s", repository, tag, resp.Status)
	}

	// The digest is returned in the 'Docker-Content-Digest' header.
	digest := resp.Header.Get("Docker-Content-Digest")
	if digest == "" {
		return "", fmt.Errorf("could not find digest for %s:%s", repository, tag)
	}

	return digest, nil
}

// Monitor checks a list of Docker images to see if a newer version tagged as 'latest' is available.
func Monitor(checkforupdates *registry_monitor.CheckUpdatesRequest) (*registry_monitor.CheckUpdatesResponse, error) {
	// Initialize the response object.
	var response registry_monitor.CheckUpdatesResponse

	// Loop through each image provided in the request.
	for _, image := range checkforupdates.Images {
		repoName := image.Repository
		if !strings.Contains(repoName, "/") {
			repoName = "library/" + repoName
		}

		// Step 1: Get the authentication token for the repository.
		token, err := getAuthToken(repoName)
		if err != nil {
			// If we can't get a token for one image, we can log it and continue with the others.
			fmt.Printf("Could not get auth token for %s. Skipping. Error: %v\n", repoName, err)
			continue
		}

		// Step 2: Get the digest of the user's current image tag.
		currentDigest, err := getImageDigest(repoName, image.Tag, token)
		if err != nil {
			// Could not find the specific tag the user has. Maybe it was deleted.
			fmt.Printf("Could not get digest for current image %s:%s. Skipping. Error: %v\n", repoName, image.Tag, err)
			continue
		}

		// Step 3: Get the digest of the 'latest' tag for comparison.
		latestDigest, err := getImageDigest(repoName, "latest", token)
		if err != nil {
			// The repository might not have a 'latest' tag.
			fmt.Printf("Could not get digest for latest tag on %s. Skipping. Error: %v\n", repoName, err)
			continue
		}

		// Step 4: Compare the digests. If they are different, an update is available.
		if currentDigest != latestDigest {
			// An update is found. Populate the update information.
			updateInfo := registry_monitor.ImagetoUpdate{
				ContainerId: image.ContainerId,
				Description: fmt.Sprintf("Update available for %s:%s. New version 'latest' is available.", image.Repository, image.Tag),
				Timestamp:   time.Now().Unix(),
			}
			// Add the found update to our response list.
			response.ImagestoUpdate = append(response.ImagestoUpdate, &updateInfo)
		}
	}

	// Return the list of found updates and no error.
	return &response, nil
}
