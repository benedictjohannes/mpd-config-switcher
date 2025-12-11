# A Config Switcher and Restarter for `mpd`

 `mpd`' is configurable via its configuration (commonly `mpd.conf`). A change in the configuration file would require a restart of `mpd`. As such, switching between `mpd` output of "ALSA" exclusive mode that play DSD natively and "Shared mode" that coexist naturally with other Pipewire (or Pulse Audio) programs, is quite cumbersome. This tool aims to dynamically switch `mpd` configuration "parts" via an easy web GUI, with
-   **Dynamic Config Discovery**: The backend scans a configurable directory for `mpd` configuration "parts" (`{mpdConfigDir}/config-*.mpd.conf.part`).
-   **Single binary**: Package as a go app with a web GUI. You can switch your `mpd` configuration with your phone!

## Installation

The easiest way to get the `mpd-config-switcher` is to download a pre-built binary from the [GitHub Releases](https://github.com/benedictjohannes/mpd-config-switcher/releases) page. Binaries are available for Linux, supporting both `amd64` and `arm64` architectures.

1.  **Download the binary**: Choose the appropriate binary for your operating system and architecture.
    *   `mpd-config-switcher-linux-amd64` (for 64-bit Linux)
    *   `mpd-config-switcher-linux-arm64` (for ARM64 Linux)
    *   You can see the [section to automatically download the latest version](#1-download-and-configure-the-binary).
2.  **Make it executable**: `chmod +x mpd-config-switcher-<os>-<arch>` (replace `<os>-<arch>` with the actual name).
3.  **Place it**: Move the binary to a directory in your PATH (e.g., `/usr/local/bin`).

Alternatively, you can [build from source.](#building-from-source)

> Tips: You can see the [Running it automatically](#running-it-automatically) section to configure it to run at boot. That way, this is always available for your usage.

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
3.  **`mpd.conf`**: The tool will concatenate the content of `base.mpd.conf.part` and the selected `config-*.mpd.conf.part` to this file. It will also add a comment `# CurrentConfig: <modeKey>` as the first line to track the currently active mode.

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

## Running it automatically

These are the easiest steps to have it run automatically on boot:

### 1. Download and configure the binary 

You can install with a single line in your terminal:

```bash
curl -sL https://github.com/benedictjohannes/mpd-config-switcher/releases/latest/download/mpd-config-switcher-$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/') -o mpd-config-switcher && chmod +x mpd-config-switcher && sudo mv mpd-config-switcher /usr/local/bin/
```

### 2. Confirm that the binary is runnable from your system path:
```bash
mpd-config-switcher --help
```
### 3. Create a systemd unit

You have two options:

#### As a user level unit 

This option is recommended if your `mpd` is running as a user level service and you use Linux Desktop.

1. Create the unit file
```bash
mkdir -p ~/.config/systemd/user
nano ~/.config/systemd/user/mpd-config-switcher.service # ofc you can use vim too
```
1. Insert this to the file
```ini
[Unit]
Description=MPD Config Switcher User Service
# Wait for network. If you run MPD as a user service, you can append 'mpd.service'
# Example: After=network.target mpd.service
After=network.target

[Service]
# Path to the executable moved in the previous step, and configure the running parameter
ExecStart=/usr/local/bin/mpd-config-switcher --port 56737 --expose
# Standard directives
Restart=always
RestartSec=5s

[Install]
WantedBy=default.target
```
1. Enable and start the unit
```bash
systemctl --user daemon-reload
systemctl --user enable --now mpd-config-switcher.service
# optional: check the service status and logs
systemctl --user status mpd-config-switcher.service
```

#### As system level unit

1. Create the unit file
```bash
sudo nano /etc/systemd/system/mpd-config-switcher.service # ofc you can use vim too
```
1. Insert this to the file
```ini
[Unit]
Description=MPD Config Switcher Service
# assuming you have mpd unit
After=network.target mpd.service

[Service]
# Set User/Group if needed, otherwise run as root (default for system units)
# User=youruser
# Group=youruser

# Path to the executable moved in the previous step, and configure the running parameter
ExecStart=/usr/local/bin/mpd-config-switcher --port 56737 --expose

# Standard directives
Restart=always
RestartSec=5s

[Install]
WantedBy=multi-user.target
```
3. Enable and start the unit
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now mpd-config-switcher.service
# optional: check the service status and logs
systemctl status mpd-config-switcher.service
```

## Disclaimer

This project started for my personal use, as I find it extremely cumbersome to switch between mpd’s "Exclusive mode" that can play DSD natively and "Shared mode" that coexist naturally with other Pipewire (or Pulse Audio) programs. 
- If you find this project useful, you can give the project a star.
- You're welcomed to create issues should you find any.

## Build from Source

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
        The `--fe-port=3000` flag tells the Go backend to reverse proxy all non-API requests to the React development server, enabling a seamless development experience.


