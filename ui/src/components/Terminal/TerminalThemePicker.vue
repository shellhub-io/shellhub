<template>
  <v-list-item
    v-if="terminalThemes.length === 0"
    class="text-center py-4"
    title="No themes available"
  />
  <v-list-item
    v-for="theme in terminalThemes"
    v-else
    :key="theme.name"
    :active="isThemeSelected(theme.name)"
    class="py-3 border-b"
    data-test="theme-item"
    hover
    @click="selectTheme(theme)"
  >
    <template #prepend>
      <div
        class="theme-preview overflow-hidden rounded pa-1 mr-2"
        :style="getThemePreviewStyle(theme)"
      >
        <code class="w-100 d-block">$ ls</code>
        <code class="w-100 d-block opacity-80">file.txt</code>
        <code class="w-100 d-block opacity-60">home</code>
      </div>
    </template>
    <v-list-item-title :class="{ 'text-primary font-weight-medium': isThemeSelected(theme.name) }">{{ theme.name }}</v-list-item-title>
    <v-list-item-subtitle>{{ theme.description }}</v-list-item-subtitle>
    <template #append>
      <v-icon
        v-if="isThemeSelected(theme.name)"
        color="primary"
        size="small"
        icon="mdi-check"
      />
    </template>
  </v-list-item>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ITerminalTheme } from "@/interfaces/ITerminal";
import useTerminalThemeStore from "@/store/modules/terminal_theme";

const emit = defineEmits<{
  "update:selectedTheme": [theme: ITerminalTheme];
}>();

const terminalThemeStore = useTerminalThemeStore();
const selectedTheme = defineModel<string>({ required: true });
const terminalThemes = computed(() => terminalThemeStore.terminalThemes || []);
const isThemeSelected = (themeName: string) => selectedTheme.value === themeName;
const selectTheme = (theme: ITerminalTheme) => {
  selectedTheme.value = theme.name;
  emit("update:selectedTheme", theme);
};

const getThemePreviewStyle = (theme: ITerminalTheme) => ({
  backgroundColor: theme.colors.background,
  color: theme.colors.foreground,
  border: `1px solid ${theme.colors.selection || theme.colors.black || "#666"}`,
});
</script>

<style scoped lang="scss">
.theme-preview {
  width: 60px;
  height: 40px;
  font-size: 8px;
  line-height: 1.25;
  font-family: monospace;
}
</style>
