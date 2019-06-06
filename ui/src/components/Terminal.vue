<template>
  <div>
    User: <input v-model="username">
    Password: <input v-model="passwd">
    Device ID: <input v-model="device">
    <button @click=open()>Connect</button>
    <div class="terminal" ref="terminal"></div>
  </div>
</template>

<script>
import { Terminal } from "xterm";
import * as fit from "xterm/lib/addons/fit/fit";
import * as attach from "xterm/lib/addons/attach/attach";
import "xterm/dist/xterm.css";

Terminal.applyAddon(fit);
Terminal.applyAddon(attach);

export default {
  name: "Terminal",
  data() {
    return {
      username: "",
      passwd: "",
      device: ""
    };
  },

  methods: {
    open() {
      const xterm = new Terminal({
        cursorBlink: true,
        fontFamily: "monospace"
      });

      xterm.open(this.$refs.terminal);
      xterm.focus();
      xterm.fit();

      const params = Object.entries({
        user: `${this.username}@${this.device}`,
        passwd: this.passwd,
        cols: xterm.cols,
        rows: xterm.rows
      })
        .map(([k, v]) => {
          return `${k}=${v}`;
        })
        .join("&");

      var ws = new WebSocket(`ws://${location.host}/ws/ssh?${params}`);

      ws.onopen = () => {
        xterm.attach(ws, true, true);
      };

      ws.onclose = () => {
        xterm.detach(ws);
      }
    }
  }
};
</script>
