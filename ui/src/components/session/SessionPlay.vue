<template>
  <fragment>
    <v-tooltip bottom>
      <template v-slot:activator="{ on }">
        <v-icon
          v-on="on"
          @click="openPlay()"
        >
          mdi-play-circle
        </v-icon>
      </template>
      <span>Play</span>
    </v-tooltip>
    <v-dialog
      v-model="dialog"
      max-width="1024px"
    >
      <v-card>
        <v-toolbar
          dark
          color="primary"
        >
          <v-toolbar-title>Play</v-toolbar-title>
          <v-spacer />
        </v-toolbar>
        <div ref="playterminal" />
        <v-card>
          <v-card-actions>
            <v-btn
              v-show="!disable"
              :disabled="disable"
              color="primary"
              class="mt-4"
              @click="connect()"
            >
              Play
            </v-btn>
            <v-slider
              v-model="currentTime"
              readonly
              min="0"
              :max="totalLength"
              :label="`${now} - ${end}`"
            />
          </v-card-actions>
        </v-card>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import moment from 'moment';
import 'xterm/css/xterm.css';

export default {
  name: 'SessionPlay',

  props: {
    uid: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      disable: false,
      currentTime: 0,
      totalLength: 0,
      timerDisplayEnd: null,
      timerDisplayNow: null,
    };
  },

  computed: {
    length() {
      return this.logs.length;
    },

    now() {
      return this.timerDisplayNow;
    },

    end() {
      return this.timerDisplayEnd;
    },

  },

  watch: {
    dialog(value) {
      if (!value) {
        this.close();
      }
    },
  },

  updated() {
    this.timerDisplayNow = this.duration(this.currentTime).display;
  },

  async created() {
    await this.$store.dispatch('sessions/getLogSession', this.uid);
    this.logs = this.$store.getters['sessions/getLogSession'];
    this.totalLength = this.duration(null).intervalLength;
    this.timerDisplayEnd = this.duration(null).display;
    this.timerDisplayNow = this.duration(this.currentTime).display;
    // eslint-disable-next-line no-console
    console.log(this.logs);
  },

  methods: {
    openPlay() {
      this.dialog = !this.dialog;
      this.xterm = new Terminal({ // instantiate
        cursorBlink: true,
        fontFamily: 'monospace',
      });
      this.fitAddon = new FitAddon(); // load fit
      this.xterm.loadAddon(this.fitAddon); // adjust screen in container
    },

    duration(timeMs) {
      let interval = 0;
      if (!timeMs) { // not params, will return metrics to max timelengtht
        const max = new Date(this.logs[this.length - 1].time);
        const min = new Date(this.logs[0].time);
        interval = max - min;
      } else { // it will format to the time argument passed
        interval = timeMs;
      }
      const duration = moment.duration(interval, 'milliseconds');
      const TimeObject = {
        seconds: Math.floor(duration.asSeconds()),
        minutes: Math.floor(duration.asMinutes()),
        hours: Math.floor(duration.asHours()),
      };
      return {
        display: `${TimeObject.hours}:${TimeObject.minutes}:${TimeObject.seconds}`, // format to slider label
        intervalLength: interval, // length of slider in Ms
      };
    },

    timer() { // Increments the slider
      if (this.currentTime >= this.totalLength) return;
      this.currentTime += 100;
      this.iterativeTimer = setTimeout(this.timer.bind(null), 100);
    },

    connect() {
      this.disable = true;
      this.xterm.open(this.$refs.playterminal);
      this.$nextTick(() => this.fitAddon.fit());
      this.fitAddon.fit();
      this.xterm.focus();
      this.print(0, this.logs);
      this.timer();
      if (this.xterm.element) { // check already existence
        this.xterm.reset();
      }
    },

    close() {
      if (this.xterm) this.xterm.dispose();
      this.disable = false;
      clearInterval(this.iterativePrinting);
      clearInterval(this.iterativeTimer);
      this.currentTime = 0;
    },

    print(i, obj) {
      this.check_hostname(i);
      if (i === this.logs.length - 1) return;
      const nextObj = this.logs[i + 1];
      const now = new Date(obj.time);
      const future = new Date(nextObj.time);
      setTimeout(this.print.bind(null, i + 1, nextObj), future.getTime() - now.getTime());
    },

    check_hostname(i) {
      if (this.logs[i].message.includes('@')) { // hostname verify,
        this.xterm.write(`${this.logs[i].message}`);
      } else {
        this.xterm.write(`${this.logs[i].message}\r\n`);
      }
    },
  },
};

</script>
