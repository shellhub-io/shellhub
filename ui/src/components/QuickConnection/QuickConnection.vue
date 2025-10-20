<template>
  <div>
    <v-btn
      @click="showDialog = true"
      color="primary"
      variant="flat"
      tabindex="0"
      aria-label="Dialog Quick Connection"
      data-test="quick-connection-open-btn"
      prepend-icon="mdi-console"
      block
    >
      Quick Connect
    </v-btn>

    <div>
      <p class="text-caption text-md font-weight-bold text-grey-darken-1 ma-1" data-test="quick-connect-instructions">
        Press <v-chip density="compact" size="small" label>Ctrl+K</v-chip> to Quick Connect!
      </p>
    </div>

    <WindowDialog
      v-model="showDialog"
      threshold="md"
      transition="dialog-bottom-transition"
      title="Quick Connect"
      description="Search and connect to your online devices"
      icon="mdi-console"
      icon-color="primary"
      show-footer
      @close="showDialog = false"
    >
      <v-card-text class="pa-6">
        <v-text-field
          label="Search your online devices!"
          variant="outlined"
          bg-color="bg-v-theme-surface"
          color="primary"
          single-line
          hide-details
          v-model.trim="filter"
          prepend-inner-icon="mdi-magnify"
          density="comfortable"
          data-test="search-text"
          autofocus
          class="shrink mx-1 mt-2"
        />
        <v-row class="mt-4 mb-0 px-5">
          <v-col
            class="px-0"
            v-for="header in headers"
            :key="header.label"
          >
            <p
              class="text-body-2 font-weight-bold text-center"
              :data-test="`${normalizeLabel(header.label)}-header`"
            >
              {{ header.label }}
            </p>
          </v-col>
        </v-row>

        <QuickConnectionList ref="listRef" :filter />
      </v-card-text>

      <template #footer>
        <v-row class="ml-2 justify-space-between font-weight-bold text-grey text-body-2" v-if="!smAndDown">
          <p>
            <v-icon color="primary" data-test="connect-icon" icon="mdi-arrow-u-left-bottom" /> To connect
          </p>
          <p>
            <v-icon color="primary" data-test="navigate-up-icon" icon="mdi-arrow-up" />
            <v-icon color="primary" data-test="navigate-down-icon" icon="mdi-arrow-down" /> To navigate
          </p>
          <p data-test="copy-sshid-instructions">
            <v-kbd class="code text-primary" elevation="0">Ctrl + C</v-kbd>
            To copy SSHID
          </p>
        </v-row>

        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="showDialog = false">Close</v-btn>
      </template>
    </WindowDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useDisplay } from "vuetify";
import { useMagicKeys } from "@vueuse/core";
import QuickConnectionList from "./QuickConnectionList.vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";

const showDialog = ref(false);
const filter = ref("");
const listRef = ref<InstanceType<typeof QuickConnectionList> | null>(null);
const { smAndDown } = useDisplay();
const headers = computed(() => [
  { label: "Hostname" },
  { label: smAndDown.value ? "OS" : "Operating System" },
  { label: "SSHID" },
  { label: "Tags" },
]);

const normalizeLabel = (label: string) => label.toLowerCase().replace(/\s+/g, "-");

useMagicKeys({
  passive: false,
  onEventFired(event) {
    if (event.ctrlKey && event.key.toLowerCase() === "k" && event.type === "keydown") {
      event.preventDefault();
      showDialog.value = !showDialog.value;
    } else if ((event.key === "ArrowDown" || event.key === "ArrowUp") && event.type === "keydown") {
      event.preventDefault();
      listRef.value?.rootEl?.focus?.();
    }
  },
});
</script>

<style scoped lang="scss">
.code {
  font-family: monospace;
  background-color: transparent;
  font-weight: 700;
  font-size: 1rem;
  border: 0;
}
</style>
