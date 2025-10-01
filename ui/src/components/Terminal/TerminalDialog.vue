<template>
  <TerminalLoginForm
    v-if="showLoginForm"
    v-model="showLoginDialog"
    v-model:loading="isConnecting"
    @submit="handleSubmit"
    @close="close"
  />

  <Terminal
    v-else
    v-model="showTerminalDialog"
    :key="terminalKey"
    :token="token"
    :private-key="privateKey ?? null"
    :passphrase="passphrase"
    :device-name
    @close="close"
  />
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import axios from "axios";
import { useRoute, onBeforeRouteLeave } from "vue-router";
import {
  IConnectToTerminal,
  LoginFormData,
  TerminalAuthMethods,
} from "@/interfaces/ITerminal";

// Components used in this dialog
import TerminalLoginForm from "./TerminalLoginForm.vue";
import Terminal from "./Terminal.vue";

// Utility to create key fingerprint for private key auth
import { convertToFingerprint } from "@/utils/sshKeys";

const { deviceUid, deviceName } = defineProps<{
  deviceUid: string;
  deviceName: string;
}>();

const route = useRoute(); // current route
const showLoginForm = ref(true); // controls whether login or terminal is shown
const terminalKey = ref(0);
const showDialog = defineModel<boolean>({ required: true }); // controls visibility of dialog
const isConnecting = ref(false);

const showLoginDialog = computed({
  get: () => showDialog.value && showLoginForm.value,
  set: (value) => { showDialog.value = value; },
});

const showTerminalDialog = computed({
  get: () => showDialog.value && !showLoginForm.value,
  set: (value) => { showDialog.value = value; },
});

// Token and private key values for terminal connection
const token = ref("");
const privateKey = ref<LoginFormData["privateKey"]>("");
const passphrase = ref(); // Passphrase for private key if needed

// Connect to terminal via password or key
const connect = async (params: IConnectToTerminal) => {
  isConnecting.value = true;
  try {
    const response = await axios.post("/ws/ssh", {
      device: deviceUid,
      ...params,
    });

    token.value = response.data.token;
    showLoginForm.value = false;
  } finally {
    isConnecting.value = false;
  }
};

// Handles private key-based connection
const connectWithPrivateKey = async (params: IConnectToTerminal) => {
  const { username, privateKey, passphrase } = params;
  const fingerprint = convertToFingerprint(privateKey as string, passphrase);
  await connect({ username, fingerprint });
};

// Triggered when the user submits login credentials
const handleSubmit = async (params: LoginFormData) => {
  if (params.authenticationMethod === TerminalAuthMethods.Password) {
    await connect(params);
    return;
  }

  await connectWithPrivateKey(params);
  privateKey.value = params.privateKey;
  passphrase.value = params.passphrase || undefined;
  showLoginForm.value = false;
};

// Reset state and close the dialog
const close = () => {
  showDialog.value = false;
  showLoginForm.value = true;
  token.value = "";
  privateKey.value = "";
  terminalKey.value++; // trigger remount
};

// Close terminal and log out when navigating away
onBeforeRouteLeave(() => {
  if (showDialog.value) {
    close();
    return false; // Prevent navigation if dialog is open
  }

  return true;
});

// Auto-open terminal when navigating to specific device route
watch(
  () => route.path,
  (path) => {
    if (path === `/devices/${deviceUid}/terminal`) showDialog.value = true;
  },
  { immediate: true },
);

// Expose for test or parent interaction
defineExpose({
  token,
  handleSubmit,
  showDialog,
  showLoginForm,
  close,
});
</script>
