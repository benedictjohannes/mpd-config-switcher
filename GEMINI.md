# Project: MPD Switcher v2 - Refactor Plan

## 1. Project Goal

Refactor the existing hard-coded Go/Fiber + React MPD switcher into a flexible, extensible, and distributable open-source tool. The tool will be a single, self-contained Go binary that serves a dynamic web UI based on user-provided configuration files.

## 2. Core Architectural Changes

1. **Dynamic Config Discovery**: The backend will no longer hard-code modes. It will scan a directory for configuration "parts."
   - Base Config: `base.mpd.conf.part` 
   - Mode Parts: The app will glob for files matching config-*.mpd.conf.part. The * will be the unique key for the mode (e.g., exclusive, pipewire).
2. Mode Metadata: The app will parse the first line of each config-*.mpd.conf.part file for a special comment to get the "Friendly Name." 
   - Syntax: `# ConfigPartName: Exclusive (DSD)`
   - If the comment is missing, the app will fall back to using the key (e.g., "exclusive").
3. Dynamic API: The frontend will be populated by API calls, not hard-coded values.
4. Single Binary Deployment: The entire built React frontend will be embedded into the Go binary using //go:embed.
5. CLI-Driven Configuration: All hard-coded paths and values will be replaced with CLI flags to allow for flexible deployment.

## 3. Backend Tasks (Go / Fiber)

### 3.1. Refactor main.go for CLI Flags

- Import the flag package.
- Define a Config struct to hold all settings.
- Populate the Config struct from CLI flags (or environment variables).

```go
type Config struct {
    Port            int
    FrontendPort    int    // For dev proxy
    ConfigDir       string
    SystemdUnitName string
    UseSudo         bool
}
var config Config

func init() {
    flag.IntVar(&config.Port, "port", 56737, "Port for the backend server.")
    flag.IntVar(&config.FrontendPort, "fe-port", 0, "DEV ONLY: Port for the frontend dev server (enables reverse proxy).")
    flag.StringVar(&config.ConfigDir, "config-dir", "/home/user/.config/mpd", "Directory containing mpd.conf parts.")
    flag.StringVar(&config.SystemdUnitName, "systemd-unit-name", "mpd.service", "The systemd unit to restart.")
    flag.BoolVar(&config.UseSudo, "sudo", false, "Use 'sudo systemctl' instead of 'systemctl --user'.")
    flag.Parse()

    // Resolve home dir if '~' is used in config-dir (optional but nice)
    // ...
}
```

### 3.2. Implement //go:embed for Frontend

Embed the frontend/dist directory.

1. Create a Fiber static handler to serve the embedded files.
```go
import "embed"

//go:embed all:frontend/dist
var frontendFS embed.FS

// In main(), after app := fiber.New():
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
        PathPrefix: "frontend/dist",
        Browse:     true,
        Index:      "index.html",
    }))
}
```
2. API routes must be defined after the proxy/static handlers.
```go
api := app.Group("/api")
api.Get("/currentmode", handleCurrentMode)
api.Get("/configparts", handleGetConfigParts)
api.Get("/switch/:modeKey", handleSwitchMode)
// ...
```
### 3.3. Create a Dynamic Config Service (`config.go`)

1. ConfigPart Struct:
```go
type ConfigPart struct {
    Key  string `json:"key"`  // e.g., "exclusive"
    Name string `json:"name"` // e.g., "Exclusive (DSD)"
    Path string `json:"-"`  // Full path to the file
}
```
2. `DiscoverModes(configDir string)` function:
    - Glob for `configDir `"/config-*.mpd.conf.part"`.
    - Loop through results, parse `Key` from filename.
    - Call `parsePartName()` for each file to get the Name.
    - Return `[]ConfigPart`.
3. `parsePartName(filePath string)` function:
    - Open `filePath`, read the first line.
    - Use a regex (`# ConfigPartName:(.*)`) to find the name, trimming white space.
      - If no match, return a default name (e.g., title-cased Key).
4. `handleGetConfigParts` (Handler for `GET /api/configparts`):
    - Calls `DiscoverModes(config.ConfigDir)`.
    - Returns the `[]ConfigPart` as JSON.

