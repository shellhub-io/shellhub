<template>
  <div ref="el" class="fill-height w-100" :style="{ backgroundColor: terminalData?.xterm?.options?.theme?.background || '#000' }"></div>
</template>

<script setup lang="ts">
import { computed, onMounted, onActivated, onUnmounted, ref, watch } from "vue";
import { useEventListener } from "@vueuse/core";
import { useDisplay } from "vuetify";
import { useRoute } from "vue-router";
import { Terminal } from "@xterm/xterm";
import "@xterm/xterm/css/xterm.css";
import { useStore } from "../store";

const store = useStore();
const route = useRoute();

const xterm = ref<Terminal | null>(null);
const el = ref<HTMLElement | null>(null);

const token = computed(() => route.params.token as string);
const terminalData = computed(() => store.getters["terminals/getTerminal"][token.value]);

const scrollbarColor = computed(() => terminalData.value?.xterm.options.theme.selection);

const { lgAndUp } = useDisplay();

function toMilliseconds(s) {
    return parseFloat(s) * (/\ds$/.test(s) ? 1000 : 1);
}

watch(lgAndUp, async(value) => {
  const drawerElement = document.querySelector(".v-navigation-drawer");

  if (drawerElement) {
    const transitionDuration = getComputedStyle(drawerElement).getPropertyValue("transition-duration");

    setTimeout(() => {
      window.dispatchEvent(new Event("resize"));
    }, toMilliseconds(transitionDuration) * 2);
  }
});

const initializeTerminal = async () => {
  if (terminalData.value && el.value) {
    xterm.value = terminalData.value.xterm;
    if (xterm.value && el.value) {
      xterm.value.open(el.value);
      xterm.value.focus();
      terminalData.value.fitAddon.fit();

      useEventListener(window, "resize", () => {
        terminalData.value.fitAddon.fit();
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

onActivated(() => {
  window.dispatchEvent(new Event("resize"));
});
</script>

<style>
@-moz-document url-prefix() {
  .xterm-viewport {
    overflow: scroll !important;
    scrollbar-width: auto;
    scrollbar-color: v-bind(scrollbarColor) transparent;
  }
}

.xterm-viewport::-webkit-scrollbar {
  height: 8px;
  width: 8px;
}
.xterm-viewport::-webkit-scrollbar-track {
  border-radius: 4px;
}

.xterm-viewport::-webkit-scrollbar-thumb {
  border-radius: 4px;
  background-color: v-bind(scrollbarColor);
}
</style>
