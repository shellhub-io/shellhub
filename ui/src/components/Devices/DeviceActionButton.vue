<template>
  <div>
    <v-btn
      v-bind="$attrs"
      size="x-small"
      color="primary"
      v-if="isInNotification"
      data-test="notification-action-button"
      @click="showDialog = true"
    >
      <v-icon>{{ icon }}</v-icon>
      Accept
    </v-btn>
    <v-list-item v-else @click="showDialog = true" data-test="list-item">
      <v-tooltip location="bottom" class="text-center" :disabled="hasAuthorization">
        <template v-slot:activator="{ props }">
          <span v-bind="props">
            <v-list-item-title data-test="action-item">
              <v-icon>{{ icon }}</v-icon>
              {{ capitalizeText(action) }}
            </v-list-item-title>
          </span>
        </template>
        <span data-test="tooltip-text"> You don't have this kind of authorization. </span>
      </v-tooltip>
    </v-list-item>
    <BaseDialog v-model="showDialog" @click:outside="close" data-test="device-action-dialog">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-5 bg-primary">
          {{ capitalizeText(variant) }} {{ capitalizeText(action) }}
        </v-card-title>
        <v-divider />
        <v-container>
          <v-alert
            v-if="isBillingActive"
            type="warning"
            text="Accepted devices in ShellHub become active in your account and are billed for the entire billing period." />
          <v-card-text class="mt-4 mb-0 pb-1">
            <p class="mb-2"> Do you want to {{ action }} this {{ variant }}? </p>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn variant="text" @click="close()" data-test="close-btn"> Close </v-btn>
            <v-btn variant="text" @click="handleClick()" data-test="action-btn"> {{ action }} </v-btn>
          </v-card-actions>
        </v-container>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { AxiosError } from "axios";
import { authorizer, actions } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { capitalizeText } from "@/utils/string";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";
import useBillingStore from "@/store/modules/billing";
import useDevicesStore from "@/store/modules/devices";
import useNotificationsStore from "@/store/modules/notifications";

interface DeviceActionButtonProps {
  name?: string;
  uid: string;
  isInNotification?: boolean;
  action?: "accept" | "reject" | "remove";
  variant: string;
}

const props = withDefaults(defineProps<DeviceActionButtonProps>(), {
  name: "Device",
  isInNotification: false,
  action: "accept",
});

const emit = defineEmits(["update"]);
const authStore = useAuthStore();
const billingStore = useBillingStore();
const devicesStore = useDevicesStore();
const { fetchNotifications } = useNotificationsStore();
const snackbar = useSnackbar();
const isBillingActive = computed(() => billingStore.isActive);
const icon = {
  accept: "mdi-check",
  reject: "mdi-close",
  remove: "mdi-delete",
}[props.action];
const hasAuthorization = computed(() => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.device[props.action]);
});

const showDialog = ref(false);

const close = () => {
  showDialog.value = false;
  emit("update", false);
};

const refreshDevices = async () => {
  try {
    await fetchNotifications();
    emit("update");
    close();
  } catch (error: unknown) {
    snackbar.showError("Failed to refresh devices.");
    handleError(error);
  }
};

const removeDevice = async () => {
  try {
    await devicesStore.removeDevice(props.uid);
  } catch (error: unknown) {
    snackbar.showError("Failed to remove device.");
    handleError(error);
  }
};

const rejectDevice = async () => {
  try {
    await devicesStore.rejectDevice(props.uid);
  } catch (error: unknown) {
    snackbar.showError("Failed to reject device.");
    handleError(error);
  }
};

const acceptDevice = async () => {
  try {
    await devicesStore.acceptDevice(props.uid);
  } catch (error: unknown) {
    const axiosError = error as AxiosError;
    switch (axiosError.response?.status) {
      case 402:
        billingStore.showBillingWarning = true;
        snackbar.showError("Couldn't accept the device. Check your billing status and try again.");
        break;
      case 403:
        snackbar.showError("You reached the maximum amount of accepted devices in this namespace.");
        break;
      case 409:
        devicesStore.duplicatedDeviceName = props.name;
        snackbar.showError("A device with that name already exists in the namespace. Rename it and try again.");
        break;
      default:
        snackbar.showError("Failed to accept device.");
        handleError(error);
    }
  }
};

const handleClick = async () => {
  if (hasAuthorization.value) {
    const currentDeviceAction = {
      accept: acceptDevice,
      reject: rejectDevice,
      remove: removeDevice,
    }[props.action];

    await currentDeviceAction();
    await refreshDevices();
  } else {
    snackbar.showError("You don't have this kind of authorization.");
  }
};

defineExpose({ showDialog, hasAuthorization });
</script>

<style scoped>
p {
  font-size: 1rem;
}
</style>
