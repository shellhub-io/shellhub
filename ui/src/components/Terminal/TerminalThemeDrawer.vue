<template>
  <v-navigation-drawer
    v-model="showDrawer"
    location="right"
    temporary
    disable-resize-watcher
    width="300"
    class="theme-drawer bg-v-theme-surface"
    data-test="theme-drawer"
    elevation="8"
  >
    <v-list class="pa-0">
      <v-list-subheader>Font Settings</v-list-subheader>
      <TerminalFontPicker @update:font-settings="$emit('update:fontSettings', $event)" />
      <v-divider />
      <v-list-subheader>Color Theme</v-list-subheader>
      <TerminalThemePicker
        v-model="selectedTheme"
        @update:selected-theme="$emit('update:selectedTheme', $event)"
      />
    </v-list>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { ITerminalTheme } from "@/interfaces/ITerminal";
import TerminalFontPicker from "./TerminalFontPicker.vue";
import TerminalThemePicker from "./TerminalThemePicker.vue";

defineEmits<{
  "update:selectedTheme": [theme: ITerminalTheme];
  "update:fontSettings": [settings: { fontFamily: string; fontSize: number }];
}>();

const selectedTheme = defineModel<string>({ required: true });
const showDrawer = defineModel<boolean>("showDrawer", { required: true });
</script>

<style scoped lang="scss">
.theme-drawer {
  position: absolute !important;
  top: 0 !important;
  right: 0 !important;
  height: 100% !important;
  z-index: 2000;
  box-shadow: -4px 0 12px rgba(0, 0, 0, 0.3);
}
</style>
