<template>
  <v-app :theme="theme">
    <v-system-bar
      v-if="showNewUiBanner"
      :height="40"
      class="new-ui-bar"
      data-test="new-ui-bar"
    >
      <span class="new-ui-bar-content">
        <code class="new-ui-bar-code">npm uninstall vue</code>
        <span>
          &mdash; 299 <code class="new-ui-bar-code">v-model</code> replaced with <code class="new-ui-bar-code">useState</code>
        </span>
        <a
          href="/v2/ui/"
          class="new-ui-bar-link"
        >
          <v-icon size="14" class="mr-1">
            mdi-react
          </v-icon>
          useNewUI() &rarr;
        </a>
      </span>
      <v-btn
        icon
        variant="text"
        size="x-small"
        class="new-ui-bar-close"
        data-test="new-ui-bar-close"
        @click="dismissNewUiBanner"
      >
        <v-icon size="14">
          mdi-close
        </v-icon>
      </v-btn>
    </v-system-bar>

    <component
      :is="layout"
      :data-test="layout + '-component'"
    />
  </v-app>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import LoginLayout from "./layouts/LoginLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import useLayoutStore from "@/store/modules/layout";
import "./assets/global.css";

const components = {
  AppLayout,
  LoginLayout,
};
const layoutStore = useLayoutStore();
const layout = computed(() => components[layoutStore.layout as keyof typeof components]);
const theme = computed(() => layoutStore.theme);

const showNewUiBanner = ref(true);
const dismissNewUiBanner = () => {
  showNewUiBanner.value = false;
};
</script>

<style scoped>
.new-ui-bar {
  justify-content: center;
  font-size: 0.8125rem;
  background: linear-gradient(90deg, #42b883 0%, #667ACC 50%, #61dafb 100%);
  color: #fff;
  letter-spacing: 0.01em;
}

.new-ui-bar-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.new-ui-bar-code {
  background: rgba(0, 0, 0, 0.25);
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-family: monospace;
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.new-ui-bar-link {
  display: inline-flex;
  align-items: center;
  color: #282c34;
  font-weight: 700;
  font-family: monospace;
  font-size: 0.7rem;
  text-decoration: none;
  background: #61dafb;
  padding: 3px 12px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  transition: background 0.2s;
}

.new-ui-bar-link:hover {
  background: #7ae2fc;
}

.new-ui-bar-close {
  position: absolute;
  right: 8px;
  color: rgba(255, 255, 255, 0.8) !important;
}
</style>
