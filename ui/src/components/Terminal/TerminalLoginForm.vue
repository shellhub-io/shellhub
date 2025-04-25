<template>
  <v-form @submit.prevent="submitForm" class="mt-2 pa-5 d-flex flex-column ga-4">
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
      :items="[TerminalAuthMethods.Password, TerminalAuthMethods.PrivateKey]"
      label="Authentication method"
      data-test="auth-method-select"
    />

    <v-select
      v-model="selectedPrivateKeyName"
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

    <v-card-actions class="mt-4 d-flex justify-end">
      <v-btn
        @click="emit('close')"
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
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import { TerminalAuthMethods } from "@/interfaces/ITerminal";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { useStore } from "@/store";

interface FormData {
  username: string;
  password: string;
  authenticationMethod: TerminalAuthMethods;
  privateKey?: string;
}

const emit = defineEmits<{
  submit: [formData: FormData];
  close: [];
}>();

const authenticationMethod = ref(TerminalAuthMethods.Password);
const showPassword = ref(false);
const privateKeys: Array<IPrivateKey> = useStore().getters["privateKey/list"];
const selectedPrivateKeyName = computed(() => (
  authenticationMethod.value === TerminalAuthMethods.PrivateKey ? privateKeys[0]?.name || "" : undefined
));
const privateKeysNames = privateKeys.map((item: IPrivateKey) => item.name);

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

const submitForm = () => {
  if (usernameError.value || passwordError.value) {
    return;
  }
  if (authenticationMethod.value === TerminalAuthMethods.PrivateKey && !selectedPrivateKeyName.value) {
    return;
  }

  const selectedPrivateKey = privateKeys.find((item: IPrivateKey) => item.name === selectedPrivateKeyName.value);
  const formData = {
    username: username.value,
    password: password.value,
    authenticationMethod: authenticationMethod.value,
    privateKey: selectedPrivateKey?.data,
  };

  emit("submit", formData);
};

defineExpose({ authenticationMethod });
</script>