### 3.4. Update API Handlers
1. `handleCurrentMode`:
   - Refactor to use config.ConfigDir.
   - Read the mainMpdConf file.
   - Read the first line to find the `# CurrentConfig: <modeKey> comment`.
     - If found, call `DiscoverModes(config.ConfigDir)` and find the matching ConfigPart struct.
     - Return the full ConfigPart object (e.g., `{"key": "exclusive", "name": "Exclusive (DSD)"}`) as JSON.
     - If not found or error, return a `{"key": "unknown", "name": "<Unknown>"}` object.
2. `handleSwitchMode` (`GET /api/switch/:modeKey`):
    - Use `modeKey` param to find the correct `ConfigPart` from the discovered list.
    - Refactor to use `config.ConfigDir` to read `base.mpd.conf.part` and the correct `config-*.mpd.conf.part`.
    - Create the new config content starting with the line `# CurrentConfig: <modeKey>`.
    - Append the `baseContent` and `targetPartContent`.
    - Write the new `mpd.conf` file.
    - Refactor `executeCommand` to call the new `restartMPD()` function.
3. `restartMPD()` function:
   - Check `config.UseSudo`.
   - If `true`, command is `sudo systemctl restart [config.SystemdUnitName]`.
   - If `false`, command is `systemctl --user restart [config.SystemdUnitName]`.

## 4. Frontend Tasks (React / Rsbuild)

### 4.1. Refactor `App.jsx` into `App.tsx`

1. Use typescript and define API types in `types.ts`
2. Remove hard-coded buttons.
3. Update State: `currentMode` will now hold an object, not a string.
```tsx
const [configParts, setConfigParts] = useState([]);
const [currentMode, setCurrentMode] = useState({ key: '', name: 'Loading...' });
```
4. Update fetchCurrentMode:
 - It will now fetch /api/currentmode.
  - On success: `const data = await response.json(); setCurrentMode(data);`
  - On error: `setCurrentMode({ key: 'unknown', name: 'Unknown' });`
5. Update useEffect on mount:
  - Call fetch('/api/configparts') and populate setConfigParts.
  - Call fetchCurrentMode().
6. Render Current Mode Dynamically: 
```tsx
<p className={`text-3xl font-bold ...`}>
    {currentMode.name}
</p>
```
7. Render buttons dynamically:
```tsx
<div className="flex flex-col space-y-4 mb-6">
    {configParts.map((part) => (
        <button
            key={part.key}
            onClick={() => switchMode(part.key, part.name)} // Pass name for optimistic message
            disabled={loading || currentMode.key === part.key} // Correctly compare keys
            className={`...`}
        >
            {loading && /* logic */ ? 'Switching...' : `Switch to ${part.name}`}
        </button>
    ))}
</div>
```
8. Update switchMode(modeKey, modeName) function:
   - The fetch URL is dynamic: `/api/switch/${modeKey}`.
   - The "optimistic" `setMessage` can now use `modeName`: `setMessage("Switching to "+ modeName "configuration...")`


## 5. Build & Documentation Tasks

### Create Makefile:
```Makefile
.PHONY: all build build-fe build-be

all: build

build: build-fe build-be

build-fe:
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Frontend build complete."

build-be:
	@echo "Building backend..."
	go build -tags netgo -ldflags "-s -w" -o mpd-switcher .
	@echo "Backend build complete: ./mpd-switcher"
```

### Update README.md to real documentation instead of rsbuild explainer
- **Rationale**: Explain that: some MPD configuration need to be changed on the MPD configuration (ie, switching exclusive mode for DSD native playback support) vs. PipeWire-shared), but MPD is  configured using configuration files. Currently the project aims to dynamically switch MPD configuration "parts" via an easy GUI, so the user should already know how to configure MPD.
- **Features**: (Dynamic config, web UI, single binary, etc.)
- **Installation**: (Download binary from GitHub Releases, chmod +x).
- **Configuration**: Explain the `base.mpd.conf.part` and `config-*.mpd.conf.part` structure, including the `# ConfigPartName: comment`.
- **Usage**: Explain all CLI flags (`--port`, `--config-dir`, etc.).
- **Disclaimer**: Explain that: This project started for my personal use, as I find it extremely cumbersome to switch between mpd’s “Exclusive mode” that can play DSD natively and “Shared mode” that coexist naturally with other Pipewire (or Pulse Audio) programs.
- **Development**: Explain the stack (Go/Fiber, React/Rsbuild), how to install dependencies, and how to use `make` or `go run . --fe-port=3000` in parallel with `npm run start` for development.