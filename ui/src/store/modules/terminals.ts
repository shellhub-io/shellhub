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
  terminals: Record<string, { xterm: Terminal, websocket: WebSocket, fitAddon: FitAddon }>;
}

const createXtermInstance = (): { xterm: Terminal, fitAddon: FitAddon } => {
  const xterm = new Terminal({
    cursorBlink: true,
    fontFamily: "monospace",
    theme: { background: "#fff0000" },
    cols: 180,
    rows: 20,
  });

  const fitAddon = new FitAddon();
  xterm.loadAddon(fitAddon);

  if (xterm.element) {
    xterm.reset();
  }

  return { xterm, fitAddon };
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

    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message));
    }
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
    setNewTab(state, { token, xterm, websocket, fitAddon }) {
      state.terminals[token] = { xterm, websocket, fitAddon };
    },
  },

  actions: {
    async fetch({ commit }, params: IConnectToTerminal) {
      try {
        const response = await axios.post("/ws/ssh", params);

        const { token } = response.data;
        const { xterm, fitAddon } = createXtermInstance();
        const websocket = createWebSocketConnection(token, xterm);

        commit("setNewTab", { token, xterm, websocket, fitAddon });
        await router.push({ name: "Connection", params: { token } });
      } catch (error) {
        commit("clearListPublicKeys");
        throw error;
      }
    },
  },
};
