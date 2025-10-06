<template>
  <WindowDialog
    v-model="showDialog"
    :force-fullscreen="true"
    @close="close"
    title="Terminal"
    :description="`Connected to ${deviceName}`"
    icon="mdi-console"
    :show-close-button="true"
    :show-footer="false"
    :scrollable="false"
    class="bg-terminal h-100"
  >
    <template #titlebar-actions>
      <v-btn
        :icon="showThemeDrawer ? 'mdi-palette' : 'mdi-palette-outline'"
        variant="text"
        @click="showThemeDrawer = !showThemeDrawer"
        data-test="theme-toggle-btn"
      />
    </template>

    <div class="ma-0 pa-0 w-100 fill-height position-relative">
      <div ref="terminal" class="terminal" data-test="terminal-container" />

      <TerminalThemeDrawer
        v-model="selectedThemeName"
        v-model:showDrawer="showThemeDrawer"
        @update:selected-theme="applyTheme"
        @update:font-settings="applyFontSettings"
        data-test="theme-picker"
      />
    </div>
  </WindowDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useEventListener } from "@vueuse/core";
// Terminal styles and required classes
import "@xterm/xterm/css/xterm.css";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";

// Type definitions
import { IParams } from "@/interfaces/IParams";
import {
  ErrorMessage,
  InputMessage,
  ITerminalTheme,
  MessageKind,
  ResizeMessage,
  SignatureMessage,
  WebTermDimensions,
} from "@/interfaces/ITerminal";

// SSH signing utilities
import {
  generateSignature,
} from "@/utils/sshKeys";

import handleError from "@/utils/handleError";
import WindowDialog from "../WindowDialog.vue";
import TerminalThemeDrawer from "./TerminalThemeDrawer.vue";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

// Props passed to the component
const { token, privateKey, passphrase } = defineProps<{
  token: string;
  privateKey?: string | null;
  passphrase?: string;
  deviceName: string;
}>();

const emit = defineEmits<{
  close: [];
}>();

const terminalThemeStore = useTerminalThemeStore();
const showDialog = defineModel<boolean>({ required: true });
const terminal = ref<HTMLElement>({} as HTMLElement); // Terminal DOM container
const xterm = ref<Terminal>({} as Terminal); // xterm.js terminal instance
const fitAddon = ref<FitAddon>(new FitAddon()); // Auto-fit terminal to container
const ws = ref<WebSocket>({} as WebSocket); // Active WebSocket connection
const selectedThemeName = computed({
  get: () => terminalThemeStore.currentThemeName,
  set: (value: string) => terminalThemeStore.setTheme(value),
});
const showThemeDrawer = ref(false);

const applyTheme = (theme: ITerminalTheme) => {
  xterm.value.options.theme = theme.colors;
};

const applyFontSettings = (settings: { fontFamily: string; fontSize: number }) => {
  xterm.value.options.fontSize = settings.fontSize;
  xterm.value.options.fontFamily = settings.fontFamily;
  fitAddon.value.fit();
};

const initializeTerminal = () => {
  xterm.value = new Terminal({
    cursorBlink: true,
    fontFamily: terminalThemeStore.currentFontFamily,
    fontSize: terminalThemeStore.currentFontSize,
  });
  applyTheme(terminalThemeStore.currentTheme);
  xterm.value.loadAddon(fitAddon.value);
};

// Returns the current terminal dimensions for WebSocket session init.
const getTerminalDimensions = (): WebTermDimensions => ({
  cols: xterm.value.cols,
  rows: xterm.value.rows,
});

// Resize terminal on window resize events
const registerResizeHandler = () => {
  useEventListener(window, "resize", () => fitAddon.value.fit());
};

// Encodes a params object as URL query string.
const encodeURLParams = (params: IParams): string => new URLSearchParams(params as Record<string, string>).toString();

// Constructs the WebSocket URL for the SSH session.
const getWebSocketUrl = (dimensions: WebTermDimensions): string => {
  const protocol = window.location.protocol === "http:" ? "ws" : "wss";
  return `${protocol}://${window.location.host}/ws/ssh?${encodeURLParams({
    token,
    ...dimensions,
  })}`;
};

