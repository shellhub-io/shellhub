<template>
  <v-row
    justify="center"
    align="center"
    class="fill-height terminal-layout"
  >
    <v-col class="fill-height">
      <div
        ref="el"
        class="fill-height"
      />
    </v-col>
  </v-row>

</template>

<script
  setup
  lang="ts"
>
import { computed, onMounted, onUnmounted, ref, watch, nextTick } from "vue";
import { useRoute } from "vue-router";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { useStore } from "../store";

const store = useStore();
const route = useRoute();
const initialized = ref(false);

const xterm = ref<Terminal | null>(null);
const el = ref<HTMLElement | null>(null);

const token = computed(() => route.params.token as string);
const terminalData = computed(() => store.getters["terminals/getTerminal"][token.value]);

const initializeTerminal = async () => {
  if (terminalData.value && el.value) {
    xterm.value = terminalData.value.xterm;
    if (xterm.value && el.value) {
      xterm.value.open(el.value);
      xterm.value.focus();
      terminalData.value.fitAddon.fit();

      terminalData.value.websocket.addEventListener("open", () => {
        const data = terminalData.value.fitAddon.proposeDimensions();
        terminalData.value.websocket.send(JSON.stringify({
          kind: 2,
          data: {
            cols: data.cols,
            rows: data.rows,
          },
        }));
      });
    }
  }
};

onMounted(() => {
  initializeTerminal();
});

onUnmounted(() => {
  if (xterm.value) {
    terminalData.value.websocket.close();
    xterm.value.dispose();
    xterm.value.reset();
    xterm.value.clear();
  }
});
</script>

<style>
.terminal {
  padding: 20px;
}

.terminal-layout {
  background-color: #ff0000;
}
</style>
