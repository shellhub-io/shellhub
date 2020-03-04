<template>
  <div>
  </div>
</template>

<script>
import { Terminal } from 'xterm';
import * as fit from 'xterm/lib/addons/fit/fit';
import * as attach from 'xterm/lib/addons/attach/attach';
import 'xterm/dist/xterm.css';

Terminal.applyAddon(fit);
Terminal.applyAddon(attach);

export default {
  name: 'Terminal',

  props: ['uid', 'username', 'password'],

  mounted() {
    this.xterm = new Terminal({
      cursorBlink: true,
      fontFamily: 'monospace'
    });
  },

  beforeDestroy() {
    //this.xterm.dispose();
  },

  watch: {
    isOpen: value => {
      if (!value) {
        this.xterm.dispose();
      }
    }
  },

  methods: {
    open() {
      setTimeout(() => {
        this.connect();
      }, 1000);
    },

    close() {
      this.toggleTerminal('');
    },

    connect() {
      this.username = '';
      this.passwd = '';
      this.device = this.$props.uid;

      setTimeout(() => {
        this.xterm.fit();
      }, 2000);

      this.xterm.open(this.$refs.terminal);
      this.xterm.fit();
      this.xterm.focus();

      const params = Object.entries({
        user: `${this.username}@${this.device}`,
        passwd: this.passwd,
        cols: this.xterm.cols,
        rows: this.xterm.rows
      })
        .map(([k, v]) => {
          return `${k}=${v}`;
        })
        .join('&');

      var ws = new WebSocket(`ws://${location.host}/ws/ssh?${params}`);

      ws.onopen = () => {
        this.xterm.attach(ws, true, true);
      };

      ws.onclose = () => {
        this.xterm.detach(ws);
      };
    }
  }
};
</script>
