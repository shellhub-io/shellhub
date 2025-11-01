<template>
  <FormDialog
    v-model="showDialog"
    @close="handleClose"
    @confirm="submitForm"
    @cancel="handleClose"
    title="Remote Desktop Login"
    icon="mdi-monitor-screenshot"
    confirm-text="Connect"
    :confirm-disabled="!isFormValid"
    cancel-text="Cancel"
    confirm-data-test="submit-btn"
    cancel-data-test="cancel-btn"
  >
    <v-card-text class="d-flex flex-column ga-4 pa-6" data-test="SRDP-login-form">
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

      <v-text-field
        v-model="display"
        label="Display"
        hint="Optional: Enter display server (e.g., :0)"
        persistent-hint
        persistent-placeholder
        data-test="display-field"
        placeholder=":0"
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { SRDPLoginFormData } from "@/interfaces/ISRDP";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";

const emit = defineEmits<{
  submit: [formData: SRDPLoginFormData];
  close: [];
}>();

const { device } = defineProps<{
  device: string;
}>();

const showDialog = defineModel<boolean>({ required: true });
const { privateKeys } = usePrivateKeysStore();
const authenticationMethod = ref(TerminalAuthMethods.Password);
const showPassword = ref(false);
const selectedPrivateKeyName = ref(privateKeys[0]?.name || "");
const privateKeysNames = privateKeys.map((item: IPrivateKey) => item.name);
const showPassphraseField = ref(false);
const display = ref("");

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

const resetFields = () => {
  resetUsernameField();
  resetPasswordField();
  resetPassphraseField();
  authenticationMethod.value = TerminalAuthMethods.Password;
  selectedPrivateKeyName.value = privateKeys[0]?.name || "";
  showPassphraseField.value = false;
  display.value = "";
};

const handleClose = () => {
  showDialog.value = false;
  resetFields();
  emit("close");
};

const submitForm = () => {
  if (!isFormValid.value) return;

  const privateKey = authenticationMethod.value === TerminalAuthMethods.PrivateKey ? getSelectedPrivateKey()?.data : undefined;

  const formData: SRDPLoginFormData = {
    device: device.value,
    username: username.value,
    password: password.value,
    privateKey,
    passphrase: showPassphraseField.value ? passphrase.value : undefined,
    display: display.value || undefined,
  };

  emit("submit", formData);
  resetFields();
};

defineExpose({
  authenticationMethod,
  togglePassphraseField,
  isFormValid,
  submitForm,
  device,
  username,
  password,
  passphrase,
  display,
  selectedPrivateKeyName,
});
</script>
