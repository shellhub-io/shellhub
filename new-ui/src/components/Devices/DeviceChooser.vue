<template>
  <v-dialog
    v-if="hasAuthorization"
    v-model="show"
    persistent
    max-width="900px"
    min-width="45vw"
  >
    <v-card class="bg-v-theme-surface" data-test="deviceChooser-dialog">
      <v-card-title class="text-headline bg-primary">
        Update account or select three devices
      </v-card-title>

      <v-card-text>
        <p class="ml-2 text-body-2">
          You currently have no subscription to the
          <a :href="url()"> premium plan </a> and the free version is limited to
          3 devices. To unlock access to all devices, you can subscribe to the
          <a :href="url()"> premium plan </a>. Case, If you want to continue on
          the free plan, you need to select three devices.
        </p>
      </v-card-text>

      <div class="mt-2">
        <v-tabs align-tabs="center" color="primary">
          <v-tab
            v-for="item in items"
            :key="item.title"
            :data-test="item.title + '-tab'"
            @click="doAction(item.action)"
          >
            {{ item.title }}
          </v-tab>
        </v-tabs>
      </div>

      <v-card-text class="mb-2 pb-0">
        <DeviceListChooser
          :action="action"
          data-test="deviceListChooser-component"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text data-test="close-btn" @click="close()"> Close </v-btn>

        <v-tooltip :disabled="!disableTooltipOrButton" top>
          <template v-slot:activator="{ props }">
            <span v-on="props">
              <v-btn
                text
                v-bind="props"
                :disabled="disableTooltipOrButton"
                data-test="accept-btn"
                @click="accept()"
              >
                Accept
              </v-btn>
            </span>
          </template>

          <span> You can select 3 devices or less. </span>
        </v-tooltip>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from "vue";
import { useStore } from "../../store";
import DeviceListChooser from "./DeviceListChooser.vue";
import hasPermision from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { INotificationsSuccess, INotificationsError } from "@/interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const action = ref("chooser");
    const items = ref([
      {
        title: "Suggested Devices",
        action: "suggestedDevices",
      },
      {
        title: "All devices",
        action: "allDevices",
      },
    ]);
    const hostname = ref(window.location.hostname);

    const show = computed({
      get() {
        return store.getters["devices/getDeviceChooserStatus"];
      },

      set(value) {
        store.dispatch("devices/setDeviceChooserStatus", value);
      },
    });

    const url = () => `${window.location.protocol}//${hostname.value}/settings/billing`;

    const disableTooltipOrButton = computed(() => (
      (store.getters["devices/getDevicesSelected"].length <= 0
          || store.getters["devices/getDevicesSelected"].length > 3)
        && action.value !== items.value[0].action
    ));

    const equalThreeDevices = computed(() => store.getters["devices/getDevicesSelected"].length === 3);

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermision(authorizer.role[role], actions.device.chooser);
      }

      return false;
    });

    const doAction = async (actionParam: string) => {
      action.value = actionParam;

      if (action.value === items.value[0].action) {
        store.dispatch("devices/getDevicesMostUsed");
      } else if (action.value === items.value[1].action) {
        const data = {
          perPage: 10,
          page: 1,
          filter: store.getters["devices/getFilter"],
          status: "accepted",
          sortStatusField: null,
          sortStatusString: "asc",
        };

        try {
          await store.dispatch("devices/setDevicesForUserToChoose", data);
        } catch (error: any) {
          if (error.response.status === 403) {
            store.dispatch("snackbar/showSnackbarErrorAssociation");
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.deviceList,
            );
          }
        }
      } else {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.deviceList,
        );
      }
    };

    const close = () => {
      store.dispatch("devices/setDeviceChooserStatus", false);
    };

    const sendDevicesChoice = async (devices: Array<any>) => {
      const choices: any = [];
      devices.forEach((device) => {
        choices.push(device.uid);
      });

      try {
        await store.dispatch("devices/postDevicesChooser", { choices });
        store.dispatch("snackbar/showSnackbarSuccessAction", INotificationsSuccess.deviceChooser);

        store.dispatch("devices/refresh");

        close();
      } catch (error: any) {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.deviceChooser);
        throw new Error(error);
      }

      store.dispatch("stats/get");
    };

    const accept = () => {
      if (action.value === items.value[0].action) {
        sendDevicesChoice(store.getters["devices/getDevicesForUserToChoose"]);
      } else {
        sendDevicesChoice(store.getters["devices/getDevicesSelected"]);
      }
    };

    return {
      show,
      url,
      disableTooltipOrButton,
      equalThreeDevices,
      action,
      doAction,
      hasAuthorization,
      items,
      accept,
      close,
    };
  },
  components: { DeviceListChooser },
});
</script>
