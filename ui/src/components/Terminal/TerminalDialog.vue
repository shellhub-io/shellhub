<template>
  <div>
    <template v-if="enableConnectButton">
      <v-btn
        :disabled="!online"
        :color="online ? 'success' : 'normal'"
        variant="outlined"
        density="comfortable"
        data-test="connect-btn"
        @click="showTerminal = true"
      >
        {{ online ? "Connect" : "Offline" }}
      </v-btn>
    </template>
    <v-dialog
      v-model="showTerminal"
      :fullscreen="$vuetify.display.smAndDown"
      :max-width="!$vuetify.display.smAndDown ? $vuetify.display.thresholds.sm : null"
      @click:outside="showTerminal = !showTerminal"
    >
      <v-card data-test="terminal-card" class="bg-v-theme-surface">
        <v-card-title
          class="text-h5 pa-4 bg-primary d-flex align-center justify-center"
        >
          Terminal
          <v-spacer />
          <v-icon @click="showTerminal = false" data-test="close-btn" class="bg-primary" size="24">mdi-close</v-icon>
        </v-card-title>

        <div class="mt-2" v-if="showLoginForm">
          <v-tabs align-tabs="center" color="primary" v-model="tabActive">
            <v-tab value="Password" block data-test="password-tab" @click="resetFieldValidation">Password</v-tab>
            <v-tab
              value="PrivateKey"
              @click="resetFieldValidation"
              block
              data-test="private-key-tab"
            >Private Key</v-tab
            >
          </v-tabs>

          <v-card-text>
            <v-window v-model="tabActive">
              <v-window-item value="Password">
                <v-form lazy-validation @submit.prevent="connectWithPassword()">
                  <v-container>
                    <v-row>
                      <v-col>
                        <v-text-field
                          v-model="username"
                          :error-messages="usernameError"
                          label="Username"
                          autofocus
                          variant="underlined"
                          hint="Enter an existing user on the device"
                          persistent-hint
                          persistent-placeholder
                          :validate-on-blur="true"
                          data-test="username-field"
                        />
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col>
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
                          hint="Enter a valid password for the user on the device"
                          persistent-hint
                          persistent-placeholder
                          data-test="password-field"
                          :type="showPassword ? 'text' : 'password'"
                          @click:append-inner="showPassword = !showPassword"
                        />
                      </v-col>
                    </v-row>
                  </v-container>

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

              <v-window-item value="PrivateKey">
                <v-form
                  lazy-validation
                  @submit.prevent="connectWithPrivateKey()">
                  <v-container>
                    <v-row>
                      <v-col>
                        <v-text-field
                          v-model="username"
                          :error-messages="usernameError"
                          label="Username"
                          autofocus
                          variant="underlined"
                          hint="Enter an existing user on the device"
                          persistent-hint
                          persistent-placeholder
                          :validate-on-blur="true"
                          data-test="username-field-pk"
                        />
                      </v-col>
                    </v-row>
                    <v-row>
                      <v-col>
                        <v-select
                          v-model="privateKey"
                          :items="nameOfPrivateKeys"
                          item-text="name"
                          item-value="data"
                          variant="underlined"
                          label="Private Key"
                          hint="Select a private key file for authentication"
                          persistent-hint
                          data-test="privatekeys-select"
                        />
                      </v-col>
                    </v-row>
                  </v-container>

                  <v-card-actions>
                    <v-spacer />
                    <v-btn
                      type="submit"
                      color="primary"
                      class="mt-4"
                      variant="flat"
                      data-test="connect2-btn-pk"
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
} from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../../store";
import {
  createKeyFingerprint,
  createSignatureOfPrivateKey,
  createSignerPrivateKey,
  parsePrivateKeySsh,
} from "../../utils/validate";
import { IPrivateKey } from "../../interfaces/IPrivateKey";
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

const getListPrivateKeys = computed(() => store.getters["privateKey/list"]);

const nameOfPrivateKeys = computed(() => {
  const list = getListPrivateKeys.value;
  return list.map((item: IPrivateKey) => item.name);
});

const connect = async (params: IConnectToTerminal) => {
  if (params.password && !username.value && !password.value) {
    return;
  }

  if (params.signature && !username.value && !privateKey.value) {
    return;
  }

  await store.dispatch("terminals/fetch", params);
};

const resetFieldValidation = () => {
  resetUsername();
  resetPassword();
};

const connectWithPassword = () => {
  connect({ username: username.value, device: props.uid, password: password.value });
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
  connect({ username: username.value, device: props.uid, fingerprint, signature });
};

</script>
