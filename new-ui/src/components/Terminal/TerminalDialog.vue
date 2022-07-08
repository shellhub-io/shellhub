<template>
  <template v-if="enableConnectButton">
    <v-btn
      :disabled="!online"
      :color="online ? 'success' : 'normal'"
      variant="outlined"
      density="comfortable"
      data-test="connect-btn"
      @click="open()"
    >
      Connect
    </v-btn>
  </template>
  <template v-else>
    <span>
      <v-icon left data-test="console-icon"> mdi-console </v-icon>
    </span>

    <span>
      <v-list-item-title class="ml-2" data-test="console-item">
        Console
      </v-list-item-title>
    </span>
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
          <v-tab value="PublicKey" @click="resetFieldValidation"
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
                  :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
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
              <v-form lazy-validation @submit.prevent="connectWithPrivateKey()">
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
      <div ref="terminal" class="mt-n6 xterm-helper" />
    </v-card-item>
  </v-dialog>
</template>

<script lang="ts">
import { useStore } from "../../store";
import { defineComponent, ref, computed, watch, nextTick } from "vue";
import { useField } from "vee-validate";
import { Terminal } from "xterm";
import { AttachAddon } from "xterm-addon-attach";
import { FitAddon } from "xterm-addon-fit";

import * as yup from "yup";
import { parsePrivateKey } from "sshpk";
import {
  createKeyFingerprint,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
  parsePrivateKeySsh,
} from "../../utils/validate";
import axios from "axios";

export default defineComponent({
  inheritAttrs: false,
  props: {
    enableConnectButton: {
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
  },
  setup(props, ctx) {
    const store = useStore();
    const tabActive = ref("Password");
    const showPassword = ref(false);
    const showLoginForm = ref(true);
    const privateKey = ref("");
    const xterm = ref<any>(null);
    const ws = ref<any>(null);
    const fitAddon = ref<any>(null);
    const terminal = ref<any>(null);
    const attachAddon = ref<any>(null);

    const showTerminal = ref(store.getters["modal/terminal"] === props.uid);

    const {
      value: username,
      errorMessage: usernameError,
      setErrors: setUsernameError,
      resetField: resetUsername,
    } = useField<string>("username", yup.string().required(), {
      initialValue: "",
    });

    const {
      value: password,
      errorMessage: passwordError,
      setErrors: setPasswordError,
      resetField: resetPassword,
    } = useField<string>("password", yup.string().required(), {
      initialValue: "",
    });

    const webTermDimensions = computed(() => {
      return {
        cols: xterm.value.cols,
        rows: xterm.value.rows,
      };
    });

    const getListPrivateKeys = computed(() => store.getters["privateKey/list"]);

    const nameOfPrivateKeys = computed(() => {
      const list = getListPrivateKeys.value;
      return list.map((item: any) => item.name);
    });

    watch(showTerminal, (value) => {
      if (!value) {
        if (ws.value) ws.value.close();
        if (xterm.value) {
          xterm.value = null;
        }
      } else {
        showLoginForm.value = true;
      }
    });

    const open = () => {
      showTerminal.value = true;
      privateKey.value = "";

      xterm.value = new Terminal({
        cursorBlink: true,
        fontFamily: "monospace",
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

    const encodeURLParams = (params: any) => {
      return Object.entries(params)
        .map(([key, value]) => `${key}=${value}`)
        .join("&");
    };

    const findPrivateKeyByName = (name: string) => {
      const list = getListPrivateKeys.value;
      return list.find((item: any) => item.name === name);
    };

    const connectWithPrivateKey = async () => {
      const privateKeyData = findPrivateKeyByName(privateKey.value);
      const pk = parsePrivateKeySsh(privateKeyData.data);
      let signature;

      if (pk.type === "ed25519") {
        const signer = createSignerPrivateKey(pk, username.value);
        signature = signer;
      } else {
        signature = await createSignatureOfPrivateKey(
          privateKeyData.data,
          username.value
        );
      }
      const fingerprint = await createKeyFingerprint(privateKeyData.data);
      connect({ fingerprint, signature });
    };

    const connect = async (params: any) => {
      if (params.passwd && !username.value && !password.value) {
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

      ws.value = new WebSocket(
        `${protocolConnectionURL}://${
          window.location.host
        }/ws/ssh?${encodeURLParams(wsInfo)}`
      );

      ws.value.onopen = () => {
        attachAddon.value = new AttachAddon(ws.value);
        xterm.value.loadAddon(attachAddon.value);
      };

      ws.value.onclose = () => {
        if (attachAddon.value) {
          attachAddon.value = null;
          // attachAddon.value.dispose();
        }
      };
    };

    const close = () => {
      showTerminal.value = false;
      store.dispatch("modal/toggleTerminal", "");
      resetFieldValidation();
    };

    return {
      showTerminal,
      getListPrivateKeys,
      open,
      resetFieldValidation,
      tabActive,
      username,
      usernameError,
      password,
      showPassword,
      passwordError,
      privateKey,
      connectWithPassword,
      connectWithPrivateKey,
      nameOfPrivateKeys,
      close,
      showLoginForm,
      terminal,
    };
  },
});
</script>

<style lang="scss" scoped>
.xterm-helper {
  background: #0f1526;
  width: 105%;
}
</style>
