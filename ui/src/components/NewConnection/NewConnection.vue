<template>
  <div>
    <v-btn
      @click="dialog = !dialog"
      color="primary"
      tabindex="0"
      variant="elevated"
      aria-label="Dialog New Connection"
      @keypress.enter="dialog = !dialog"
      data-test="device-add-btn"
      :size="size"
      prepend-icon="mdi-link"
    >
      New Connection
    </v-btn>

    <v-dialog
      v-model="dialog"
      width="1000"
      transition="dialog-bottom-transition"
    >
      <v-card class="bg-v-theme-surface">
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
          </v-row>
          <NewConnectionCard ref="list" />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" data-test="close-btn" @click="dialog = !dialog">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
// eslint-disable-next-line import/no-extraneous-dependencies
import { useMagicKeys } from "@vueuse/core";
import { defineComponent, onMounted, computed, ref, onUnmounted } from "vue";
import { useRouter } from "vue-router";
import axios, { AxiosError } from "axios";
import NewConnectionCard from "./NewConnectionCard.vue";
import { useStore } from "../../store";
import handleError from "../../utils/handleError";

export default defineComponent({
  props: {
    size: {
      type: String,
      default: "default",
      required: false,
    },
  },
  setup() {
    const list = ref<InstanceType<typeof NewConnectionCard>>();
    const dialog = ref(false);
    const store = useStore();
    const router = useRouter();
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

      try {
        store.dispatch("devices/search", {
          filter: encodedFilter,
          status: store.getters["devices/getStatus"],
        });
      } catch {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    };

    const isDeviceList = computed(() => router.currentRoute.value.name === "listDevices");

    onMounted(async () => {
      try {
        await store.dispatch("stats/get");
        show.value = true;
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          if (axiosError.response?.status === 403) store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch("snackbar/showSnackbarErrorDefault");
        }
        handleError(error);
      }
    });

    onUnmounted(async () => {
      await store.dispatch("devices/setFilter", null);
    });
    const { ctrlK } = useMagicKeys({
      onEventFired(e) {
        if (e.ctrlKey && e.key === "k" && e.type === "keydown") {
          dialog.value = !dialog.value;
          e.preventDefault();
        }
      },
    });

    const { arrowUD } = useMagicKeys({
      onEventFired(e) {
        if ((e.key === "ArrowDown" || e.key === "ArrowUp") && e.type === "keydown") {
          console.log(list.value);
        }
      },
    });

    return {
      dialog,
      list,
      ctrlK,
      filter,
      arrowUD,
      searchDevices,
      isDeviceList,
    };
  },
  components: { NewConnectionCard },
});
</script>

<style lang="scss" scoped>
.code {
  font-family: monospace;
  font-size: 85%;
  font-weight: normal;
}
</style>
