<template>
  <v-navigation-drawer
    v-model="showDrawer"
    location="right"
    disable-resize-watcher
    width="300"
    class="theme-drawer bg-v-theme-surface"
    data-test="theme-drawer"
    elevation="8"
  >
    <v-list class="pa-0">
      <v-list-item v-if="terminalThemes.length === 0" class="text-center py-4" title="No themes available" />
      <v-list-item
        v-for="theme in terminalThemes"
        :key="theme.name"
        @click="selectTheme(theme)"
        :class="{ 'v-list-item--active': selectedTheme === theme.name }"
        class="py-3"
        data-test="theme-item"
      >
        <template #prepend>
          <div class="theme-preview" :style="getThemePreviewStyle(theme)">
            <div class="preview-text">
              <div class="preview-line">$ ls -la</div>
              <div class="preview-line file">drwxr-xr-x user</div>
              <div class="preview-line dir">-rw-r--r-- file.txt</div>
            </div>
          </div>
        </template>

        <v-list-item-title>{{ theme.name }}</v-list-item-title>
        <v-list-item-subtitle>{{ theme.description }}</v-list-item-subtitle>

        <template #append>
          <v-icon
            v-if="selectedTheme === theme.name"
            color="primary"
            size="small"
            icon="mdi-check"
          />
        </template>
      </v-list-item>
    </v-list>
  </v-navigation-drawer>
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
const showDrawer = defineModel<boolean>("showDrawer", { default: false });
const terminalThemes = computed(() => terminalThemeStore.terminalThemes || []);

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
.theme-drawer {
  position: absolute !important;
  top: 0 !important;
  right: 0 !important;
  height: 100% !important;
  z-index: 2000;
  box-shadow: -4px 0 12px rgba(0, 0, 0, 0.3);
}

.theme-preview {
  width: 60px;
  height: 40px;
  border-radius: 4px;
  padding: 4px;
  font-family: monospace;
  font-size: 8px;
  line-height: 1.2;
  overflow: hidden;
  margin-right: 8px;
  position: relative;

  .preview-text {
    width: 100%;
    height: 100%;

    .preview-line {
      margin-bottom: 1px;

      &.file {
        opacity: 0.8;
      }

      &.dir {
        opacity: 0.6;
      }
    }
  }
}

.v-list-item {
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));

  &:hover {
    background: rgba(var(--v-theme-primary), 0.08);
  }

  &.v-list-item--active {
    background: rgba(var(--v-theme-primary), 0.12);

    .v-list-item-title {
      color: rgb(var(--v-theme-primary));
      font-weight: 500;
    }
  }
}
</style>
