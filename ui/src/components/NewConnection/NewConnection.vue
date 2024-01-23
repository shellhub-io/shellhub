<template>
  <div>
    <v-btn
      @click="dialog = !dialog"
      color="primary"
      tabindex="0"
      variant="elevated"
      aria-label="Dialog New Connection"
      data-test="new-connection-add-btn"
      prepend-icon="mdi-link"
    >
      New Connection
    </v-btn>
    <div>
      <p
        class="text-caption text-md font-weight-bold text-grey-darken-1 ma-1">Press "Ctrl + K" to Quick Connect!</p>
    </div>
    <v-dialog
      v-model="dialog"
      width="1000"
      transition="dialog-bottom-transition"
    >
      <v-card class="bg-v-theme-surface content">
        <div class="pa-5">
          <v-row>
            <v-col>
              <v-text-field
                label="Search your online devices with ease!"
                variant="solo"
                color="primary"
                single-line
                hide-details
                v-model.trim="filter"
                v-on:keyup="searchDevices"
                prepend-inner-icon="mdi-magnify"
                density="comfortable"
                data-test="search-text"
                autofocus
                class="shrink mx-1"
              />
            </v-col>
          </v-row>
        </div>
        <v-card-text class="mt-4 mb-0 pb-1 flex">
          <v-row>
            <v-col>
              <p class="text-body-2 mb-2 font-weight-bold text-center">
                Hostname
              </p>
            </v-col>
            <v-col>
              <p class="text-body-2 mb-2 font-weight-bold text-center">
                Operating System
              </p>
            </v-col>
            <v-col>
              <p class="text-body-2 mb-2 font-weight-bold text-center">
                SSHID
              </p>
            </v-col>
            <v-col>
              <p class="text-body-2 mr-3 font-weight-bold text-center">
                Tags
              </p>
            </v-col>
          </v-row>
          <NewConnectionList ref="list" />
        </v-card-text>
        <v-card-actions>
          <v-row class="ml-2">
            <v-col>
              <p class="text-body-2 mb-0 font-weight-bold text-grey-darken-1">
                <v-icon color="#7284D0" data-test="connect-icon"> mdi-arrow-u-left-bottom </v-icon>To connect</p>
            </v-col>
            <v-col>
              <p class="text-body-2 mb-0 font-weight-bold text-grey-darken-1">
                <v-icon color="#7284D0" data-test="navigate-up-icon"> mdi-arrow-up </v-icon>
                <v-icon color="#7284D0" data-test="navigate-down-icon"> mdi-arrow-down  </v-icon>
                To navigate
              </p>
            </v-col>
            <v-col>
              <p
                class="text-body-2 font-weight-bold text-grey-darken-1"
                data-test="copy-sshid-instructions">Press "Ctrl + C" to copy SSHID</p>
            </v-col>
          </v-row>
          <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
// eslint-disable-next-line import/no-extraneous-dependencies
import { useMagicKeys } from "@vueuse/core";
import { watch, ref, onUnmounted } from "vue";
import axios, { AxiosError } from "axios";
import NewConnectionList from "./NewConnectionList.vue";
import { useStore } from "../../store";
import handleError from "../../utils/handleError";

const list = ref<InstanceType<typeof NewConnectionList>>();
const dialog = ref(false);
const store = useStore();
const filter = ref("");
const show = ref(false);
const searchDevices = () => {
  let encodedFilter = "";

  if (filter.value) {
    const filterToEncodeBase64 = [
      {
        type: "property",
        params: { name: "name", operator: "contains", value: filter.value },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
  }

  if (dialog.value === false) {
    encodedFilter = "";
  }

  try {
    store.dispatch("devices/searchQuickConnection", {
      page: store.getters["devices/getPage"],
      perPage: store.getters["devices/getPerPage"],
      filter: encodedFilter,
      status: store.getters["devices/getStatus"],
    });
  } catch {
    store.dispatch("snackbar/showSnackbarErrorDefault");
  }
};

watch(dialog, async (value) => {
  if (!value) return;

  try {
    await store.dispatch("stats/get");
    show.value = true;
  } catch (error: unknown) {
    const axiosError = error as AxiosError;
    switch (axios.isAxiosError(error)) {
      case axiosError.response?.status === 403: {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
        break; }
      default: store.dispatch("snackbar/showSnackbarErrorDefault");
    }
    handleError(error);
  }
});

onUnmounted(async () => {
  await store.dispatch("devices/setFilter", null);
});

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const keyboardMacros = useMagicKeys({
  passive: false,
  onEventFired(e) {
    if (e.ctrlKey && e.key === "k" && e.type === "keydown") {
      e.preventDefault();
      dialog.value = !dialog.value;
    } else if ((e.key === "ArrowDown" || e.key === "ArrowUp") && e.type === "keydown") {
      e.preventDefault();
      list.value?.rootEl?.focus();
    }
  },
});
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}

.content {
  min-height: 70vh;
  max-height: 70vh;
  overflow: auto;
}

</style>
