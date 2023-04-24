<template>
  <div>
    <template v-if="onlineDevices.length === 0">
      <v-card class="bg-v-theme-surface mx-auto py-3 border mt-5">
        <v-card-title class="text-center d-flex justify-center pa-1">
          <div>
            <v-icon size="x-large">
              mdi-laptop-off
            </v-icon>
          </div>
        </v-card-title>
        <v-row>
          <v-col class="text-center d-flex justify-center pa-5">
            <p>No online device was found!</p>
          </v-col>
        </v-row>
      </v-card>
    </template>
    <v-list ref="rootEl" nav bg-color="transparent" class="content-card" data-test="devices-list">
      <v-col v-for="(item, i) in onlineDevices" :key="i" class="ma-0 mb-3 pa-0">
        <v-card :key="i">
          <v-list-item @click="open(i)" @keydown="copyMacro(sshidAddress(item))" :key="i" class="ma-0 pa-0 card">
            <v-row align="center" no-gutters>
              <TerminalDialog :uid="item.uid" :online="item.online" ref="terminalFn" data-test="terminalDialog-component" />
              <v-col class="text-center" md="3">
                {{ item.name }}
              </v-col>
              <v-col class="text-center pr-6 text-truncate" md="3">
                <DeviceIcon :icon="item.info.id" />
                <span>{{ item.info.pretty_name }}</span>
              </v-col>
              <v-col class="text-truncate" md="3">
                <v-chip class="bg-grey-darken-4">
                  <v-tooltip location="bottom">
                    <template v-slot:activator="{ props }">
                      <span
                        v-bind="props"
                        @click="copyText(sshidAddress(item))"
                        @keypress="copyText(sshidAddress(item))"
                        class="hover-text">
                        {{ sshidAddress(item) }}
                      </span>
                    </template>
                    <span>Copy ID</span>
                  </v-tooltip>
                </v-chip>
              </v-col>
              <v-col md="3">
                <div class="text-center">
                  <div v-if="item.tags[0]">
                    <v-tooltip v-for="(tag, index) in item.tags" :key="index" location="bottom" :disabled="!showTag(tag)">
                      <template #activator="{ props }">
                        <v-chip size="small" v-bind="props" v-on="props" class="mr-1">
                          {{ displayOnlyTenCharacters(tag) }}
                        </v-chip>
                      </template>

                      <span v-if="showTag(tag)">
                        {{ tag }}
                      </span>
                    </v-tooltip>
                  </div>
                  <div v-else>
                    <v-chip size="small" color="grey-darken-2"> No tags </v-chip>
                  </div>

                </div>
              </v-col>
            </v-row>
          </v-list-item>
        </v-card>
      </v-col>
    </v-list>
  </div>
</template>

<script lang="ts">
import { useMagicKeys } from "@vueuse/core";
import { defineComponent, ref, onMounted, computed, watch } from "vue";
import { VList } from "vuetify/components";
import { useRouter } from "vue-router";
import TerminalDialog from "../Terminal/TerminalDialog.vue";
import { useStore } from "../../store";
import { formatDate } from "../../utils/formateDate";
import { displayOnlyTenCharacters } from "../../utils/string";
import showTag from "../../utils/tag";
import DeviceIcon from "../Devices/DeviceIcon.vue";
import {
  INotificationsCopy,
  INotificationsError,
} from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";
import { IDevice } from "../../interfaces/IDevice";

export default defineComponent({
  setup(props, context) {
    const terminalFn = ref<InstanceType<typeof TerminalDialog>>();
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const itemsPerPage = ref(10);
    const page = ref();
    const rootEl = ref<VList>();
    context.expose({ rootEl });
    let encodedFilter = "";
    const filterToEncodeBase64 = [
      {
        type: "property",
        params: { name: "online", operator: "eq", value: true },
      },
    ];
    encodedFilter = btoa(JSON.stringify(filterToEncodeBase64));
    const filter = ref(encodedFilter);

    interface Device {
      online: boolean
    }

    const devices = computed(() => store.getters["devices/list"]);
    const onlineDevices = computed(() => devices.value.filter((item: Device) => item.online));
    const open = (i: number) => {
      if (terminalFn.value !== undefined) {
        const items = terminalFn.value;
        const item = items[i];
        item.open();
      }
    };

    onMounted(async () => {
      try {
        loading.value = true;
        await store.dispatch("devices/fetch", {
          perPage: itemsPerPage.value,
          page: page.value,
          status: "accepted",
          filter: filter.value,
          sortStatusField: "",
          sortStatusString: "",
        });
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceList,
        );
        handleError(error);
      } finally {
        loading.value = false;
      }
    });

    const getDevices = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;

        await store.dispatch("devices/fetch", {
          perPage: perPagaeValue,
          page: pageValue,
          status: "accepted",
          filter: filter.value,
          sortStatusField: store.getters["devices/getSortStatusField"],
          sortStatusString: store.getters["devices/getSortStatusString"],
        });

        loading.value = false;
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceList,
        );
        handleError(error);
      }
    };

    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };

    watch(itemsPerPage, async () => {
      await getDevices(itemsPerPage.value, page.value);
    });

    const goToNamespace = (namespace: string) => {
      router.push({ name: "namespaceDetails", params: { id: namespace } });
    };

    const redirectToDevice = (deviceId: string) => {
      router.push({ name: "detailsDevice", params: { id: deviceId } });
    };

    const sshidAddress = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

    const refreshDevices = () => {
      getDevices(itemsPerPage.value, page.value);
    };

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.deviceSSHID,
        );
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
            store.dispatch(
              "snackbar/showSnackbarCopy",
              INotificationsCopy.deviceSSHID,
            );
            e.preventDefault();
          }
        },
      });
    };

    return {
      itemsPerPage,
      page,
      loading,
      devices,
      onlineDevices,
      rootEl,
      terminalFn,
      open,
      showTag,
      displayOnlyTenCharacters,
      formatDate,
      goToNamespace,
      changeItemsPerPage,
      redirectToDevice,
      sshidAddress,
      copyText,
      copyMacro,
      refreshDevices,
    };
  },
  components: {
    DeviceIcon,
    TerminalDialog,
  },
});
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