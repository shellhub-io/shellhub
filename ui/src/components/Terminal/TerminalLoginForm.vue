<template>
  <FormDialog
    v-bind="$attrs"
    v-model="showDialog"
    title="Connect to Device"
    icon="mdi-console"
    confirm-text="Connect"
    :confirm-loading="isConnecting"
    :confirm-disabled="!isFormValid"
    cancel-text="Cancel"
    confirm-data-test="submit-btn"
    cancel-data-test="cancel-btn"
    @close="handleClose"
    @confirm="submitForm"
    @cancel="handleClose"
  >
    <v-card-text
      class="d-flex flex-column ga-4 pa-6"
      data-test="terminal-login-form"
    >
      <v-text-field
        v-model="username"
        :error-messages="usernameError"
        label="Username"
        autofocus
        hint="Enter an existing user on the device"
        persistent-hint
        persistent-placeholder
        :validate-on-blur="true"
        autocomplete="username"
        data-test="username-field"
      />

      <v-select
        v-model="authenticationMethod"
        :items="[TerminalAuthMethods.Password, TerminalAuthMethods.PrivateKey]"
        label="Authentication method"
        data-test="auth-method-select"
        hide-details
        class="mb-2"
        @update:model-value="togglePassphraseField"
      />

      <PrivateKeySelectWithAdd
        v-if="authenticationMethod === TerminalAuthMethods.PrivateKey"
        v-model="selectedPrivateKeyName"
        @key-added="togglePassphraseField"
        @update:model-value="togglePassphraseField"
      />

      <v-text-field
        v-else
        v-model="password"
        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :error-messages="passwordError"
        label="Password"
        required
        hint="Enter a valid password for the user on the device"
        persistent-hint
        persistent-placeholder
        autocomplete="current-password"
        data-test="password-field"
        :type="showPassword ? 'text' : 'password'"
        @click:append-inner="showPassword = !showPassword"
        @keydown.enter.prevent="submitForm"
      />

      <v-text-field
        v-if="showPassphraseField"
        v-model="passphrase"
        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
        :error-messages="passphraseValidationError || passphraseError"
        label="Passphrase"
        required
        hint="Enter the key's passphrase"
        persistent-hint
        persistent-placeholder
        autocomplete="current-password"
        data-test="passphrase-field"
        :type="showPassword ? 'text' : 'password'"
        @click:append-inner="showPassword = !showPassword"
        @keydown.enter.prevent="submitForm"
        @update:model-value="passphraseValidationError = ''"
      />

      <v-alert
        v-if="props.sshid"
        color="primary"
        variant="tonal"
        density="compact"
        class="mt-4"
        data-test="sshid-hint"
        role="status"
        aria-live="polite"
      >
        <div class="d-flex align-center justify-space-between ga-2">
          <div class="text-body-2">
            <v-icon
              icon="mdi-lightbulb-on-outline"
              class="mr-1"
            />
            <strong>Did you know?</strong> You can also connect from your local
            terminal using the SSHID.
          </div>
          <v-btn
            size="small"
            variant="text"
            data-test="show-sshid-examples-btn"
            text="Show me how"
            @click="showTerminalHelper = true"
          />
        </div>
      </v-alert>
    </v-card-text>
  </FormDialog>

  <SSHIDHelper
    v-if="props.sshid"
    v-model="showTerminalHelper"
    :sshid="props.sshid"
  />
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import SSHIDHelper from "./SSHIDHelper.vue";
import PrivateKeySelectWithAdd from "@/components/PrivateKeys/PrivateKeySelectWithAdd.vue";
import { LoginFormData, TerminalAuthMethods } from "@/interfaces/ITerminal";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";
import { isKeyValid } from "@/utils/sshKeys";

const props = defineProps<{
  sshid?: string;
}>();

const emit = defineEmits<{
  submit: [formData: LoginFormData];
  close: [];
}>();

const showDialog = defineModel<boolean>({ required: true });
const isConnecting = defineModel<boolean>("loading", { default: false });
const { privateKeys } = usePrivateKeysStore();
const authenticationMethod = ref(TerminalAuthMethods.Password);
const showPassword = ref(false);
const selectedPrivateKeyName = ref(privateKeys[0]?.name || "");
const showPassphraseField = ref(false);
const showTerminalHelper = ref(false);

const {
  value: username,
  errorMessage: usernameError,
  resetField: resetUsernameField,
} = useField<string>("username", yup.string().required(), {
  initialValue: "",
});

const {
  value: password,
  errorMessage: passwordError,
  resetField: resetPasswordField,
} = useField<string>("password", yup.string().required(), {
  initialValue: "",
});

const {
  value: passphrase,
  errorMessage: passphraseError,
  resetField: resetPassphraseField,
} = useField<string>("passphrase", yup.string().required(), {
  initialValue: "",
});

const passphraseValidationError = ref("");

const getSelectedPrivateKey = () =>
  privateKeys.find(
    (item: IPrivateKey) => item.name === selectedPrivateKeyName.value,
  );

const isFormValid = computed(() => {
  if (usernameError.value) return false;

  if (authenticationMethod.value === TerminalAuthMethods.Password) {
    return !passwordError.value && !!username.value && !!password.value;
  }

  if (authenticationMethod.value === TerminalAuthMethods.PrivateKey) {
    if (!selectedPrivateKeyName.value) return false;
    if (showPassphraseField.value && !passphrase.value) return false;
    return !passphraseError.value && !!username.value;
  }

  return false;
});

const togglePassphraseField = () => {
  if (authenticationMethod.value === TerminalAuthMethods.PrivateKey) {
    const hasPassphrase = getSelectedPrivateKey()?.hasPassphrase || false;
    showPassphraseField.value = hasPassphrase;
  } else showPassphraseField.value = false;
  showPassword.value = false;
  resetPassphraseField();
};

const resetFields = () => {
  resetUsernameField();
  resetPasswordField();
  resetPassphraseField();
  authenticationMethod.value = TerminalAuthMethods.Password;
  selectedPrivateKeyName.value = privateKeys[0]?.name || "";
  showPassphraseField.value = false;
};

const handleClose = () => {
  showDialog.value = false;
  resetFields();
  emit("close");
};

const submitForm = () => {
  if (!isFormValid.value) return;

  const selectedKey = authenticationMethod.value === TerminalAuthMethods.PrivateKey
    ? getSelectedPrivateKey()
    : undefined;

  if (showPassphraseField.value && selectedKey) {
    if (!isKeyValid("private", selectedKey.data, passphrase.value)) {
      passphraseValidationError.value = "Wrong passphrase";
      return;
    }
  }

  const formData: LoginFormData = {
    username: username.value,
    password: password.value,
    authenticationMethod: authenticationMethod.value,
    privateKey: selectedKey?.data,
    passphrase: showPassphraseField.value ? passphrase.value : undefined,
  };

  emit("submit", formData);
  resetFields();
};
</script>
