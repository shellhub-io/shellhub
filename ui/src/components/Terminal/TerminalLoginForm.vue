<template>
  <FormDialog
    v-model="showDialog"
    @close="handleClose"
    @confirm="submitForm"
    @cancel="handleClose"
    title="Terminal Login"
    icon="mdi-login"
    confirm-text="Connect"
    :confirm-loading="isConnecting"
    :confirm-disabled="!isFormValid"
    cancel-text="Cancel"
    confirm-data-test="submit-btn"
    cancel-data-test="cancel-btn"
  >
    <v-card-text class="d-flex flex-column ga-4 pa-6" data-test="terminal-login-form">
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
        v-model="authenticationMethod"
        @update:model-value="togglePassphraseField"
        :items="[TerminalAuthMethods.Password, TerminalAuthMethods.PrivateKey]"
        label="Authentication method"
        data-test="auth-method-select"
        hide-details
        class="mb-2"
      />

      <v-select
        v-model="selectedPrivateKeyName"
        @update:model-value="togglePassphraseField"
        v-if="authenticationMethod === TerminalAuthMethods.PrivateKey"
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
        v-else
        :error-messages="passwordError"
        label="Password"
        required
        hint="Enter a valid password for the user on the device"
        persistent-hint
        persistent-placeholder
        data-test="password-field"
        :type="showPassword ? 'text' : 'password'"
        @click:append-inner="showPassword = !showPassword"
        @keydown.enter.prevent="submitForm"
      />

      <v-text-field
        color="primary"
        v-if="showPassphraseField"
        :append-inner-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
        v-model="passphrase"
        :error-messages="passphraseError"
        label="Passphrase"
        required
        hint="Enter the key's passphrase"
        persistent-hint
        persistent-placeholder
        data-test="passphrase-field"
        :type="showPassword ? 'text' : 'password'"
        @click:append-inner="showPassword = !showPassword"
        @keydown.enter.prevent="submitForm"
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import { LoginFormData, TerminalAuthMethods } from "@/interfaces/ITerminal";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";
import FormDialog from "../FormDialog.vue";

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
const privateKeysNames = privateKeys.map((item: IPrivateKey) => item.name);
const showPassphraseField = ref(false);

const {
  value: username,
  errorMessage: usernameError,
} = useField<string>("username", yup.string().required(), {
  initialValue: "",
});

const {
  value: password,
  errorMessage: passwordError,
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

const getSelectedPrivateKey = () => privateKeys.find((item: IPrivateKey) => item.name === selectedPrivateKeyName.value);

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

const handleClose = () => {
  showDialog.value = false;
  emit("close");
};

const submitForm = () => {
  if (!isFormValid.value) return;

  const privateKey = authenticationMethod.value === TerminalAuthMethods.PrivateKey ? getSelectedPrivateKey()?.data : undefined;

  const formData: LoginFormData = {
    username: username.value,
    password: password.value,
    authenticationMethod: authenticationMethod.value,
    privateKey,
    passphrase: showPassphraseField.value ? passphrase.value : undefined,
  };

  emit("submit", formData);
};

defineExpose({
  authenticationMethod,
  togglePassphraseField,
  isFormValid,
  submitForm,
  username,
  password,
  passphrase,
  selectedPrivateKeyName,
});
</script>
