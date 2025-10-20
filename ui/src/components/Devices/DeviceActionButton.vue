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
      <v-tooltip location="bottom" class="text-center" :disabled="canPerformDeviceAction">
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
    <MessageDialog
      v-model="showDialog"
      @close="close"
      @confirm="handleClick"
      @cancel="close"
      :title="`${capitalizeText(variant)} ${capitalizeText(action)}`"
      :description="`Do you want to ${action} this ${variant}?`"
      icon="mdi-help-circle"
      :icon-color="action === 'accept' ? 'primary' : action === 'reject' ? 'warning' : 'error'"
      :confirm-text="capitalizeText(action)"
      :confirm-color="action === 'accept' ? 'primary' : action === 'reject' ? 'warning' : 'error'"
      cancel-text="Close"
      confirm-data-test="action-btn"
      cancel-data-test="close-btn"
      data-test="device-action-dialog"
    >
      <v-alert
        v-if="isBillingActive"
        type="warning"
        class="mx-4 mb-4"
        text="Accepted devices in ShellHub become active in your account and are billed for the entire billing period."
      />
    </MessageDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { AxiosError } from "axios";
import hasPermission from "@/utils/permission";
import { capitalizeText } from "@/utils/string";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
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
const canPerformDeviceAction = hasPermission(`device:${props.action}`);

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
  if (canPerformDeviceAction) {
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

defineExpose({ showDialog, canPerformDeviceAction });
</script>

<style scoped>
p {
  font-size: 1rem;
}
</style>
