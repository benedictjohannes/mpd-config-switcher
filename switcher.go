package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Define constants for MPD configuration paths
const (
	mpdConfigDir        = "/home/benedict/.config/mpd"
	mainMpdConf         = mpdConfigDir + "/mpd.conf"
	baseConfPart        = mpdConfigDir + "/base.mpd.conf.part"
	exclusiveConfPart   = mpdConfigDir + "/out-exclusive.mpd.conf.part"
	pipewireConfPart    = mpdConfigDir + "/out-pipewire.mpd.conf.part"
)

// executeCommand runs a shell command in the specified directory
func executeCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = mpdConfigDir // Set the working directory for the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s %s, error: %w, output: %s", command, strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// getCurrentMpdMode reads mpd.conf and determines the current mode
func getCurrentMpdMode() (string, error) {
	content, err := ioutil.ReadFile(mainMpdConf)
	if err != nil {
		return "Unknown", fmt.Errorf("failed to read %s: %w", mainMpdConf, err)
	}

	if strings.Contains(string(content), "exclusive") {
		return "Exclusive (DSD)", nil
	} else if strings.Contains(string(content), "pipewire") {
		return "PipeWire", nil
	}
	return "Unknown", nil
}

// handleCurrentMode responds with the current MPD mode
func handleCurrentMode(c *fiber.Ctx) error {
	mode, err := getCurrentMpdMode()
	if err != nil {
		log.Printf("Error getting current mode: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to determine current MPD mode"})
	}
	return c.JSON(fiber.Map{"mode": mode})
}

// handleSwitchMode switches the MPD configuration and restarts the service
func handleSwitchMode(c *fiber.Ctx) error {
	targetMode := c.Params("mode")

	var targetConfPart string
	var modeName string

	switch targetMode {
	case "exclusive":
		targetConfPart = exclusiveConfPart
		modeName = "Exclusive (DSD)"
	case "pipewire":
		targetConfPart = pipewireConfPart
		modeName = "PipeWire"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid target mode specified. Use 'exclusive' or 'pipewire'."})
	}

	log.Printf("Attempting to switch to %s mode...", modeName)

	// Read base config and target part, then combine
	baseContent, err := ioutil.ReadFile(baseConfPart)
	if err != nil {
		log.Printf("Error reading base config part: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read base MPD configuration part."})
	}

	targetPartContent, err := ioutil.ReadFile(targetConfPart)
	if err != nil {
		log.Printf("Error reading target config part (%s): %v", targetConfPart, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read target MPD configuration part."})
	}

	newMpdConfContent := fmt.Sprintf("%s\n%s", baseContent, targetPartContent)

	// Write the new mpd.conf
	err = ioutil.WriteFile(mainMpdConf, []byte(newMpdConfContent), 0644) // 0644 for read/write by owner, read by others
	if err != nil {
		log.Printf("Error writing new mpd.conf: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write new MPD configuration."})
	}
	log.Printf("Successfully wrote new mpd.conf for %s mode.", modeName)

	// Restart MPD user service
	_, err = executeCommand("systemctl", "--user", "restart", "mpd.service")
	if err != nil {
		log.Printf("Error restarting MPD service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to restart MPD service: %v", err)})
	}
	log.Printf("MPD service restarted successfully for %s mode.", modeName)

	return c.JSON(fiber.Map{"message": fmt.Sprintf("MPD successfully switched to %s mode.", modeName)})
}

func main() {
	app := fiber.New()

	app.Get("/api/currentmode", handleCurrentMode)
	app.Get("/api/switch/:mode", handleSwitchMode)

	portStr := os.Getenv("PORT")
	port := 56737 // Default port
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("Warning: Could not parse PORT environment variable '%s': %v", portStr, err)
		} else if p > 0 && p <= 65535 {
			port = p
		} else {
			log.Printf("Warning: Invalid port number '%s' from environment variable. Using default port.", portStr)
		}
	}
	log.Printf("MPD Switcher Go backend listening on :%d", port)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
