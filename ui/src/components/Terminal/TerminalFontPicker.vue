<template>
  <div class="d-flex flex-column align-center pa-4">
    <v-select
      v-model="currentFontFamily"
      :items="availableFonts"
      class="w-100"
      label="Font Family"
      @update:model-value="updateFontSettings"
    />
    <v-number-input
      v-model="currentFontSize"
      control-variant="split"
      variant="outlined"
      density="compact"
      :min="8"
      :max="32"
      hide-details
      inset
      class="w-100"
      label="Font Size"
      @update:model-value="updateFontSettings"
    />
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from "pinia";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

const emit = defineEmits<{
  "update:fontSettings": [settings: { fontFamily: string; fontSize: number }];
}>();

const terminalThemeStore = useTerminalThemeStore();
const { availableFonts, currentFontFamily, currentFontSize } = storeToRefs(terminalThemeStore);

const updateFontSettings = async () => {
  await terminalThemeStore.setFontSettings(currentFontFamily.value, currentFontSize.value);
  emit("update:fontSettings", {
    fontFamily: currentFontFamily.value,
    fontSize: currentFontSize.value,
  });
};
</script>
