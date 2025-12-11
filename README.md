# A Config Switcher and Restarter for `mpd`

 `mpd`' is configurable via its configuration (commonly `mpd.conf`). A change in the configuration file would require a restart of `mpd`. As such, switching between `mpd` output of "ALSA" exclusive mode that play DSD natively and "Shared mode" that coexist naturally with other Pipewire (or Pulse Audio) programs, is quite cumbersome. This tool aims to dynamically switch `mpd` configuration "parts" via an easy web GUI, with
-   **Dynamic Config Discovery**: The backend scans a configurable directory for `mpd` configuration "parts" (`{mpdConfigDir}/config-*.mpd.conf.part`).
-   **Single binary**: Package as a go app with a web GUI. You can switch your `mpd` configuration with your phone!

## Installation

The easiest way to get the `mpd-config-switcher` is to download a pre-built binary from the [GitHub Releases](https://github.com/benedictjohannes/mpd-config-switcher/releases) page. Binaries are available for Linux, supporting both `amd64` and `arm64` architectures.

1.  **Download the binary**: Choose the appropriate binary for your operating system and architecture.
    *   `mpd-config-switcher-linux-amd64` (for 64-bit Linux)
    *   `mpd-config-switcher-linux-arm64` (for ARM64 Linux)
2.  **Make it executable**: `chmod +x mpd-config-switcher-<os>-<arch>` (replace `<os>-<arch>` with the actual name).
3.  **Place it**: Move the binary to a directory in your PATH (e.g., `/usr/local/bin`).

## Building from Source

If you prefer to build the `mpd-config-switcher` from source, follow these steps:

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/benedictjohannes/mpd-config-switcher.git
    cd mpd-switcher
    ```
2.  **Install Dependencies**:
    *   Go: Ensure Go (version 1.22 or higher) is installed.
    *   Node.js & npm: Ensure Node.js (version 20 or higher) and npm are installed.
3.  **Build the binary**:
    ```bash
    make
    ```
    This command will build the frontend, embed it into the Go binary, and produce a single `mpd-config-switcher` executable in the project root.


## Configuration

The MPD Config Switcher works by combining configuration "parts" into your main `mpd.conf`.

1.  **`base.mpd.conf.part`**: This file contains the common parts of your `mpd.conf` that are always present, regardless of the active configuration/"mode".
2.  **`config-*.mpd.conf.part`**: These files define the specific configurations for each "mode".
    -   The `*` in the filename will be used as the unique `key` for the mode (e.g., `config-exclusive.mpd.conf.part` for the `exclusive` mode).
    -   **Friendly Name**: The mode's friendly name displayed in the UI can be configured in each `config-*.mpd.conf.part` file by setting a special comment as the first line:
        ```
        # ConfigPartName: Exclusive (DSD)
        ```
        If this comment is missing, the tool will fall back to a title-cased version of the `key`.
3.  **`mpd.conf`**: The tool will write the combined content of `base.mpd.conf.part` and the selected `config-*.mpd.conf.part` to this file. It will also add a comment `# CurrentConfig: <modeKey>` as the first line to track the currently active mode.

**Example Directory Structure (`--config-dir`):**

```
/home/user/.config/mpd/
├── base.mpd.conf.part
├── config-exclusive.mpd.conf.part
├── config-pipewire.mpd.conf.part
└── mpd.conf (what the `mpd-config-switcher` will output)
```

You can see parts of my own `mpd` configuration in the [ExampleConfiguration](./ExampleConfiguration)

## Usage

Run the `mpd-config-switcher` binary with the desired CLI flags:

```bash
./mpd-config-switcher --port 56737 --config-dir ~/.config/mpd --systemd-unit-name mpd.service --sudo=false --expose=true
```

**Available CLI Flags:**

All flags are optional:

-   `--port <int>`: Port for the backend server (default: `56737`).
-   `--fe-port <int>`: **DEV ONLY**: Port for the frontend dev server (enables reverse proxy).
-   `--config-dir <path>`: Directory containing `mpd.conf` parts (default: `~/.config/mpd`, supports `~` for home directory).
-   `--systemd-unit-name <string>`: The systemd unit to restart (default: `mpd.service`).
-   `--sudo`: Restart `mpd` using `sudo systemctl` instead of `systemctl --user` (default: `false`).
-   `--expose`: Listen to all interfaces, so you can use this app from the comfort of other devices in your LAN/WiFi. (default: `false`)
-   `--open`: When true, launch the main browser to the app URL. (default: `false`)

## Disclaimer

This project started for my personal use, as I find it extremely cumbersome to switch between mpd’s "Exclusive mode" that can play DSD natively and "Shared mode" that coexist naturally with other Pipewire (or Pulse Audio) programs. 
- If you find this project useful, you can give the project a star.
- You're welcomed to create issues should you find any.

## Development

**Stack:**
-   **Backend**: Go / Fiber
-   **Frontend**: React / Rsbuild

**Getting Started:**

1.  **Install Dependencies**:
    -   Go: Ensure Go is installed.
    -   Node.js & npm: Ensure Node.js and npm are installed.
    -   Frontend: `npm install` in the project root.

2.  **Development Workflow**:
    -   **Frontend Development**: In one terminal, run `npm run dev` (or `npm start`). This will start the Rsbuild development server, usually on port `3000`.
    -   **Backend Development**: In another terminal, run the Go backend, pointing it to the frontend dev server:
        ```bash
        go run . --fe-port=3000
        ```
        This will proxy API requests to the Go backend and all other requests to the React dev server.


