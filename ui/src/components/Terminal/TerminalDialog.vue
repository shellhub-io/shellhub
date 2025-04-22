<template>
  <div>
    <template v-if="enableConnectButton">
      <v-btn
        :disabled="!online"
        :color="online ? 'success' : 'normal'"
        variant="outlined"
        density="comfortable"
        data-test="connect-btn"
        @click="open()"
      >
        {{ online ? "Connect" : "Offline" }}
      </v-btn>
    </template>

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

        <div class="ma-0 pa-0 w-100 fill-height position-relative" v-if="!showLoginForm">
          <div ref="terminal" class="terminal" />
        </div>

        <div class="mt-2" v-if="showLoginForm">
          <v-form @submit.prevent="submitForm" class="pa-5 d-flex flex-column ga-4">
            <v-text-field
              v-model="username"
              :error-messages="usernameError"
              label="Username"
              autofocus
              hint="Enter an existing user on the device"
              persistent-hint
              persistent-placeholder
              :validate-on-blur="true"
              data-test="username-field"
            />

            <v-select
              class="mt-2"
              v-model="authenticationMethod"
              :items="[AuthMethods.Password, AuthMethods.PrivateKey]"
              label="Authentication method"
              data-test="auth-method-select"
            />

            <v-select
              v-model="privateKey"
              v-if="authenticationMethod === AuthMethods.PrivateKey"
              :items="privateKeysNames"
              item-text="name"
              item-value="data"
              label="Private Key"
              hint="Select a private key file for authentication"
              persistent-hint
              data-test="private-keys-select"
            />

            <v-text-field
              color="primary"
              :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
              v-model="password"
              v-if="authenticationMethod === AuthMethods.Password"
              :error-messages="passwordError"
              label="Password"
              required
              hint="Enter a valid password for the user on the device"
              persistent-hint
              persistent-placeholder
              data-test="password-field"
              :type="showPassword ? 'text' : 'password'"
              @click:append-inner="showPassword = !showPassword"
            />

            <v-card-actions class="mt-4 d-flex justify-end">
              <v-btn
                @click="close"
                data-test="cancel-btn"
              >
                Cancel
              </v-btn>
              <v-btn
                type="submit"
                color="primary"
                data-test="submit-btn"
              >
                Connect
              </v-btn>
            </v-card-actions>
          </v-form>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  nextTick,
  watch,
} from "vue";
import { useField } from "vee-validate";
import "xterm/css/xterm.css";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import * as yup from "yup";
import axios from "axios";
import { useEventListener } from "@vueuse/core";
import { useRoute } from "vue-router";
import { useDisplay } from "vuetify";
import { useStore } from "@/store";
import {
  createKeyFingerprint,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
  parsePrivateKeySsh,
} from "@/utils/validate";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { IParams } from "@/interfaces/IParams";
import { IConnectToTerminal, TerminalAuthMethods } from "@/interfaces/ITerminal";
import TerminalLoginForm from "./TerminalLoginForm.vue";

enum MessageKind {
  Input = 1,
  Resize,
}

interface Message {
  kind: MessageKind;
  data: unknown;
}

const { uid } = defineProps({
  enableConnectButton: {
    type: Boolean,
    default: false,
  },
  enableConsoleIcon: {
    type: Boolean,
    default: false,
  },
  uid: {
    type: String,
    required: true,
  },
  online: {
    type: Boolean,
    default: false,
  },
});
const store = useStore();
const route = useRoute();
const authenticationMethod = ref(AuthMethods.Password);
const showPassword = ref(false);
const showLoginForm = ref(true);
const privateKey = ref("");
const xterm = ref<(Terminal)>({} as Terminal);
const ws = ref<WebSocket>({} as WebSocket);
const fitAddon = ref<FitAddon>({} as FitAddon);
const terminal = ref<HTMLElement>({} as HTMLElement);
const showDialog = ref(store.getters["modal/terminal"] === uid);
const { smAndDown, thresholds } = useDisplay();

const {
  value: username,
  errorMessage: usernameError,
  resetField: resetUsername,
} = useField<string>("username", yup.string().required(), {
  initialValue: "",
});

const {
  value: password,
  errorMessage: passwordError,
  resetField: resetPassword,
} = useField<string>("password", yup.string().required(), {
  initialValue: "",
});

