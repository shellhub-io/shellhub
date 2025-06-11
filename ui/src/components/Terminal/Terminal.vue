<template>
  <!-- Container that fills the available height and hosts the terminal -->
  <div class="ma-0 pa-0 w-100 fill-height position-relative">
    <!-- The xterm.js terminal will be mounted here -->
    <div ref="terminal" class="terminal" data-test="terminal-container" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { useEventListener } from "@vueuse/core";

// Terminal styles and required classes
import "xterm/css/xterm.css";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";

// Type definitions
import { IParams } from "@/interfaces/IParams";
import {
  InputMessage,
  MessageKind,
  ResizeMessage,
  SignatureMessage,
  WebTermDimensions,
} from "@/interfaces/ITerminal";

// SSH signing utilities
import {
  parsePrivateKeySsh,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
} from "@/utils/validate";
import handleError from "@/utils/handleError";

// Props passed to the component
const { token, privateKey } = defineProps<{
  token: string; // JWT token for WebSocket authentication
  privateKey?: string | null; // Optional SSH private key for challenge-response auth
}>();

// Refs and runtime state
const terminal = ref<HTMLElement>({} as HTMLElement); // Terminal DOM container
const xterm = ref<Terminal>({} as Terminal); // xterm.js terminal instance
const fitAddon = ref<FitAddon>(new FitAddon()); // Auto-fit terminal to container
const ws = ref<WebSocket>({} as WebSocket); // Active WebSocket connection
const textEncoder = new TextEncoder(); // Converts strings to Uint8Array
const isReady = ref(false); // Tracks if WS is open and ready

// Initialize terminal instance and attach the fit addon
const initializeTerminal = () => {
  xterm.value = new Terminal({
    cursorBlink: true,
    fontFamily: "monospace",
    theme: {
      background: "#0f1526",
    },
  });

  xterm.value.loadAddon(fitAddon.value);
};

// Get terminal dimensions for WebSocket URL
const getWebTermDimensions = (): WebTermDimensions => ({
  cols: xterm.value.cols,
  rows: xterm.value.rows,
});

// Convert token and dimensions into query params for WS URL
const encodeURLParams = (params: IParams): string => Object.entries(params).map(([key, value]) => `${key}=${value}`).join("&");

// Check if WebSocket is open and usable
const isWebSocketOpen = () => isReady.value && ws.value.readyState === WebSocket.OPEN;

// Construct the WebSocket URL with protocol, host, and query
const getWebSocketUrl = (dimensions: WebTermDimensions): string => {
  const protocol = window.location.protocol === "http:" ? "ws" : "wss";
  const wsInfo = { token, ...dimensions };

  return `${protocol}://${window.location.host}/ws/ssh?${encodeURLParams(wsInfo)}`;
};

// Set up terminal events for user input and resize events
const setupTerminalEvents = () => {
  // Send user input over WebSocket
  xterm.value.onData((data) => {
    if (!isWebSocketOpen()) return;

    const message: InputMessage = {
      kind: MessageKind.Input,
      data: [...textEncoder.encode(data)],
    };

    ws.value.send(JSON.stringify(message));
  });

  // Send terminal resize info over WebSocket
  xterm.value.onResize(({ cols, rows }) => {
    if (!isWebSocketOpen()) return;

    const message: ResizeMessage = {
      kind: MessageKind.Resize,
      data: { cols, rows },
    };

    ws.value.send(JSON.stringify(message));
  });
};

// Write text output to the terminal UI
const writeToTerminal = (data: string) => {
  xterm.value.write(data);
};

// Handles signing of SSH challenge using the user's private key
const signWebSocketChallenge = async (
  key: string,
  base64Challenge: string,
): Promise<string> => {
  const challenge = atob(base64Challenge);
  const parsedKey = parsePrivateKeySsh(key);

  if (parsedKey.type === "ed25519") {
    return createSignerPrivateKey(parsedKey, challenge);
  }

  return decodeURIComponent(await createSignatureOfPrivateKey(parsedKey, challenge));
};

// Initialize WebSocket and its message handling
const setupWebSocketEvents = () => {
  ws.value.onopen = () => {
    fitAddon.value.fit(); // Adjust terminal to container
    isReady.value = true;

    // If using public key auth, expect challenge message first
    if (privateKey) {
      ws.value.onmessage = async (event) => {
        try {
          const parsed = JSON.parse(event.data) as SignatureMessage;
          if (parsed.kind === MessageKind.Signature) {
            const signature = await signWebSocketChallenge(privateKey, parsed.data);
            ws.value.send(JSON.stringify({ kind: 3, data: signature }));

            // After challenge is signed, switch to raw message handling
            ws.value.onmessage = (e) => writeToTerminal(e.data);
          }
        } catch (err) {
          writeToTerminal("\r\nFailed to sign challenge.\r\n");
          handleError(err);
          ws.value.close();
        }
      };

      return; // Skip password-mode handler
    }

    // For password-based logins, simply write messages to the terminal
    ws.value.onmessage = (event) => {
      writeToTerminal(event.data);
    };
  };

  ws.value.onclose = () => {
    writeToTerminal("\r\nConnection ended\r\n");
    isReady.value = false;
  };
};

// Connect and initialize WebSocket session
const initializeWebSocket = () => {
  const dimensions = getWebTermDimensions();
  const wsUrl = getWebSocketUrl(dimensions);
  ws.value = new WebSocket(wsUrl);
  setupWebSocketEvents();
};

// Mount lifecycle: Initialize terminal and WebSocket
onMounted(() => {
  initializeTerminal();
  xterm.value.open(terminal.value);
  xterm.value.focus();

  setupTerminalEvents();
  initializeWebSocket();
});

// Resize the terminal when window is resized
useEventListener(window, "resize", () => {
  fitAddon.value.fit();
});

// Cleanup lifecycle: close WebSocket if active
onUnmounted(() => {
  if (isWebSocketOpen()) ws.value.close();
});

// Optional expose for testing or parent communication
defineExpose({ xterm, ws });
</script>

<style scoped lang="scss">
.terminal {
  position: absolute;
  inset: 0; // Fills the container completely
}
</style>
