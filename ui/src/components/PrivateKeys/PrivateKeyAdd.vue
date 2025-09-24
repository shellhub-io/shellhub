<template>
  <FormDialog
    v-model="showDialog"
    @close="close"
    @cancel="close"
    @confirm="create"
    title="New Private Key"
    icon="mdi-key"
    confirm-text="Save"
    cancel-text="Cancel"
    :confirm-disabled="!!privateKeyDataError || !!nameError || (hasPassphrase && !!passphraseError)"
    confirm-data-test="private-key-save-btn"
    cancel-data-test="private-key-cancel-btn"
    data-test="private-key-dialog"
  >
    <div class="px-6 pt-4">
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
        data-test="name-field"
      />

      <v-textarea
        v-model="privateKeyData"
        label="Private key data"
        required
        messages="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
        :error-messages="privateKeyDataError"
        @update:model-value="validatePrivateKeyData"
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
        data-test="passphrase-field"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { convertToFingerprint, isKeyValid, parsePrivateKey } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "../FormDialog.vue";
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

  if (!isKeyValid("private", privateKeyData.value, passphrase.value || undefined)) {
    setPrivateKeyDataError("Invalid private key data");
    return true;
  }

  return false;
};

const validatePrivateKeyData = () => {
  try {
    parsePrivateKey(privateKeyData.value, passphrase.value || undefined);
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
    const fingerprint = convertToFingerprint(privateKeyData.value, passphrase.value);
    privateKeysStore.addPrivateKey({
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
