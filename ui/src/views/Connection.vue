<template>
  <v-layout
    class="w-100 terminal-layout"
    data-test="device-title"
  >
    <div ref="terminal"/>
  </v-layout>

</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch, nextTick } from "vue";
import { useRoute } from "vue-router";
import { Terminal } from "xterm";
import "xterm/css/xterm.css";
import { useStore } from "../store";

const store = useStore();
const route = useRoute();
const initialized = ref(false);

const xterm = ref<Terminal | null>(null);
const terminal = ref<HTMLElement | null>(null);

const token = computed(() => route.params.token as string);
const terminalData = computed(() => store.getters["terminals/getTerminal"][token.value]);

const initializeTerminal = async () => {
  if (terminalData.value && terminal.value) {
    await nextTick();

    xterm.value = terminalData.value.xterm;
    if (xterm.value && terminal.value) {
      xterm.value.onRender(() => {
        if (!initialized.value) {
          terminalData.value.fitAddon.fit();
          initialized.value = true;
        }
      });

      xterm.value.open(terminal.value);
      xterm.value.focus();
    }
  }
};

watch([token, terminalData], () => {
  if (terminalData.value) {
    initializeTerminal();
  }
});

onMounted(() => {
  if (terminalData.value) {
    initializeTerminal();
  }
});
</script>

<style>
.terminal {
  padding: 20px;
}

.terminal-layout {
  background-color: black;
}
</style>
