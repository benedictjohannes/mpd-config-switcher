package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"embed"
	"flag"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// openBrowser tries to open the URL in a browser
func openBrowser(url string) bool {
	err := exec.Command("xdg-open", url).Start()
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
		return false
	}
	return true
}

//go:embed all:dist
var frontendFS embed.FS

type Config struct {
	Port            int
	FrontendPort    int // For dev proxy
	ConfigDir       string
	SystemdUnitName string
	UseSudo         bool
	Expose          bool
	OpenBrowser     bool
}

var config Config

func init() {
	flag.IntVar(&config.Port, "port", 6279, "Port for the backend server.")
	flag.IntVar(&config.FrontendPort, "fe-port", 0, "DEV ONLY: Port for the frontend dev server (enables reverse proxy).")
	flag.StringVar(&config.ConfigDir, "config-dir", "~/.config/mpd", "Directory containing mpd.conf parts.")
	flag.StringVar(&config.SystemdUnitName, "systemd-unit-name", "mpd.service", "The systemd unit to restart.")
	flag.BoolVar(&config.UseSudo, "sudo", false, "Use 'sudo systemctl' instead of 'systemctl --user'.")
	flag.BoolVar(&config.Expose, "expose", false, "Listen to all interfaces, so you can use this app from the comfort of other devices in your LAN.")
	flag.BoolVar(&config.OpenBrowser, "open", false, "When true, launch the main browser to the app URL.")
	flag.Parse()

	// Resolve home dir if '~' is used in config-dir
	if strings.HasPrefix(config.ConfigDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home directory: %v", err)
		}
		config.ConfigDir = strings.Replace(config.ConfigDir, "~", homeDir, 1)
	}
}

// executeCommand runs a shell command in the specified directory
func executeCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = config.ConfigDir // Set the working directory for the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s %s, error: %w, output: %s", command, strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// handleCurrentMode responds with the current MPD mode
func handleCurrentMode(c *fiber.Ctx) error {
	mainMpdConf := config.ConfigDir + "/mpd.conf"
	content, err := os.ReadFile(mainMpdConf)
	if err != nil {
		log.Printf("Error reading %s: %v", mainMpdConf, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"key": "unknown", "name": "<Unknown>"})
	}

	// Read the first line to find the # CurrentConfig: <modeKey> comment
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	if scanner.Scan() {
		firstLine := scanner.Text()
		re := regexp.MustCompile(`# CurrentConfig:\s*(.*)`)
		matches := re.FindStringSubmatch(firstLine)
		if len(matches) > 1 {
			modeKey := strings.TrimSpace(matches[1])
			parts, err := DiscoverModes(config.ConfigDir)
			if err != nil {
				log.Printf("Error discovering modes: %v", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"key": "unknown", "name": "<Unknown>"})
			}
			for _, part := range parts {
				if part.Key == modeKey {
					return c.JSON(part)
				}
			}
		}
	}

	return c.JSON(fiber.Map{"key": "unknown", "name": "<Unknown>"})
}

// handleSwitchMode switches the MPD configuration and restarts the service
func handleSwitchMode(c *fiber.Ctx) error {
	targetModeKey := c.Params("mode")

	parts, err := DiscoverModes(config.ConfigDir)
	if err != nil {
		log.Printf("Error discovering modes: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to discover config parts."})
	}

	var targetPart ConfigPart
	found := false
	for _, part := range parts {
		if part.Key == targetModeKey {
			targetPart = part
			found = true
			break
		}
	}

	if !found {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Invalid target mode specified: %s", targetModeKey)})
	}

	log.Printf("Attempting to switch to %s mode...", targetPart.Name)

	baseConfPartPath := filepath.Join(config.ConfigDir, "base.mpd.conf.part")
	mainMpdConfPath := filepath.Join(config.ConfigDir, "mpd.conf")

	// Read base config and target part, then combine
	baseContent, err := os.ReadFile(baseConfPartPath)
	if err != nil {
		log.Printf("Error reading base config part: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read base MPD configuration part."})
	}

	targetPartContent, err := os.ReadFile(targetPart.FullPath)
	if err != nil {
		log.Printf("Error reading target config part (%s): %v", targetPart.FullPath, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read target MPD configuration part."})
	}

	newMpdConfContent := fmt.Sprintf("# CurrentConfig: %s\n%s\n%s", targetPart.Key, baseContent, targetPartContent)

	// Write the new mpd.conf
	err = os.WriteFile(mainMpdConfPath, []byte(newMpdConfContent), 0644) // 0644 for read/write by owner, read by others
	if err != nil {
		log.Printf("Error writing new mpd.conf: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write new MPD configuration."})
	}
	log.Printf("Successfully wrote new mpd.conf for %s mode.", targetPart.Name)

	// Restart MPD user service
	err = restartMPD()
	if err != nil {
		log.Printf("Error restarting MPD service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to restart MPD service: %v", err)})
	}
	log.Printf("MPD service restarted successfully for %s mode.", targetPart.Name)

	return c.JSON(fiber.Map{"message": fmt.Sprintf("MPD successfully switched to %s mode.", targetPart.Name)})
}

// restartMPD restarts the MPD systemd service
func restartMPD() error {
	var cmdArgs []string
	if config.UseSudo {
		cmdArgs = []string{"sudo", "systemctl", "restart", config.SystemdUnitName}
	} else {
		cmdArgs = []string{"systemctl", "--user", "restart", config.SystemdUnitName}
	}

	_, err := executeCommand(cmdArgs[0], cmdArgs[1:]...)
	return err
}

func main() {
	app := fiber.New()
	api := app.Group("/api")
	api.Get("/currentmode", handleCurrentMode)
	api.Get("/configparts", handleGetConfigParts)
	api.Get("/switch/:mode", handleSwitchMode)

	// Note: This must be defined *before* the /api group.
	if config.FrontendPort > 0 {
		// DEV MODE: Proxy to fe-port
		proxyURL := fmt.Sprintf("http://localhost:%d", config.FrontendPort)
		// Proxy all non-API routes to the frontend dev server
		app.Use(func(c *fiber.Ctx) error {
			if strings.HasPrefix(c.Path(), "/api") {
				return c.Next()
			}
			return proxy.Forward(proxyURL)(c)
		})
	} else {
		// PROD MODE: Serve embedded files
		app.Use("/", filesystem.New(filesystem.Config{
			Root:       http.FS(frontendFS),
			PathPrefix: "dist",
			Browse:     true,
			Index:      "index.html",
		}))
	}

	listenAddr := fmt.Sprintf(":%d", config.Port)
	if config.Expose {
		listenAddr = fmt.Sprintf("0.0.0.0:%d", config.Port)
	}

	log.Printf("MPD Switcher Go backend listening on %s", listenAddr)

	// Open browser if the flag is set
	if config.OpenBrowser {
		appURL := fmt.Sprintf("http://localhost:%d", config.Port)
		if config.Expose {
			// Try to get local IP if exposed, otherwise use localhost
			// This is a simplification; a real implementation might need to discover the actual IP
			appURL = fmt.Sprintf("http://%s:%d", "localhost", config.Port)
		}
		log.Printf("Opening browser to %s", appURL)
		go func() {
			// Give the server a moment to start before opening the browser
			// This is a simple delay, a more robust solution might poll the server
			time.Sleep(1 * time.Second)
			if !openBrowser(appURL) {
				log.Printf("Please open your web browser and navigate to %s", appURL)
			}
		}()
	}

	log.Fatal(app.Listen(listenAddr))
}
