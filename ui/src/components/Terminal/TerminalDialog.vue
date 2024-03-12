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
      v-model="showTerminal"
      max-width="1024px"
      min-width="55vw"
      @click:outside="close"
    >
      <v-card data-test="terminal-dialog" class="bg-v-theme-surface">
        <v-card-title
          class="text-h5 pa-4 bg-primary d-flex align-center justify-center"
        >
          Terminal
          <v-spacer />
          <v-icon @click="close()" class="bg-primary" size="24">mdi-close</v-icon>
        </v-card-title>

        <div class="mt-2" v-if="showLoginForm">
          <v-tabs align-tabs="center" color="primary" v-model="tabActive">
            <v-tab value="Password" @click="resetFieldValidation">Password</v-tab>
            <v-tab
              value="PublicKey"
              @click="resetFieldValidation"
            >Private Key</v-tab
            >
          </v-tabs>

          <v-card-text>
            <v-window v-model="tabActive">
              <v-window-item value="Password">
                <v-form lazy-validation @submit.prevent="connectWithPassword()">
                  <v-text-field
                    v-model="username"
                    :error-messages="usernameError"
                    label="Username"
                    autofocus
                    variant="underlined"
                    :validate-on-blur="true"
                    data-test="username-field"
                  />

                  <v-text-field
                    color="primary"
                    :append-inner-icon="
                      showPassword ? 'mdi-eye' : 'mdi-eye-off'
                    "
                    v-model="password"
                    :error-messages="passwordError"
                    label="Password"
                    required
                    variant="underlined"
                    data-test="password-text"
                    :type="showPassword ? 'text' : 'password'"
                    @click:append-inner="showPassword = !showPassword"
                  />

                  <v-card-actions>
                    <v-spacer />
                    <v-btn
                      type="submit"
                      color="primary"
                      class="mt-4"
                      variant="flat"
                      data-test="connect2-btn"
                    >
                      Connect
                    </v-btn>
                  </v-card-actions>
                </v-form>
              </v-window-item>

              <v-window-item value="PublicKey">
                <v-form
                  lazy-validation
                  @submit.prevent="connectWithPrivateKey()"
                >
                  <v-text-field
                    v-model="username"
                    :error-messages="usernameError"
                    label="Username"
                    autofocus
                    variant="underlined"
                    :validate-on-blur="true"
                    data-test="username-field"
                  />

                  <v-select
                    v-model="privateKey"
                    :items="nameOfPrivateKeys"
                    item-text="name"
                    item-value="data"
                    variant="underlined"
                    label="Private Keys"
                    data-test="privatekeys-select"
                  />

                  <v-card-actions>
                    <v-spacer />
                    <v-btn
                      type="submit"
                      color="primary"
                      class="mt-4"
                      variant="flat"
                      data-test="connect2-btn"
                    >
                      Connect
                    </v-btn>
                  </v-card-actions>
                </v-form>
              </v-window-item>
            </v-window>
          </v-card-text>
        </div>
      </v-card>
      <v-card-item class="ma-0 pa-0 w-100">
        <div ref="terminal" />
      </v-card-item>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  watch,
  nextTick,
  onUnmounted,
} from "vue";
import { useField } from "vee-validate";
import "xterm/css/xterm.css";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import * as yup from "yup";
import axios from "axios";
import { useStore } from "../../store";
import {
  createKeyFingerprint,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
  parsePrivateKeySsh,
} from "../../utils/validate";
import { IPrivateKey } from "../../interfaces/IPrivateKey";
import { IParams } from "../../interfaces/IParams";
import { IConnectToTerminal } from "../../interfaces/ITerminal";

const props = defineProps({
  enableConnectButton: {
    type: Boolean,
    required: false,
    default: false,
  },
  enableConsoleIcon: {
    type: Boolean,
    required: false,
    default: false,
  },
  uid: {
    type: String,
    required: true,
  },
  online: {
    type: Boolean,
    required: false,
    default: false,
  },
  show: {
    type: Boolean,
    required: false,
    default: false,
  },
});
const store = useStore();
const tabActive = ref("Password");
const showPassword = ref(false);
const showLoginForm = ref(true);
const privateKey = ref("");
const xterm = ref<(Terminal)>({} as Terminal);
const ws = ref<WebSocket>({} as WebSocket);
const fitAddon = ref<FitAddon>({} as FitAddon);
const terminal = ref<HTMLElement>({} as HTMLElement);
const uid = computed(() => props.uid);
const showTerminal = ref(store.getters["modal/terminal"] === uid.value);

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

const getListPrivateKeys = computed(() => store.getters["privateKey/list"]);

const nameOfPrivateKeys = computed(() => {
  const list = getListPrivateKeys.value;
  return list.map((item: IPrivateKey) => item.name);
});

watch(showTerminal, (value) => {
  if (!value) {
    if (ws.value) ws.value.close();
    if (xterm.value) {
      xterm.value.dispose();
    }
  } else {
    showLoginForm.value = true;
  }
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
    device: props.uid,
    username: username.value,
    ...params,
  });

  const { token } = response.data;

  showLoginForm.value = false;
  nextTick(() => fitAddon.value.fit());

  if (!xterm.value.element) {
    xterm.value.open(terminal.value);
  }

  fitAddon.value.fit();
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

      enum MessageKind {
        Input = 1,
        Resize,
      }

      interface Message {
        kind: MessageKind;
        data: unknown;
      }

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
        xterm.value.writeln("Connection ended");
      };
};

const open = () => {
  showTerminal.value = true;
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

  store.dispatch("modal/toggleTerminal", props.uid);

  if (xterm.value.element) {
    xterm.value.reset();
  }
};

const resetFieldValidation = () => {
  resetUsername();
  resetPassword();
};

const connectWithPassword = () => {
  connect({ password: password.value });
};

const findPrivateKeyByName = (name: string) => {
  const list = getListPrivateKeys.value;
  return list.find((item: IPrivateKey) => item.name === name);
};

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

const close = () => {
  showTerminal.value = false;
  store.dispatch("modal/toggleTerminal", "");
  resetFieldValidation();
};

onUnmounted(() => {
  close();
});

defineExpose({ open });
</script>

<!-- <style lang="scss" scoped>
.xterm-helper {
  background: #0f1526;
  width: 105%;
}
</style> -->
