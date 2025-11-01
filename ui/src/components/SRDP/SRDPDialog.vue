<template>
  <SRDPLoginForm
    v-if="showLoginForm"
    v-model="showLoginDialog"
    :device="deviceUid"
    @submit="handleSubmit"
    @close="close"
  />

  <SRDP
    v-else
    v-model="showSRDPDialog"
    :key="SRDPKey"
    :device="device"
    :username="username"
    :password="password"
    :display="display"
    :device-name
    @close="close"
  />
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import { useRoute, onBeforeRouteLeave } from "vue-router";
import { SRDPLoginFormData } from "@/interfaces/ISRDP";

// Components used in this dialog
import SRDPLoginForm from "./SRDPLoginForm.vue";
import SRDP from "./SRDP.vue";

const { deviceUid, deviceName } = defineProps<{
  deviceUid: string;
  deviceName: string;
}>();

const route = useRoute(); // current route
const showLoginForm = ref(true); // controls whether login or SRDP viewer is shown
const SRDPKey = ref(0); // key to force re-render
const showDialog = defineModel<boolean>({ required: true }); // controls visibility of dialog

const showLoginDialog = computed({
  get: () => showDialog.value && showLoginForm.value,
  set: (value) => { showDialog.value = value; },
});

const showSRDPDialog = computed({
  get: () => showDialog.value && !showLoginForm.value,
  set: (value) => { showDialog.value = value; },
});

// Connection credentials - passed directly to WASM
const device = ref("");
const username = ref("");
const password = ref("");
const display = ref("");

// Triggered when the user submits login credentials
// No need to authenticate - WASM handles it internally
const handleSubmit = async (params: SRDPLoginFormData) => {
  // Use device from props, and username/password/display from form
  device.value = deviceUid;
  username.value = params.username;
  password.value = params.password;
  display.value = params.display || "";
  showLoginForm.value = false;
};

// Reset state and close the dialog
const close = () => {
  showDialog.value = false;
  showLoginForm.value = true;
  device.value = "";
  username.value = "";
  password.value = "";
  display.value = "";
  SRDPKey.value++; // trigger remount
};

// Close SRDP viewer when navigating away
onBeforeRouteLeave(() => {
  if (showDialog.value) {
    close();
    return false; // Prevent navigation if dialog is open
  }

  return true;
});

// Auto-open SRDP viewer when navigating to specific device route
watch(
  () => route.path,
  (path) => {
    if (path === `/devices/${deviceUid}/desktop`) showDialog.value = true;
  },
  { immediate: true },
);

// Expose for test or parent interaction
defineExpose({
  device,
  username,
  handleSubmit,
  showDialog,
  showLoginForm,
  close,
});
</script>
