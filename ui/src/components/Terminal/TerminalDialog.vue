<template>
  <BaseDialog
    v-model="showDialog"
    :forceFullscreen="!showLoginForm"
    @click:outside="close"
    @keydown.esc="close"
  >
    <v-card data-test="terminal-card" class="bg-v-theme-surface">
      <v-card-title
        class="text-h5 pa-4 bg-primary d-flex align-center justify-space-between"
      >
        Terminal
        <v-icon v-if="!showLoginForm" @click="close()" data-test="close-terminal-btn" size="24">mdi-close</v-icon>
      </v-card-title>

      <TerminalLoginForm
        v-if="showLoginForm"
        @submit="(params) => handleSubmit(params)"
        @close="close"
      />
      <Terminal
        v-else
        :key="terminalKey"
        :token="token"
        :privateKey="privateKey ?? null"
        :passphrase="passphrase"
      />
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import axios from "axios";
import { useEventListener } from "@vueuse/core";
import { useRoute, onBeforeRouteLeave } from "vue-router";
import {
  IConnectToTerminal,
  LoginFormData,
  TerminalAuthMethods,
} from "@/interfaces/ITerminal";

// Components used in this dialog
import TerminalLoginForm from "./TerminalLoginForm.vue";
import Terminal from "./Terminal.vue";
import BaseDialog from "../BaseDialog.vue";

// Utility to create key fingerprint for private key auth
import { createKeyFingerprint } from "@/utils/validate";

// Props: Device UID to connect the terminal session to
const { deviceUid } = defineProps<{
  deviceUid: string;
}>();

const route = useRoute(); // current route
const showLoginForm = ref(true); // controls whether login or terminal is shown
const terminalKey = ref(0);
const showDialog = defineModel<boolean>({ required: true }); // controls visibility of dialog

// Token and private key values for terminal connection
const token = ref("");
const privateKey = ref<LoginFormData["privateKey"]>("");
const passphrase = ref(); // Passphrase for private key if needed

// Connect to terminal via password or key
const connect = async (params: IConnectToTerminal) => {
  const response = await axios.post("/ws/ssh", {
    device: deviceUid,
    ...params,
  });

  token.value = response.data.token;
  showLoginForm.value = false;
};

// Handles private key-based connection
const connectWithPrivateKey = async (params: IConnectToTerminal) => {
  const { username, privateKey, passphrase } = params;
  const fingerprint = createKeyFingerprint(privateKey, passphrase);
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

// Track timing of ESC presses to close terminal on double ESC
let lastEscPress = 0;
const handleEscKey = (event: KeyboardEvent) => {
  if (event.key === "Escape" && !showLoginForm.value) {
    const currentTime = new Date().getTime();
    if (currentTime - lastEscPress < 400) {
      close();
    }
    lastEscPress = currentTime;
  }
};

// Bind ESC key listener
useEventListener("keyup", handleEscKey);

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
