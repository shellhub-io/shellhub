<template>
<v-dialog v-model="show">
    <template v-slot:activator="{ on }">
        <v-btn color="primary" dark @click="open()" icon text><font-awesome-icon icon="terminal">terminal</font-awesome-icon></v-btn>
    </template>
    <v-card>
        <v-toolbar dark color="primary" v-if="visible">
            <v-btn icon dark @click="close()">
                <v-icon>close</v-icon>
            </v-btn>
            <v-toolbar-title>Terminal</v-toolbar-title>
            <v-spacer></v-spacer>
        </v-toolbar>
        <div ref="terminal"></div>
    </v-card>
</v-dialog>
</template>

<script>
import { Terminal } from "xterm";
import * as fit from "xterm/lib/addons/fit/fit";
import * as attach from "xterm/lib/addons/attach/attach";
import "xterm/dist/xterm.css";

Terminal.applyAddon(fit);
Terminal.applyAddon(attach);

export default {
  name: "TerminalDialog",

  props: ["uid"],

  data() {
    return {
      username: "",
      passwd: "",
      visible: false
    };
  },

  mounted() {
    this.xterm = new Terminal({
      cursorBlink: true,
      fontFamily: "monospace"
    });
  },

  watch: {
    show(value) {
      if (!value) {
        this.ws.close();
      }
    }
  },

  computed: {
    show: {
      get() {
        return this.$store.getters["modals/terminal"] === this.$props.uid;
      },

      set(value) {
        if (value) {
          this.$store.dispatch("modals/toggleTerminal", this.$props.uid);
        } else {
          this.$store.dispatch("modals/toggleTerminal", "");
        }
      }
    }
  },

  methods: {
    open() {
      this.username = prompt("Username:");
      this.passwd = prompt("Password:");

      this.$store.dispatch("modals/toggleTerminal", this.$props.uid);

      if (this.xterm.element) {
        this.xterm.reset();
      }

      setTimeout(() => {
        this.visible = true;
        this.connect();
      }, 1000);
    },

    close() {
      this.$store.dispatch("modals/toggleTerminal", "");
    },

    connect() {
      this.$nextTick(() => this.xterm.fit());

      if (!this.xterm.element) {
        this.xterm.open(this.$refs.terminal);
      }

      this.xterm.fit();
      this.xterm.focus();

      const params = Object.entries({
        user: `${this.username}@${this.$props.uid}`,
        passwd: this.passwd,
        cols: this.xterm.cols,
        rows: this.xterm.rows
      })
        .map(([k, v]) => {
          return `${k}=${v}`;
        })
        .join("&");

      this.ws = new WebSocket(`ws://${location.host}/ws/ssh?${params}`);

      this.ws.onopen = () => {
        this.xterm.attach(this.ws, true, true);
      };

      this.ws.onclose = () => {
        this.xterm.detach(this.ws);
      };
    }
  }
};
</script>
