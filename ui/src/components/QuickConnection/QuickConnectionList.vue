<template>
  <template v-if="onlineDevices.length === 0">
    <v-card class="bg-v-theme-surface mx-auto py-3 border mt-5" data-test="no-online-devices">
      <v-card-title class="text-center d-flex justify-center pa-1">
        <div>
          <v-icon size="x-large" data-test="no-online-devices-icon">
            mdi-laptop-off
          </v-icon>
        </div>
      </v-card-title>
      <v-row>
        <v-col class="text-center d-flex justify-center pa-5">
          <p data-test="no-online-devices-message">There are currently no devices online.</p>
        </v-col>
      </v-row>
    </v-card>
  </template>
  <v-list ref="rootEl" nav bg-color="transparent" class="content-card" data-test="devices-list">
    <v-col v-for="(item, i) in onlineDevices" :key="i" class="ma-0 mb-3 pa-0">
      <v-card :key="i" data-test="device-card">
        <v-list-item
          @click="openDialog(item.uid)"
          @keydown="copyMacro(sshidAddress(item))"
          :key="i"
          class="ma-0 pa-0 card"
          data-test="device-list-item"
        >
          <v-row align="center" no-gutters>
            <v-col class="text-center" md="3" data-test="device-name">
              {{ item.name }}
            </v-col>
            <v-col class="text-center pr-6 text-truncate" md="3" data-test="device-info">
              <DeviceIcon :icon="item.info.id" />
              <span>{{ item.info.pretty_name }}</span>
            </v-col>
            <v-col class="text-truncate" md="3" data-test="device-ssh-id">
              <v-chip class="bg-grey-darken-4">
                <v-tooltip location="bottom">
                  <template v-slot:activator="{ props }">
                    <span
                      v-bind="props"
                      @click.stop="copyText(sshidAddress(item))"
                      @keypress.stop="copyText(sshidAddress(item))"
                      class="hover-text"
                      data-test="copy-id-button">
                      {{ sshidAddress(item) }}
                    </span>

                  </template>
                  <span>Copy ID</span>
                </v-tooltip>
              </v-chip>
            </v-col>
            <v-col md="3" data-test="device-tags">
              <div class="text-center">
                <div v-if="item.tags[0]">
                  <v-tooltip v-for="(tag, index) in item.tags" :key="index" location="bottom" :disabled="!showTag(tag)">
                    <template #activator="{ props }">
                      <v-chip size="small" v-bind="props" class="mr-1" data-test="tag-chip">
                        {{ displayOnlyTenCharacters(tag) }}
                      </v-chip>
                    </template>

                    <span v-if="showTag(tag)" data-test="tag-name">
                      {{ tag }}
                    </span>
                  </v-tooltip>
                </div>
                <div v-else>
                  <v-chip size="small" color="grey-darken-2" data-test="no-tags-chip"> No tags </v-chip>
                </div>
              </div>
            </v-col>
          </v-row>
        </v-list-item>
      </v-card>
    </v-col>
  </v-list>

  <TerminalDialog
    v-model="showDialog"
    :deviceUid="selectedDeviceUid"
    data-test="terminalDialog-component"
  />
</template>

<script setup lang="ts">
import { useMagicKeys } from "@vueuse/core";
import { ref, onMounted, computed, watch } from "vue";
import { VList } from "vuetify/components";
import TerminalDialog from "../Terminal/TerminalDialog.vue";
import { useStore } from "@/store";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import handleError from "@/utils/handleError";
import { IDevice } from "@/interfaces/IDevice";
import useSnackbar from "@/helpers/snackbar";

interface Device {
  online: boolean
}

const store = useStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref();
const rootEl = ref<VList>();
const selectedDeviceUid = ref("");
const showDialog = ref(false);

defineExpose({ rootEl });

let encodedFilter = "";

const filterToEncodeBase64 = [
  {
    type: "property",
    params: { name: "online", operator: "eq", value: true },
  },
];
encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));

const filter = ref(encodedFilter);

const devices = computed(() => store.getters["devices/listQuickConnection"]);

const onlineDevices = computed(() => devices.value.filter((item: Device) => item.online));

const openDialog = (deviceUid: string) => {
  selectedDeviceUid.value = deviceUid;
  showDialog.value = true;
};

onMounted(async () => {
  try {
    loading.value = true;
    await store.dispatch("devices/fetchQuickDevices", {
      perPage: itemsPerPage.value,
      page: page.value,
      status: "accepted",
      filter: filter.value,
      sortStatusField: "",
      sortStatusString: "",
    });
  } catch (error: unknown) {
    snackbar.showError("An error occurred while loading devices.");
    handleError(error);
  } finally {
    loading.value = false;
  }
});

const getDevices = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;

    await store.dispatch("devices/fetchQuickDevices", {
      perPage: perPageValue,
      page: pageValue,
      status: "accepted",
      filter: filter.value,
      sortStatusField: store.getters["devices/getSortStatusField"],
      sortStatusString: store.getters["devices/getSortStatusString"],
    });

    loading.value = false;
  } catch (error: unknown) {
    snackbar.showError("An error occurred while loading devices.");
    handleError(error);
  }
};

watch(itemsPerPage, async () => {
  await getDevices(itemsPerPage.value, page.value);
});

const sshidAddress = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    snackbar.showInfo("Device SSHID copied to clipboard.");
  }
};

const copyMacro = (value: string | undefined) => {
  let executed = false;

  return useMagicKeys({
    passive: false,
    onEventFired(e) {
      if (!executed && value && e.ctrlKey && e.key === "c" && e.type === "keydown") {
        executed = true;
        navigator.clipboard.writeText(value);
        snackbar.showInfo("Device SSHID copied to clipboard.");
        e.preventDefault();
      }
    },
  });
};
</script>

<style scoped>
.card:hover,
.card:focus {
  border-left: 5px solid #7284d0;
  border-right: 5px solid #7284d0;
  transition: ease-in-out 200ms;
}

.card:not(:focus, :hover) {
  opacity: 0.70;
}

.content-card {
  max-height: 45vh;
  overflow: auto;
}

::-webkit-scrollbar {
    width: 6px;
  }
    ::-webkit-scrollbar-track {
      background-color: rgb(255 255 255 / 10%);
      border-radius: 10px;
    }
    ::-webkit-scrollbar-thumb {
      background-color: rgb(0 0 0 / 80%);
      border-radius: 10px;
    }
</style>