const webTermDimensions = computed(() => ({
  cols: xterm.value.cols,
  rows: xterm.value.rows,
}));

const privateKeys = store.getters["privateKey/list"];

const privateKeysNames = privateKeys.map((item: IPrivateKey) => item.name);

useEventListener(window, "resize", () => {
  nextTick(() => {
    fitAddon.value.fit();
  });
});

watch(showDialog, (value) => {
  if (!value) showLoginForm.value = true;
});

const encodeURLParams = (params: IParams) => Object.entries(params)
  .map(([key, value]) => `${key}=${value}`)
  .join("&");

const connect = async (params: IConnectToTerminal) => {
  if (params.password && !username.value && !password.value) {
    return;
  }

  if (params.signature && !username.value && !privateKey.value) {
    return;
  }

  const response = await axios.post("/ws/ssh", {
    device: uid,
    username: username.value,
    ...params,
  });

  const { token } = response.data;

  showLoginForm.value = false;
  nextTick(() => {
    if (!xterm.value.element) {
      xterm.value.open(terminal.value);
    }

    xterm.value.focus();

    let protocolConnectionURL = "";

    if (window.location.protocol === "http:") {
      protocolConnectionURL = "ws";
    } else {
      protocolConnectionURL = "wss";
    }

    const wsInfo = { token, ...webTermDimensions.value };

    const enc = new TextEncoder();
    ws.value = new WebSocket(
      `${protocolConnectionURL}://${
        window.location.host
      }/ws/ssh?${encodeURLParams(wsInfo)}`,
    );

    ws.value.onopen = () => {
      fitAddon.value.fit();
    };

    ws.value.onmessage = (ev) => {
      xterm.value.write(ev.data);
    };

    xterm.value.onData((data) => {
      const message: Message = {
        kind: MessageKind.Input,
        data: [...enc.encode(data)],
      };

      ws.value.send(JSON.stringify(message));
    });

    xterm.value.onResize((data) => {
      const message: Message = {
        kind: MessageKind.Resize,
        data: { cols: data.cols, rows: data.rows },
      };

      ws.value.send(JSON.stringify(message));
    });

    ws.value.onclose = () => {
      xterm.value.write("\r\nConnection ended");
    };
  });
};

const open = () => {
  showDialog.value = true;
  privateKey.value = "";

  xterm.value = new Terminal({
    cursorBlink: true,
    fontFamily: "monospace",
    theme: {
      background: "#0f1526",
    },
  });

  fitAddon.value = new FitAddon();
  xterm.value.loadAddon(fitAddon.value);

  store.dispatch("modal/toggleTerminal", uid);

  if (xterm.value.element) {
    xterm.value.reset();
  }
};

watch(() => route.path, (path) => {
  if (path === `/devices/${uid}/terminal`) {
    open();
  }
}, { immediate: true });

const resetFieldValidation = () => {
  resetUsername();
  resetPassword();
};

const connectWithPassword = () => {
  connect({ password: password.value });
};

const findPrivateKeyByName = (name: string) => privateKeys.find((item: IPrivateKey) => item.name === name);

const connectWithPrivateKey = async () => {
  const privateKeyData = findPrivateKeyByName(privateKey.value);
  const pk = parsePrivateKeySsh(privateKeyData.data);
  let signature;

  if (pk.type === "ed25519") {
    const signer = createSignerPrivateKey(pk, username.value);
    signature = signer;
  } else {
    signature = decodeURIComponent(await createSignatureOfPrivateKey(
      privateKeyData.data,
      username.value,
    ));
  }
  const fingerprint = await createKeyFingerprint(privateKeyData.data);
  connect({ fingerprint, signature });
};

const submitForm = () => {
  if (authenticationMethod.value === AuthMethods.Password) {
    connectWithPassword();
  } else if (authenticationMethod.value === AuthMethods.PrivateKey) {
    connectWithPrivateKey();
  }
};

const close = () => {
  if (ws.value.OPEN) {
    ws.value.close();
  }
  showDialog.value = false;
  xterm.value.clear();
  resetFieldValidation();
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

defineExpose({ open, showDialog, showLoginForm, encodeURLParams, submitForm, connect, privateKey, xterm, fitAddon, ws, close });
</script>

<style lang="scss" scoped>
.terminal {
  position: absolute;
  top: 0px;
  bottom: 0px;
  left: 0;
  right:0;
  margin-right: 0px;
}
</style>