// Determines if the current WebSocket connection is open and usable.
const isWebSocketOpen = (): boolean => ws.value.readyState === WebSocket.OPEN;

// Binds terminal input and resize events to WebSocket messages.
const setupTerminalEvents = () => {
  // Send user input over WebSocket
  xterm.value.onData((data) => {
    if (!isWebSocketOpen()) return;

    const message: InputMessage = {
      kind: MessageKind.Input,
      data: data.slice(0, 4096), // Limit input to 4096 characters
    };
    ws.value.send(JSON.stringify(message));
  });

  // Send terminal resize info over WebSocket
  xterm.value.onResize(({ cols, rows }) => {
    if (!isWebSocketOpen()) return;

    const resizeMsg: ResizeMessage = {
      kind: MessageKind.Resize,
      data: { cols, rows },
    };
    ws.value.send(JSON.stringify(resizeMsg));
  });
};

// Handles signing a challenge received from the backend.
const signWebSocketChallenge = async (
  key: string,
  base64Challenge: Base64URLString,
) => {
  const challengeBuffer = Buffer.from(base64Challenge, "base64");
  return generateSignature(key, challengeBuffer, passphrase);
};

// Parses and handles JSON-structured WebSocket messages (e.g., challenge-response).
type IncomingMessage = SignatureMessage | ErrorMessage;

const handleJsonMessage = async (message: string): Promise<void> => {
  try {
    const parsed: IncomingMessage = JSON.parse(message);

    switch (parsed.kind) {
      case MessageKind.Error: {
        xterm.value.write(parsed.data);
        break;
      }
      // If using public key auth, expect challenge message first
      case MessageKind.Signature: {
        if (!privateKey) return;

        const signature = await signWebSocketChallenge(privateKey, parsed.data);
        ws.value.send(
          JSON.stringify({ kind: MessageKind.Signature, data: signature }),
        );

        // Register resize handler
        registerResizeHandler();

        break;
      }

      default:
        break;
    }
  } catch (error) {
    handleError(error);
  }
};

// Handles WebSocket messages, delegating binary vs. JSON text messages.
const handleWebSocketMessage = async (rawData: Blob | string): Promise<void> => {
  if (rawData instanceof Blob) {
    // For password-based logins, always just write messages to the terminal
    xterm.value.write(await rawData.text());
    registerResizeHandler();
  } else {
    await handleJsonMessage(rawData);
  }
};

// Sets up WebSocket event handlers: message and close.
const setupWebSocketEvents = () => {
  ws.value.onmessage = async (event) => { await handleWebSocketMessage(event.data); };
  ws.value.onclose = () => {
    xterm.value.write("\r\nConnection ended\r\n");
  };
};

const close = () => {
  showDialog.value = false;
  emit("close");
};

// Initializes the WebSocket session with terminal dimensions.
const initializeWebSocket = () => {
  fitAddon.value.fit(); // Ensure terminal is sized correctly before connecting
  const dimensions = getTerminalDimensions();
  ws.value = new WebSocket(getWebSocketUrl(dimensions));
  setupWebSocketEvents();
};

let lastEscPress = 0;
const handleEscKey = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    const currentTime = new Date().getTime();
    if (currentTime - lastEscPress < 400) {
      close();
    }
    lastEscPress = currentTime;
  }
};

useEventListener("keyup", handleEscKey);

// Lifecycle: Setup terminal and WebSocket calls
onMounted(async () => {
  await terminalThemeStore.loadInitialFont();
  await terminalThemeStore.loadThemes();
  initializeTerminal();
  setupTerminalEvents();
  xterm.value.open(terminal.value);
  initializeWebSocket();
  xterm.value.focus();
});

// Cleanup lifecycle: close WebSocket if active on unMount
onUnmounted(() => {
  if (isWebSocketOpen()) ws.value.close();
});

// Optional expose for testing or parent communication
defineExpose({ xterm, ws, showDialog });
</script>

<style scoped lang="scss">
.terminal {
  position: absolute;
  inset: 0; // Fills the container completely
}
</style>
