<template>
  <div class="ma-0 pa-0 w-100 fill-height position-relative">
    <div ref="terminal" class="terminal" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from "vue";
import { useEventListener } from "@vueuse/core";
import "xterm/css/xterm.css";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { IParams } from "@/interfaces/IParams";

enum MessageKind {
  Input = 1,
  Resize,
}

interface Message {
  kind: MessageKind;
  data: unknown;
}

interface WebTermDimensions {
  cols: number;
  rows: number;
}

const { token } = defineProps<{
  token: string;
}>();

const terminal = ref<HTMLElement>({} as HTMLElement);
const xterm = ref<Terminal>({} as Terminal);
const fitAddon = ref<FitAddon>(new FitAddon());
const ws = ref<WebSocket>({} as WebSocket);
const textEncoder = new TextEncoder();

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

const getWebTermDimensions = (): WebTermDimensions => ({
  cols: xterm.value.cols,
  rows: xterm.value.rows,
});

const encodeURLParams = (params: IParams): string => Object.entries(params).map(([key, value]) => `${key}=${value}`).join("&");

const getWebSocketUrl = (dimensions: WebTermDimensions): string => {
  const protocol = window.location.protocol === "http:" ? "ws" : "wss";
  const wsInfo = { token, ...dimensions };

  return `${protocol}://${window.location.host}/ws/ssh?${encodeURLParams(wsInfo)}`;
};

const setupTerminalEvents = () => {
  xterm.value.onData((data) => {
    const message: Message = {
      kind: MessageKind.Input,
      data: [...textEncoder.encode(data)],
    };
    ws.value.send(JSON.stringify(message));
  });

  xterm.value.onResize((data) => {
    const message: Message = {
      kind: MessageKind.Resize,
      data: { cols: data.cols, rows: data.rows },
    };
    ws.value.send(JSON.stringify(message));
  });
};

const setupWebSocketEvents = () => {
  ws.value.onopen = () => {
    fitAddon.value.fit();
  };

  ws.value.onmessage = (event) => {
    xterm.value.write(event.data);
  };

  ws.value.onclose = () => {
    xterm.value.write("\r\nConnection ended");
  };
};

const initializeWebSocket = () => {
  const dimensions = getWebTermDimensions();
  const wsUrl = getWebSocketUrl(dimensions);
  ws.value = new WebSocket(wsUrl);
  setupWebSocketEvents();
};

onMounted(() => {
  initializeTerminal();
  xterm.value.open(terminal.value);
  xterm.value.focus();

  setupTerminalEvents();
  initializeWebSocket();
});

useEventListener(window, "resize", () => {
  fitAddon.value.fit();
});

onUnmounted(() => {
  if (ws.value.OPEN) ws.value.close();
  xterm.value.clear();
});
</script>

<style scoped lang="scss">
.terminal {
  position: absolute;
  inset: 0;
  margin-right: 0px;
}
</style>
