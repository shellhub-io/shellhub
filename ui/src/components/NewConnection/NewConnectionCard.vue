<template>
  <v-list ref="rootEl" nav bg-color="transparent" class="pa-0 ma-0">
    <v-col v-for="(item, i) in onlineDevices" :key="i" class="ma-0 pa-0">
      <v-card :key="i">
        <v-list-item href="/" class="ma-1 pa-0">
          <v-row align="center">
            <v-col class="text-center">
              {{ item.name }}
            </v-col>
            <v-col class="text-center pr-6 text-truncate">
              <DeviceIcon :icon="item.info.id" />
              <span>{{ item.info.pretty_name }}</span>
            </v-col>
            <v-col class="text-truncate">
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
          </v-row>
        </v-list-item>
      </v-card>
      <div class="ma-1">
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
  </v-list>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, computed } from "vue";
import { VList } from "vuetify/components";
import { useRouter } from "vue-router";
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
    const store = useStore();
    const router = useRouter();
    const loading = ref(false);
    const filter = ref("");
    const itemsPerPage = ref(10);
    const page = ref(1);
    const rootEl = ref<VList>();
    context.expose({ rootEl });

    const deviceDeleteShow = ref([]);

    const devices = computed(() => store.getters["devices/list"]);
    const onlineDevices = computed(() => devices.value.filter((item) => item.online));
    const numberDevices = computed<number>(
      () => store.getters["devices/getNumberDevices"],
    );

    onMounted(async () => {
      try {
        loading.value = true;
        await store.dispatch("devices/fetch", {
          perPage: itemsPerPage.value,
          page: page.value,
          filter: "",
          status: "accepted",
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

        const hasDevices = await store.dispatch("devices/fetch", {
          perPage: perPagaeValue,
          page: pageValue,
          status: "accepted",
          filter: filter.value,
          sortStatusField: store.getters["devices/getSortStatusField"],
          sortStatusString: store.getters["devices/getSortStatusString"],
        });

        if (!hasDevices) {
          page.value--;
        }

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

    const goToNamespace = (namespace: string) => {
      router.push({ name: "namespaceDetails", params: { id: namespace } });
    };

    const redirectToDevice = (deviceId: string) => {
      router.push({ name: "detailsDevice", params: { id: deviceId } });
    };

    const sshidAddress = (item: IDevice) => `${item.namespace}.${item.name}@${window.location.hostname}`;

    const copyText = (value: string | undefined) => {
      if (value) {
        navigator.clipboard.writeText(value);
        store.dispatch(
          "snackbar/showSnackbarCopy",
          INotificationsCopy.deviceSSHID,
        );
      }
    };

    const refreshDevices = () => {
      getDevices(itemsPerPage.value, page.value);
    };

    return {
      itemsPerPage,
      page,
      loading,
      devices,
      deviceDeleteShow,
      numberDevices,
      onlineDevices,
      showTag,
      displayOnlyTenCharacters,
      formatDate,
      goToNamespace,
      changeItemsPerPage,
      redirectToDevice,
      sshidAddress,
      copyText,
      refreshDevices,
    };
  },
  components: {
    DeviceIcon,
  },
});
</script>

<style scoped>
.list:hover,
.list:focus  {
  border: 2px green solid;
}

</style>
