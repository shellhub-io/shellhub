<template>
  <fragment>
    <v-tooltip bottom>
      <template v-slot:activator="{ on }">
        <v-icon
          v-on="on"
          @click="open()"
        >
          mdi-console
        </v-icon>
      </template>
      <span>Terminal</span>
    </v-tooltip>
    <v-dialog
      v-model="show"
      max-width="1024px"
    >
      <v-card>
        <v-toolbar
          dark
          color="primary"
        >
          <v-btn
            icon
            dark
            @click="close()"
          >
            <v-icon>close</v-icon>
          </v-btn>
          <v-toolbar-title>Terminal</v-toolbar-title>

          <v-spacer />
        </v-toolbar>

        <v-card
          v-if="showLoginForm"
          class="ma-0 pa-6"
          outlined
        >
          <v-form
            ref="form"
            v-model="valid"
            lazy-validation
            @submit.prevent="connect()"
          >
            <v-text-field
              ref="username"
              v-model="username"
              label="Username"
              autofocus
              :rules="[rules.required]"
              :validate-on-blur="true"
            />
            <v-text-field
              v-model="passwd"
              label="Password"
              type="password"
              :rules="[rules.required]"
              :validate-on-blur="true"
            />
            <v-card-actions>
              <v-spacer />
              <v-btn
                type="submit"
                color="primary"
                class="mt-4"
              >
                Connect
              </v-btn>
            </v-card-actions>
          </v-form>
        </v-card>

        <div ref="terminal" />
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { FitAddon } from 'xterm-addon-fit';

import 'xterm/css/xterm.css';

export default {
  name: 'TerminalDialog',

  props: {
    uid: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      username: '',
      passwd: '',
      showLoginForm: true,
      valid: true,
      rules: {
        required: (value) => !!value || 'Required',
      },
    };
  },

  computed: {
    show: {
      get() {
        return this.$store.getters['modals/terminal'] === this.$props.uid;
      },

      set(value) {
        if (value) {
          this.$store.dispatch('modals/toggleTerminal', this.$props.uid);
        } else {
          this.$store.dispatch('modals/toggleTerminal', '');
        }
      },
    },
  },

  watch: {
    show(value) {
      if (!value) {
        if (this.ws) this.ws.close();
        if (this.xterm) this.xterm.dispose();

        this.username = '';
        this.passwd = '';
        this.showLoginForm = true;
      } else {
        requestAnimationFrame(() => {
          this.$refs.username.focus();
        });
      }
    },
  },

  methods: {
    open() {
      this.xterm = new Terminal({
        cursorBlink: true,
        fontFamily: 'monospace',
      });

      this.fitAddon = new FitAddon();
      this.xterm.loadAddon(this.fitAddon);

      this.$store.dispatch('modals/toggleTerminal', this.$props.uid);

      if (this.xterm.element) {
        this.xterm.reset();
      }
    },

    close() {
      this.$store.dispatch('modals/toggleTerminal', '');
    },

    connect() {
      let protocolConnectionURL = '';

      if (!this.$refs.form.validate(true)) {
        return;
      }

      this.showLoginForm = false;
      this.$nextTick(() => this.fitAddon.fit());

      if (!this.xterm.element) {
        this.xterm.open(this.$refs.terminal);
      }

      this.fitAddon.fit();
      this.xterm.focus();

      const params = Object.entries({
        user: `${this.username}@${this.$props.uid}`,
        passwd: encodeURIComponent(this.passwd),
        cols: this.xterm.cols,
        rows: this.xterm.rows,
      })
        .map(([k, v]) => `${k}=${v}`)
        .join('&');

      if (window.location.protocol === 'http:') {
        protocolConnectionURL = 'ws';
      } else {
        protocolConnectionURL = 'wss';
      }

      this.ws = new WebSocket(`${protocolConnectionURL}://${window.location.host}/ws/ssh?${params}`);

      this.ws.onopen = () => {
        this.attachAddon = new AttachAddon(this.ws);
        this.xterm.loadAddon(this.attachAddon);
      };

      this.ws.onclose = () => {
        this.attachAddon.dispose();
      };
    },
  },
};

</script>
