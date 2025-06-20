<template>
  <div>
    <v-btn
      @click="dialog = !dialog"
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

    <v-dialog v-model="dialog" width="1000" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface content" max-height="700">
        <div class="pa-5">
          <v-row>
            <v-col>
              <v-text-field
                label="Search your online devices!"
                variant="outlined"
                bg-color="bg-v-theme-surface"
                color="primary"
                single-line
                hide-details
                v-model.trim="filter"
                @keyup="searchDevices"
                prepend-inner-icon="mdi-magnify"
                density="comfortable"
                data-test="search-text"
                autofocus
                class="shrink mx-1 mt-2"
              />

            </v-col>
          </v-row>
        </div>

        <v-card-text class="mt-4 mb-0 pb-1 flex">
          <v-row>
            <v-col
              v-for="header in headers"
              :key="header.label"
            >
              <p
                class="text-body-2 mb-2 font-(weight-bold) text-center"
                :data-test="`${normalizeLabel(header.label)}-header`"
              >
                {{ header.label }}
              </p>
            </v-col>
          </v-row>

          <QuickConnectionList ref="listRef" />
        </v-card-text>

        <v-card-actions>
          <v-row class="ml-2">
            <v-col>
              <p class="text-body-2 mb-0 font-weight-bold text-grey-darken-1">
                <v-icon color="#7284D0" data-test="connect-icon">mdi-arrow-u-left-bottom</v-icon> To connect
              </p>
            </v-col>
            <v-col>
              <p class="text-body-2 mb-0 font-weight-bold text-grey-darken-1">
                <v-icon color="#7284D0" data-test="navigate-up-icon">mdi-arrow-up</v-icon>
                <v-icon color="#7284D0" data-test="navigate-down-icon">mdi-arrow-down</v-icon> To navigate
              </p>
            </v-col>
            <v-col>
              <p class="text-body-2 font-weight-bold text-grey-darken-1" data-test="copy-sshid-instructions">
                Press "Ctrl + C" to copy SSHID
              </p>
            </v-col>
          </v-row>

          <v-btn variant="text" data-test="close-btn" @click="dialog = false">Close</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from "vue";
import { useMagicKeys } from "@vueuse/core";
import axios, { AxiosError } from "axios";
import QuickConnectionList from "./QuickConnectionList.vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const dialog = ref(false);
const snackbar = useSnackbar();
const filter = ref("");
const listRef = ref<InstanceType<typeof QuickConnectionList> | null>(null);
const store = useStore();

const headers = [
  { label: "Hostname" },
  { label: "Operating System" },
  { label: "SSHID" },
  { label: "Tags" },
];

const normalizeLabel = (label: string) => label.toLowerCase().replace(/\s+/g, "-");

useMagicKeys({
  passive: false,
  onEventFired(event) {
    if (event.ctrlKey && event.key.toLowerCase() === "k" && event.type === "keydown") {
      event.preventDefault();
      dialog.value = !dialog.value;
    } else if ((event.key === "ArrowDown" || event.key === "ArrowUp") && event.type === "keydown") {
      event.preventDefault();
      listRef.value?.rootEl?.focus?.();
    }
  },
});

const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value.trim()) {
    const filterObj = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: filter.value },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterObj));
  }

  if (!dialog.value) {
    encodedFilter = "";
  }

  store.dispatch("devices/searchQuickConnection", {
    page: store.getters["devices/getPage"],
    perPage: store.getters["devices/getPerPage"],
    filter: encodedFilter,
    status: store.getters["devices/getStatus"],
  }).catch(() => {
    snackbar.showError("An error occurred while searching for devices.");
  });
};

watch(dialog, async (isOpen) => {
  if (!isOpen) return;

  try {
    await store.dispatch("stats/get");
  } catch (err: unknown) {
    const error = err as AxiosError;
    if (axios.isAxiosError(error) && error.response?.status === 403) {
      snackbar.showError("You don't have permission to access this feature.");
    } else {
      snackbar.showError("An error occurred while fetching stats.");
    }
    handleError(error);
  }
});

onUnmounted(() => {
  store.dispatch("devices/setFilter", "");
});
</script>

<style scoped lang="scss">
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
