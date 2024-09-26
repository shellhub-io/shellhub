import { Module } from "vuex";
import axios from "axios";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import { WebLinksAddon } from '@xterm/addon-web-links';
import { WebglAddon } from "@xterm/addon-webgl";
import { ImageAddon } from 'xterm-addon-image';
import { State } from "..";
import { IConnectToTerminal } from "@/interfaces/ITerminal";
import { IParams } from "@/interfaces/IParams";
import { router } from "@/router";

const webTermDimensions = {
  cols: 0,
  rows: 0,
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
  terminals: Record<string, { xterm: Terminal, websocket: WebSocket, fitAddon: FitAddon, uid: string }>;
  themes: Array<{ name: string, file: string, dark: boolean }>;
}

const createXtermInstance = (theme: unknown = {}): { xterm: Terminal, fitAddon: FitAddon } => {
  const xterm = new Terminal({
    cursorBlink: true,
    fontFamily: "monospace",
    fontSize: 14,
    theme: theme || { background: "#fff0000" },
  });

  const fitAddon = new FitAddon();
  xterm.loadAddon(fitAddon);
  xterm.loadAddon(new WebglAddon());
  xterm.loadAddon(new WebLinksAddon());
  xterm.loadAddon(new ImageAddon());

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

  ws.onmessage = (ev) => {
    xterm.write(ev.data);
  };

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
    themes: [],
  },

  getters: {
    getTerminal: (state) => state.terminals,
    getThemes: (state) => state.themes,
    getFontSize: (state) => state.terminals.fontSize,
    findThemeByName: (state) => (themeName: string) => state.themes.find((theme) => theme.name === themeName),
  },

  mutations: {
    setNewTab(state, { token, xterm, websocket, fitAddon, uid }) {
      state.terminals[token] = { xterm, websocket, fitAddon, uid };
    },
    removeTerminal(state, token) {
      if (state.terminals[token]) {
        state.terminals[token].xterm.dispose();
        state.terminals[token].websocket.close();
        delete state.terminals[token];
      }
    },
    setThemes(state, themes) {
      state.themes = themes;
    },
    applyTheme(state, { token, theme }) {
      const terminal = state.terminals[token];
      terminal.xterm.options.theme = theme;
    },
    setFontSize(state, { token, fontSize }) {
      if (state.terminals[token]) {
        state.terminals[token].xterm.options.fontSize = fontSize;
        state.terminals[token].fitAddon.fit();
      }
    },
  },

  actions: {
    async fetchThemes({ commit }) {
      try {
        const response = await axios.get("/xtermthemes/metadata.json");
        commit("setThemes", response.data);
      } catch (error) {
        console.error("Error fetching themes:", error);
      }
    },
    async fetch({ commit }, params: IConnectToTerminal) {
      try {
        const response = await axios.post("/ws/ssh", params);
        const uid = params.device;
        const { token } = response.data;
        const { xterm, fitAddon } = createXtermInstance();
        const websocket = createWebSocketConnection(token, xterm);

        websocket.addEventListener("open", () => {
          const data = fitAddon.proposeDimensions();
          const message: Message = {
            kind: MessageKind.Resize,
            data: { cols: data?.cols, rows: data?.rows },
          };
          websocket.send(JSON.stringify(message));
        });

        commit("setNewTab", { token, xterm, websocket, fitAddon, uid });
        await router.push({ name: "Connection", params: { token } });
      } catch (error) {
        commit("clearListPublicKeys");
        throw error;
      }
    },
    removeTerminal({ commit }, token) {
      commit("removeTerminal", token);
    },
    async applyTheme({ commit }, { token, themeName }) {
      await axios.get(`/xtermthemes/${themeName}`).then((response) => {
        commit("applyTheme", { token, theme: response.data });
      });
    },
    changeFontSize(context, { token, adjustment }) {
      const currentFontSize = context.state.terminals[token].xterm.options.fontSize || 14;
      let newFontSize = currentFontSize + adjustment;
      if (newFontSize < 12) {
        newFontSize = 12;
      }
      context.commit("setFontSize", { token, fontSize: newFontSize });
    },

  },
};