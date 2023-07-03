<template>
  <div>
    <v-list-item v-if="notificationStatus">
      <v-btn
        v-bind="$attrs"
        size="x-small"
        color="primary"
        data-test="notification-btn"
        @click="dialog = !dialog"
      >
        <v-icon> {{ icon }} </v-icon>
        Accept
      </v-btn>
    </v-list-item>
    <v-list-item
      @click="dialog = !dialog"
      v-else>
      <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
        <template v-slot:activator="{ props }">
          <span v-bind="props">
            <v-list-item-title data-test="action-item" v-on="props">
              <v-icon> {{ icon }}</v-icon>
              {{ capitalizeText(action) }}
            </v-list-item-title>
          </span>
        </template>
        <span> You don't have this kind of authorization. </span>
      </v-tooltip>
    </v-list-item>
    <v-dialog max-width="450px" v-model="dialog" @click:outside="close" v-bind="$attrs">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-5 bg-primary">
          Are you sure?
        </v-card-title>
        <v-divider />
        <v-container>

          <v-alert
            v-if="billingActive"
            type="warning"
            text="Please note that each accepted device will be billed individually.
          Removing a device later will not affect the charges for that device." />
          <v-card-text class="mt-4 mb-0 pb-1">
            <p class="text-body-2 mb-2">
              You are about to {{ action }} this device.
            </p>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn variant="text" @click="close()"> Close </v-btn>

            <v-btn variant="text" @click="doAction()">
              {{ action }}
            </v-btn>
          </v-card-actions>
        </v-container>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "../../store";
import { authorizer, actions } from "../../authorizer";
import hasPermission from "../../utils/permission";
import { INotificationsError } from "../../interfaces/INotifications";
import { capitalizeText } from "../../utils/string";
import handleError from "../../utils/handleError";

export default defineComponent({
  props: {
    uid: {
      type: String,
      required: true,
    },

    notificationStatus: {
      type: Boolean,
      required: false,
      default: false,
    },

    action: {
      type: String,
      default: "accept",
      validator: (value: string) => ["accept", "reject", "remove"].includes(value),
    },

    show: {
      type: Boolean,
      default: false,
      required: false,
    },
  },
  setup(props, ctx) {
    const store = useStore();

    const billingActive = computed(() => store.getters["billing/active"]);

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.device[props.action],
        );
      }

      return false;
    });

    const dialog = ref(false);

    const close = () => {
      dialog.value = false;
      ctx.emit("update:show", false);
    };

    const refreshStats = async () => {
      try {
        await store.dispatch("stats/get");
      } catch (error: unknown) {
        store.dispatch("snackbar/showSnackbarErrorDefault");
        handleError(error);
      }
    };

    const refreshDevices = () => {
      try {
        ctx.emit("update");
        if (
          window.location.pathname === "/devices/pending"
          || window.location.pathname === "/devices"
        ) {
          store.dispatch("devices/refresh");
          store.dispatch("notifications/fetch");
        }

        close();
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.deviceList,
        );
        handleError(error);
      }
    };

    const removeDevice = async () => {
      try {
        await store.dispatch("devices/remove", props.uid);
        refreshDevices();
      } catch (error: unknown) {
        close();

        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceDelete,
        );
        handleError(error);
      }
    };

    const rejectDevice = async () => {
      try {
        await store.dispatch("devices/reject", props.uid);
        refreshStats();
        refreshDevices();
      } catch (error: unknown) {
        close();

        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceRejecting,
        );
        handleError(error);
      }
    };

    const acceptDevice = async () => {
      try {
        await store.dispatch("devices/accept", props.uid);
        refreshStats();
        refreshDevices();
      } catch (error: unknown) {
        if (axios.isAxiosError(error)) {
          const axiosError = error as AxiosError;
          if (axiosError.response?.status === 402) {
            store.dispatch("users/setStatusUpdateAccountDialogByDeviceAction", true);
          }
        }
        close();

        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.deviceAccepting,
        );
        handleError(error);
      }
    };

    const doAction = () => {
      if (hasAuthorization.value) {
        switch (props.action) {
          case "accept":
            acceptDevice();
            break;
          case "reject":
            rejectDevice();
            break;
          case "remove":
            removeDevice();
            break;
          default:
        }
      } else {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
      }
    };

    const findIcon = () => {
      switch (props.action) {
        case "accept":
          return "mdi-check";
        case "reject":
          return "mdi-close";
        case "remove":
          return "mdi-delete";
        default:
          return "";
      }
    };

    const icon = ref(findIcon());

    return {
      icon,
      doAction,
      close,
      hasAuthorization,
      capitalizeText,
      dialog,
      billingActive,
    };
  },
});
</script>
