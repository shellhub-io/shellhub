<template>
  <v-dialog
    v-model="showDialog"
    :fullscreen="!showLoginForm || smAndDown"
    :max-width="smAndDown || !showLoginForm ? undefined : thresholds.sm"
    @click:outside="close"
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
        :token
      />
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import axios from "axios";
import { useEventListener } from "@vueuse/core";
import { useRoute } from "vue-router";
import { useDisplay } from "vuetify";
import {
  createKeyFingerprint,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
  parsePrivateKeySsh,
} from "@/utils/validate";
import { IConnectToTerminal, TerminalAuthMethods } from "@/interfaces/ITerminal";
import TerminalLoginForm from "./TerminalLoginForm.vue";
import Terminal from "./Terminal.vue";

const { deviceUid } = defineProps<{
  deviceUid: string;
}>();

const route = useRoute();
const showLoginForm = ref(true);
const showDialog = defineModel<boolean>();
const { smAndDown, thresholds } = useDisplay();
const token = ref("");

watch(showDialog, (value) => {
  if (!value) showLoginForm.value = true;
});

const connect = async (params: IConnectToTerminal) => {
  const response = await axios.post("/ws/ssh", {
    device: deviceUid,
    ...params,
  });

  token.value = response.data.token;

  showLoginForm.value = false;
};

const connectWithPrivateKey = async (params: IConnectToTerminal) => {
  const { username, privateKey } = params;
  const parsedPrivateKey = parsePrivateKeySsh(privateKey);
  const fingerprint = await createKeyFingerprint(privateKey);

  let signature;
  if (parsedPrivateKey.type === "ed25519") {
    const signer = createSignerPrivateKey(parsedPrivateKey, username);
    signature = signer;
  } else {
    signature = decodeURIComponent(await createSignatureOfPrivateKey(
      parsedPrivateKey,
      username,
    ));
  }

  connect({ username, fingerprint, signature });
};

const handleSubmit = (params) => {
  if (params.authenticationMethod === TerminalAuthMethods.Password) {
    connect(params);
  } else connectWithPrivateKey(params);
};

const close = () => {
  showDialog.value = false;
  store.dispatch("modal/toggleTerminal", "");
};

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

useEventListener("keyup", handleEscKey);

watch(() => route.path, (path) => {
  if (path === `/devices/${deviceUid}/terminal`) showDialog.value = true;
}, { immediate: true });

defineExpose({ showDialog, showLoginForm, handleSubmit, connect, close });
</script>
