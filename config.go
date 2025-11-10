package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ConfigPart represents a single MPD configuration part.
type ConfigPart struct {
	Key      string `json:"key"`  // e.g., "exclusive"
	Name     string `json:"name"` // e.g., "Exclusive (DSD)"
	FullPath string `json:"-"`    // Full path to the file
}

// DiscoverModes scans the config directory for config part files and extracts their metadata.
func DiscoverModes(configDir string) ([]ConfigPart, error) {
	matches, err := filepath.Glob(filepath.Join(configDir, "config-*.mpd.conf.part"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob config parts: %w", err)
	}

	var parts []ConfigPart
	for _, match := range matches {
		fileName := filepath.Base(match)
		key := strings.TrimSuffix(strings.TrimPrefix(fileName, "config-"), ".mpd.conf.part")

		name := parsePartName(match, key)

		parts = append(parts, ConfigPart{
			Key:      key,
			Name:     name,
			FullPath: match,
		})
	}
	return parts, nil
}

// parsePartName reads the first line of a config part file to find the friendly name.
func parsePartName(filePath, defaultKey string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return strings.Title(defaultKey) // Fallback to title-cased key
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		firstLine := scanner.Text()
		re := regexp.MustCompile(`# ConfigPartName:\s*(.*)`)
		matches := re.FindStringSubmatch(firstLine)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return strings.Title(defaultKey) // Fallback to title-cased key
}

// handleGetConfigParts is the Fiber handler for GET /api/configparts
func handleGetConfigParts(c *fiber.Ctx) error {
	parts, err := DiscoverModes(config.ConfigDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to discover config parts: %v", err)})
	}
	return c.JSON(parts)
}
