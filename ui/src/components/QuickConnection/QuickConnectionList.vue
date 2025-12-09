<template>
  <template v-if="onlineDevices.length === 0">
    <v-card
      class="bg-v-theme-surface mx-auto py-3 border mt-5"
      data-test="no-online-devices"
    >
      <v-card-title class="text-center d-flex justify-center pa-1 mt-5">
        <v-icon
          size="x-large"
          data-test="no-online-devices-icon"
          icon="mdi-laptop-off"
        />
      </v-card-title>
      <p
        class="text-center pa-5"
        data-test="no-online-devices-message"
      >
        There are currently no devices online.
      </p>
    </v-card>
  </template>
  <v-list
    v-else
    ref="rootEl"
    nav
    class="content-card pa-0"
    data-test="devices-list"
  >
    <v-list-item
      v-for="(item, i) in onlineDevices"
      :key="i"
      class="ma-0 pa-2 item border"
      data-test="device-list-item"
      @click="openDialog(item.uid, item.name)"
      @keydown="openTerminalMacro(item)"
    >
      <v-row
        align="center"
        no-gutters
      >
        <v-col
          class="text-center"
          md="3"
          data-test="device-name"
        >
          {{ item.name }}
        </v-col>
        <v-col
          class="text-center text-truncate"
          md="3"
          data-test="device-info"
        >
          <DeviceIcon :icon="item.info.id" />
          <span>{{ item.info.pretty_name }}</span>
        </v-col>
        <v-col
          class="text-truncate text-center"
          md="3"
          data-test="device-ssh-id"
        >
          <v-chip class="bg-grey-darken-4">
            <v-tooltip location="bottom">
              <template #activator="{ props }">
                <CopyWarning
                  ref="copyRef"
                  :copied-item="'Device SSHID'"
                  :macro="getSshid(item)"
                >
                  <template #default="{ copyText }">
                    <span
                      v-bind="props"
                      tabindex="0"
                      class="hover-text text-mono"
                      data-test="copy-id-button"
                      @click.stop="handleSshidClick(item, copyText)"
                      @keypress.enter.stop="handleSshidClick(item, copyText)"
                    >
                      {{ getSshid(item) }}
                    </span>
                  </template>
                </CopyWarning>
              </template>
              <span>Copy SSHID</span>
            </v-tooltip>

            <template #append>
              <v-tooltip location="bottom">
                <template #activator="{ props }">
                  <v-icon
                    v-bind="props"
                    icon="mdi-help-circle-outline"
                    size="small"
                    color="primary"
                    class="ml-2"
                    data-test="sshid-help-btn"
                    @click.stop="forceOpenTerminalHelper(item)"
                  />
                </template>
                <span>What is an SSHID?</span>
              </v-tooltip>
            </template>
          </v-chip>
        </v-col>
        <v-col
          md="3"
          data-test="device-tags"
          class="text-center"
        >
          <div v-if="item.tags[0]">
            <v-tooltip
              v-for="(tag, index) in item.tags"
              :key="index"
              location="bottom"
              :disabled="!showTag(tag.name)"
            >
              <template #activator="{ props }">
                <v-chip
                  size="small"
                  v-bind="props"
                  data-test="tag-chip"
                >
                  {{ displayOnlyTenCharacters(tag.name) }}
                </v-chip>
              </template>

              <span data-test="tag-name">
                {{ tag.name }}
              </span>
            </v-tooltip>
          </div>
          <div v-else>
            <v-chip
              size="small"
              color="grey-darken-2"
              data-test="no-tags-chip"
            >
              No tags
            </v-chip>
          </div>
        </v-col>
      </v-row>
    </v-list-item>
  </v-list>
  <SSHIDHelper
    v-if="showTerminalHelper"
    v-model="showTerminalHelper"
    :sshid="selectedSshid"
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
import SSHIDHelper from "../Terminal/SSHIDHelper.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import { displayOnlyTenCharacters } from "@/utils/string";
import showTag from "@/utils/tag";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import handleError from "@/utils/handleError";
import { IDevice } from "@/interfaces/IDevice";
import useSnackbar from "@/helpers/snackbar";
import useDevicesStore from "@/store/modules/devices";

const props = defineProps<{ filter?: string }>();

const devicesStore = useDevicesStore();
const snackbar = useSnackbar();
const loading = ref(false);
const rootEl = ref<VList>();
const selectedDeviceUid = ref("");
const selectedDeviceName = ref("");
const showDialog = ref(false);
const showTerminalHelper = ref(false);
const selectedSshid = ref("");
const onlineDevices = computed(() => devicesStore.onlineDevices);

const filter = computed(() =>
  btoa(
    JSON.stringify([
      {
        type: "property",
        params: { name: "online", operator: "eq", value: true },
      },
      {
        type: "property",
        params: { name: "name", operator: "contains", value: props.filter },
      },
      { type: "operator", params: { name: "and" } },
    ]),
  ),
);

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

const getSshid = (item: IDevice) =>
  `${item.namespace}.${item.name}@${window.location.hostname}`;

const openTerminalHelper = (item: IDevice) => {
  selectedSshid.value = getSshid(item);
  showTerminalHelper.value = true;
};

const handleSshidClick = (item: IDevice, copyFn: (text: string) => void) => {
  copyFn(getSshid(item));
};

const forceOpenTerminalHelper = (item: IDevice) => {
  openTerminalHelper(item);
};

const openTerminalMacro = (value: IDevice) => {
  let executed = false;

  return useMagicKeys({
    passive: false,
    onEventFired(e) {
      if (
        !executed
        && value
        && e.ctrlKey
        && e.key === "c"
        && e.type === "keydown"
      ) {
        executed = true;
        openTerminalHelper(value);
        e.preventDefault();
      }
    },
  });
};

watch(filter, async () => {
  await getDevices();
});

onMounted(async () => {
  await getDevices();
});

defineExpose({ rootEl });
</script>

<style scoped lang="scss">
.item {
  transition: ease-in-out 200ms;

  &:hover,
  &:focus {
    border-left: 5px solid !important;
    border-right: 5px solid !important;
    border-color: #7284d0 !important;
  }

  &:not(:focus, :hover) {
    opacity: 0.7;
  }
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
