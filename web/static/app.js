class YouTubeRemote {
  constructor() {
    this.ws = null;
    this.connected = false;
    this.initElements();
    this.initEventListeners();
    this.autoConnect();
  }

  initElements() {
    this.urlInput = document.getElementById("url-input");
    this.openUrlBtn = document.getElementById("open-url");
    this.prevBtn = document.getElementById("prev");
    this.playBtn = document.getElementById("play");
    this.nextBtn = document.getElementById("next");
    this.volumeDownBtn = document.getElementById("volume-down");
    this.volumeUpBtn = document.getElementById("volume-up");
    this.volumeSlider = document.getElementById("volume-slider");
    this.statusElement = document.getElementById("status");
  }

  initEventListeners() {
    this.openUrlBtn.addEventListener("click", () => this.openUrl());
    this.prevBtn.addEventListener("click", () => this.sendCommand("previous"));
    this.playBtn.addEventListener("click", () => this.sendCommand("play"));
    this.nextBtn.addEventListener("click", () => this.sendCommand("next"));
    this.volumeDownBtn.addEventListener("click", () =>
      this.sendCommand("volumeDown"),
    );
    this.volumeUpBtn.addEventListener("click", () =>
      this.sendCommand("volumeUp"),
    );
    this.volumeSlider.addEventListener("change", () => this.setVolume());

    // Handle Enter key in URL input
    this.urlInput.addEventListener("keypress", (e) => {
      if (e.key === "Enter") {
        this.openUrl();
      }
    });

    // Auto-reconnect on window focus
    window.addEventListener("focus", () => {
      if (!this.connected) {
        this.autoConnect();
      }
    });
  }

  autoConnect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    console.log("Connecting to WebSocket server:", wsUrl);

    try {
      this.ws = new WebSocket(wsUrl);
    } catch (error) {
      console.error("Failed to connect:", error);
      this.updateStatus("Failed to connect to server", false);
      return;
    }

    this.ws.onopen = () => {
      this.connected = true;
      this.updateStatus("Connected", true);
      this.enableControls(true);
    };

    this.ws.onclose = () => {
      this.connected = false;
      this.updateStatus("Disconnected", false);
      this.enableControls(false);
      // Try to reconnect after 3 seconds
      setTimeout(() => this.autoConnect(), 3000);
    };

    this.ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      this.updateStatus("Connection error", false);
    };

    this.ws.onmessage = (event) => {
      try {
        const response = JSON.parse(event.data);
        if (response.error) {
          this.updateStatus(`Error: ${response.error}`, false);
        }
      } catch (e) {
        console.error("Failed to parse message:", e);
      }
    };
  }

  sendCommand(action, value = "") {
    if (!this.connected) {
      this.updateStatus("Not connected", false);
      return;
    }

    const command = {
      action: action,
      value: value,
    };

    try {
      this.ws.send(JSON.stringify(command));
    } catch (e) {
      console.error("Failed to send command:", e);
      this.updateStatus("Failed to send command", false);
    }
  }

  openUrl() {
    const url = this.urlInput.value.trim();
    if (!url) {
      this.updateStatus("Please enter a YouTube URL", false);
      return;
    }

    this.sendCommand("open", url);
    this.urlInput.value = "";
  }

  setVolume() {
    const volume = this.volumeSlider.value;
    this.sendCommand("volume", volume);
  }

  updateStatus(message, isConnected) {
    this.statusElement.textContent = message;
    this.statusElement.className = `status ${
      isConnected ? "connected" : "disconnected"
    }`;
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
    }
  }

  enableControls(enabled) {
    const buttons = [
      this.openUrlBtn,
      this.prevBtn,
      this.playBtn,
      this.nextBtn,
      this.volumeDownBtn,
      this.volumeUpBtn,
    ];

    buttons.forEach((button) => {
      button.disabled = !enabled;
    });
    this.volumeSlider.disabled = !enabled;
    this.urlInput.disabled = !enabled;
  }
}

// Initialize the remote control when the page loads
window.addEventListener("load", () => {
  new YouTubeRemote();
});
