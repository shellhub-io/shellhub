import { Module } from "vuex";
import axios from "axios";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { State } from "..";
import { IConnectToTerminal } from "@/interfaces/ITerminal";
import { IParams } from "@/interfaces/IParams";
import { router } from "@/router";

const webTermDimensions = {
  cols: 180,
  rows: 50,
};

const encodeURLParams = (params: IParams): string => Object.entries(params)
  .map(([key, value]) => `${key}=${encodeURIComponent(String(value))}`)
  .join("&");

enum MessageKind {
  Input = 1,
  Resize,
}

interface Message {
  kind: MessageKind;
  data: unknown;
}

export interface TerminalState {
  terminals: Record<string, { xterm: Terminal, websocket: WebSocket }>;
}

const createXtermInstance = (): Terminal => {
  const xterm = new Terminal({
    cursorBlink: true,
    fontFamily: "monospace",
    theme: { background: "#fff0000" },
    cols: webTermDimensions.cols,
    rows: webTermDimensions.rows,
  });

  const fitAddon = new FitAddon();
  xterm.loadAddon(fitAddon);

  if (xterm.element) {
    xterm.reset();
  }

  fitAddon.fit();
  xterm.focus();

  return xterm;
};

const createWebSocketConnection = (token: string, xterm: Terminal): WebSocket => {
  const protocol = window.location.protocol === "http:" ? "ws" : "wss";
  const wsInfo = { token, ...webTermDimensions };
  const url = `${protocol}://${window.location.host}/ws/ssh?${encodeURLParams(wsInfo)}`;

  const ws = new WebSocket(url);
  const enc = new TextEncoder();

  ws.onmessage = (ev) => xterm.write(ev.data);

  xterm.onData((data) => {
    const message: Message = {
      kind: MessageKind.Input,
      data: [...enc.encode(data)],
    };
    ws.send(JSON.stringify(message));
  });

  xterm.onResize((data) => {
    const message: Message = {
      kind: MessageKind.Resize,
      data: { cols: data.cols, rows: data.rows },
    };
    ws.send(JSON.stringify(message));
  });

  ws.onclose = () => xterm.write("\r\nConnection ended");

  return ws;
};

export const terminals: Module<TerminalState, State> = {
  namespaced: true,
  state: {
    terminals: {},
  },

  getters: {
    getTerminal: (state) => state.terminals,
  },

  mutations: {
    setNewTab(state, { token, xterm, websocket }) {
      state.terminals[token] = { xterm, websocket };
    },
    resizeTerminal(state, { token, cols, rows }) {
      const terminal = state.terminals[token];
      if (terminal) {
        terminal.xterm.resize(cols, rows);
        // Optionally trigger fitAddon.fit() if FitAddon is used
      }
    },
  },

  actions: {
    async fetch({ commit }, params: IConnectToTerminal) {
      try {
        const response = await axios.post("/ws/ssh", params);

        const { token } = response.data;
        const xterm = createXtermInstance();
        const websocket = createWebSocketConnection(token, xterm);

        commit("setNewTab", { token, xterm, websocket });
        await router.push({ name: "Connection", params: { token } });
      } catch (error) {
        commit("clearListPublicKeys");
        throw error;
      }
    },
  },
};
