<template>
  <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary" data-test="card-title">
        New Private Key
      </v-card-title>
      <form @submit.prevent="create" class="mt-1">
        <v-card-text>

          <v-alert
            class="text-subtitle-2 mb-2"
            title="ShellHub never stores your private keys."
            text="They stay secure in your browser's local storage and are not shared with ShellHub's servers."
            color="primary"
            density="compact"
            variant="tonal"
            data-test="privacy-policy-alert"
          />
          <v-text-field
            v-model="name"
            :error-messages="nameError"
            label="Key name"
            placeholder="Name used to identify the private key"
            variant="underlined"
            data-test="name-field"
          />

          <v-textarea
            v-model="privateKeyData"
            label="Private key data"
            required
            messages="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
            :error-messages="privateKeyDataError"
            @update:model-value="validatePrivateKeyData"
            variant="underlined"
            data-test="private-key-field"
            rows="5"
          />

          <v-text-field
            v-if="hasPassphrase"
            v-model="passphrase"
            :error-messages="passphraseError"
            label="Passphrase"
            class="mt-4"
            hint="The key is encrypted and needs a passphrase. The passphrase is not stored."
            persistent-hint
            placeholder="Enter passphrase for encrypted key"
            variant="underlined"
            data-test="passphrase-field"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            @click="close"
            data-test="private-key-cancel-btn"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            type="submit"
            data-test="private-key-save-btn"
            :disabled="!!privateKeyDataError || !!nameError || (hasPassphrase && !!passphraseError)"
          >
            Save
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { createKeyFingerprint, parsePrivateKeySsh, validateKey } from "@/utils/validate";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const emit = defineEmits(["update"]);
const privateKeysStore = usePrivateKeysStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });
const hasPassphrase = ref(false);

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
  resetField: resetName,
} = useField<string>("name", yup.string().required("Name is required"), {
  initialValue: "",
});

const {
  value: privateKeyData,
  errorMessage: privateKeyDataError,
  setErrors: setPrivateKeyDataError,
  resetField: resetPrivateKeyData,
} = useField<string>("privateKeyData", yup.string().required("Private key data is required"), {
  initialValue: "",
});

const {
  value: passphrase,
  errorMessage: passphraseError,
  setErrors: setPassphraseError,
  resetField: resetPassphrase,
} = useField<string>("passphrase", yup.string().required("Passphrase is required"), {
  initialValue: "",
});

const hasValidationError = () => {
  if (name.value === "") {
    setNameError("Name is required");
    return true;
  }

  if (privateKeyData.value === "") {
    setPrivateKeyDataError("Private key data is required");
    return true;
  }

  if (hasPassphrase.value && passphrase.value === "") {
    setPassphraseError("Passphrase is required");
    return true;
  }

  if (!validateKey("private", privateKeyData.value, passphrase.value || undefined)) {
    setPrivateKeyDataError("Invalid private key data");
    return true;
  }

  return false;
};

const validatePrivateKeyData = () => {
  try {
    parsePrivateKeySsh(privateKeyData.value, passphrase.value || undefined);
  } catch (err) {
    if ((err as { name: string }).name === "KeyEncryptedError") {
      hasPassphrase.value = true;
      return;
    }

    setPrivateKeyDataError("Invalid private key data");
  }
};

const resetFields = () => {
  resetName();
  resetPrivateKeyData();
  resetPassphrase();
  hasPassphrase.value = false;
};

const close = () => {
  resetFields();
  showDialog.value = false;
};

const handleCreationError = (error: Error) => {
  if (error.name === "KeyParseError") {
    setPassphraseError("Passphrase is incorrect or missing.");
    return;
  }

  switch (error.message) {
    case "both": {
      setNameError("Name is already used");
      setPrivateKeyDataError("Private key data is already used");
      break;
    }
    case "name": {
      setNameError("Name is already used");
      break;
    }
    case "private_key": {
      setPrivateKeyDataError("Private key data is already used");
      break;
    }
    default: {
      snackbar.showError("Failed to create private key.");
      handleError(error);
    }
  }
};

const create = async () => {
  if (hasValidationError()) return;

  try {
    const fingerprint = createKeyFingerprint(privateKeyData.value, passphrase.value);
    await privateKeysStore.addPrivateKey({
      name: name.value,
      data: privateKeyData.value,
      hasPassphrase: hasPassphrase.value,
      fingerprint,
    });
    snackbar.showSuccess("Private key created successfully.");
    emit("update");
    close();
  } catch (error) {
    handleCreationError(error as Error);
  }
};

defineExpose({ privateKeyDataError, nameError, showDialog });
</script>
