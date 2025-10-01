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
          @click="openDialog(item.uid, item.name)"
          @keydown="openTerminalMacro(item)"
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
                    <CopyWarning
                      ref="copyRef"
                      :copied-item="'Device SSHID'"
                      :bypass="shouldOpenTerminalHelper()"
                      :macro="getSshid(item)"
                    >
                      <template #default="{ copyText }">
                        <span
                          v-bind="props"
                          tabindex="0"
                          class="hover-text"
                          @click.stop="handleSshidClick(item, copyText)"
                          @keypress.enter.stop="handleSshidClick(item, copyText)"
                          data-test="copy-id-button"
                        >
                          {{ getSshid(item) }}
                        </span>
                      </template>
                    </CopyWarning>
                  </template>
                  <span>Copy ID</span>
                </v-tooltip>
              </v-chip>
            </v-col>
            <v-col md="3" data-test="device-tags">
              <div class="text-center">
                <div v-if="item.tags[0]">
                  <v-tooltip v-for="(tag, index) in item.tags" :key="index" location="bottom" :disabled="!showTag(tag.name)">
                    <template #activator="{ props }">
                      <v-chip size="small" v-bind="props" class="mr-1" data-test="tag-chip">
                        {{ displayOnlyTenCharacters(tag.name) }}
                      </v-chip>
                    </template>

                    <span data-test="tag-name">
                      {{ tag.name }}
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
  <TerminalHelper
    v-if="showTerminalHelper"
    v-model="showTerminalHelper"
    :sshid="selectedSshid"
    :user-id="userId"
    :show-checkbox="true"
  />
  <TerminalDialog
    v-model="showDialog"
    :device-uid="selectedDeviceUid"
    :device-name="selectedDeviceName"
    data-test="terminalDialog-component"
  />
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from "vue";
import { VList } from "vuetify/components";
import { useMagicKeys } from "@vueuse/core";
import TerminalDialog from "../Terminal/TerminalDialog.vue";
import TerminalHelper from "../Terminal/TerminalHelper.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import handleError from "@/utils/handleError";
import { IDevice } from "@/interfaces/IDevice";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps<{ filter?: string; }>();

const authStore = useAuthStore();
const devicesStore = useDevicesStore();
const snackbar = useSnackbar();
const loading = ref(false);
const rootEl = ref<VList>();
const selectedDeviceUid = ref("");
const selectedDeviceName = ref("");
const showDialog = ref(false);
const showTerminalHelper = ref(false);
const selectedSshid = ref("");
const userId = authStore.id;
const onlineDevices = computed(() => devicesStore.onlineDevices);

const filter = computed(() => btoa(JSON.stringify([
  {
    type: "property",
    params: { name: "online", operator: "eq", value: true },
  },
  {
    type: "property",
    params: { name: "name", operator: "contains", value: props.filter },
  },
  { type: "operator", params: { name: "and" } },
])));

const openDialog = (deviceUid: string, deviceName: string) => {
  selectedDeviceUid.value = deviceUid;
  selectedDeviceName.value = deviceName;
  showDialog.value = true;
};

const getDevices = async () => {
  try {
    loading.value = true;
    await devicesStore.fetchOnlineDevices(filter.value);
  } catch (error: unknown) {
    snackbar.showError("An error occurred while loading devices.");
    handleError(error);
  }

  loading.value = false;
};

const getSshid = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

const openTerminalHelper = (item: IDevice) => {
  selectedSshid.value = getSshid(item);
  showTerminalHelper.value = true;
};

const shouldOpenTerminalHelper = () => {
  try {
    const dispensedUsers = JSON.parse(localStorage.getItem("dispenseTerminalHelper") || "[]");
    return !dispensedUsers.includes(userId);
  } catch {
    return true;
  }
};

const handleSshidClick = (item: IDevice, copyFn: (text: string) => void) => {
  if (shouldOpenTerminalHelper()) {
    openTerminalHelper(item);
    return;
  }
  copyFn(getSshid(item));
};

const openTerminalMacro = (value: IDevice) => {
  let executed = false;

  return useMagicKeys({
    passive: false,
    onEventFired(e) {
      if (!executed && value && e.ctrlKey && e.key === "c" && e.type === "keydown") {
        executed = true;
        openTerminalHelper(value);
        e.preventDefault();
      }
    },
  });
};

watch(filter, async () => { await getDevices(); });

onMounted(async () => { await getDevices(); });

defineExpose({ rootEl });
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
