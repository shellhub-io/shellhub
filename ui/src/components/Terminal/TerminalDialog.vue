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

        <TerminalLoginForm
          v-if="showLoginForm"
          v-model:authenticationMethod="authenticationMethod"
          @submit="(params) => submitForm(params)"
          @close="close()"
        />
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
import "xterm/css/xterm.css";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
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
const authenticationMethod = ref(TerminalAuthMethods.Password);
const showLoginForm = ref(true);
const xterm = ref<(Terminal)>({} as Terminal);
const ws = ref<WebSocket>({} as WebSocket);
const fitAddon = ref<FitAddon>({} as FitAddon);
const terminal = ref<HTMLElement>({} as HTMLElement);
const showDialog = ref(store.getters["modal/terminal"] === uid);
const { smAndDown, thresholds } = useDisplay();

const webTermDimensions = computed(() => ({
  cols: xterm.value.cols,
  rows: xterm.value.rows,
}));

watch(showDialog, (value) => {
  if (!value) showLoginForm.value = true;
});

const encodeURLParams = (params: IParams) => Object.entries(params)
  .map(([key, value]) => `${key}=${value}`)
  .join("&");

const connect = async (params: IConnectToTerminal) => {
  const response = await axios.post("/ws/ssh", {
    device: uid,
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

const connectWithPrivateKey = async (params: IConnectToTerminal) => {
  const { username, privateKey } = params;
  const parsedPrivateKey = parsePrivateKeySsh(privateKey);
  const fingerprint = await createKeyFingerprint(parsedPrivateKey);

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

const submitForm = (params) => {
  if (params.authenticationMethod === TerminalAuthMethods.Password) {
    connect(params);
  } else connectWithPrivateKey(params);
};

const close = () => {
  if (ws.value.OPEN) {
    ws.value.close();
  }
  showDialog.value = false;
  xterm.value.clear();
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

useEventListener(window, "resize", () => {
  nextTick(() => {
    fitAddon.value.fit();
  });
});

defineExpose({ open, showDialog, showLoginForm, encodeURLParams, submitForm, connect, xterm, fitAddon, ws, close });
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
