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
    :confirm-disabled
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

      <FileTextComponent
        v-model="privateKeyData"
        class="mb-2"
        enable-paste
        allow-extensionless
        textarea-label="Private key data"
        description-text="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
        :validator="encryptionAwareValidator"
        :invalid-message="ftcInvalidMessage"
        data-test="private-key-field"
        @error="onPrivateKeyError"
        @update:model-value="onPrivateKeyInput"
        @file-processed="onPrivateKeyFileProcessed"
        @mode-changed="onFileTextModeChanged"
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
import { ref, computed, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import {
  convertToFingerprint,
  parsePrivateKey,
} from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "../FormDialog.vue";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const emit = defineEmits(["update"]);
const privateKeysStore = usePrivateKeysStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });

const hasPassphrase = ref(false);
const encryptedDetected = ref(false);
const ftcInvalidMessage = ref("Invalid private key data");

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
} = useField<string>(
  "privateKeyData",
  yup.string().required("Private key data is required"),
  { initialValue: "" },
);

const {
  value: passphrase,
  errorMessage: passphraseError,
  setErrors: setPassphraseError,
  resetField: resetPassphrase,
} = useField<string>("passphrase", yup.string(), {
  initialValue: "",
});

const encryptionAwareValidator = (text: string): boolean => {
  const t = (text || "").trim();
  encryptedDetected.value = false;

  if (!t) {
    ftcInvalidMessage.value = "Private key data is required";
    return false;
  }

  try {
    parsePrivateKey(t, undefined);
    ftcInvalidMessage.value = "Invalid private key data";
    return true;
  } catch (err) {
    const { name } = (err as { name?: string });
    if (name === "KeyEncryptedError") {
      encryptedDetected.value = true;
      hasPassphrase.value = true;
      return true;
    }
    ftcInvalidMessage.value = "Invalid private key data";
    return false;
  }
};

const onPrivateKeyError = (errorMsg: string) => {
  if (errorMsg && !encryptedDetected.value) {
    setPrivateKeyDataError(errorMsg);
  } else if (encryptedDetected.value) {
    setPrivateKeyDataError("");
  } else {
    setPrivateKeyDataError("");
  }
};

const onPrivateKeyInput = () => {
  const text = (privateKeyData.value || "").trim();

  if (!text) {
    hasPassphrase.value = false;
    encryptedDetected.value = false;
    setPrivateKeyDataError("");
    setPassphraseError("");
    return;
  }

  if (passphraseError.value === "Incorrect passphrase") {
    setPassphraseError("");
  }

  try {
    parsePrivateKey(text, undefined);
    hasPassphrase.value = false;
    encryptedDetected.value = false;
    setPrivateKeyDataError("");
    setPassphraseError("");
  } catch (err) {
    const e = err as { name?: string };
    if (e.name === "KeyEncryptedError") {
      hasPassphrase.value = true;
      encryptedDetected.value = true;
      setPrivateKeyDataError("");
      if (!passphrase.value) {
        setPassphraseError("Passphrase for this private key is required");
      }
    } else {
      hasPassphrase.value = false;
      encryptedDetected.value = false;
      setPrivateKeyDataError("Invalid private key data");
      setPassphraseError("");
    }
  }
};

const onPrivateKeyFileProcessed = () => {
  onPrivateKeyInput();
};

const onFileTextModeChanged = () => {
  resetPrivateKeyData();
  resetPassphrase();
  hasPassphrase.value = false;
  encryptedDetected.value = false;
  setPrivateKeyDataError("");
  setPassphraseError("");
  privateKeyData.value = "";
};

watch(privateKeyData, (val) => {
  const text = (val || "").trim();
  if (!text) {
    resetPassphrase();
    hasPassphrase.value = false;
    encryptedDetected.value = false;
    setPassphraseError("");
    setPrivateKeyDataError("");
  }
});

const confirmDisabled = computed(() => {
  const nameReady = Boolean(name.value && name.value.trim());
  const keyReady = Boolean(privateKeyData.value && privateKeyData.value.trim());
  const passReady = hasPassphrase.value ? Boolean(passphrase.value && passphrase.value.trim()) : true;
  const anyError = Boolean(
    nameError.value
    || privateKeyDataError.value
    || (hasPassphrase.value && passphraseError.value),
  );

  return !(nameReady && keyReady && passReady) || anyError;
});

const resetFields = () => {
  resetName();
  resetPrivateKeyData();
  resetPassphrase();
  hasPassphrase.value = false;
  encryptedDetected.value = false;
  ftcInvalidMessage.value = "Invalid private key data";
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
    case "both":
      setNameError("Name is already used");
      setPrivateKeyDataError("Private key data is already used");
      break;
    case "name":
      setNameError("Name is already used");
      break;
    case "private_key":
      setPrivateKeyDataError("Private key data is already used");
      break;
    default:
      snackbar.showError("Failed to create private key.");
      handleError(error);
  }
};

const create = async () => {
  let hasError = false;

  if (!name.value || !name.value.trim()) {
    setNameError("Name is required");
    hasError = true;
  }

  if (!privateKeyData.value || !privateKeyData.value.trim()) {
    setPrivateKeyDataError("Private key data is required");
    hasError = true;
  }

  if (hasPassphrase.value && (!passphrase.value || !passphrase.value.trim())) {
    setPassphraseError("Passphrase for this private key is required");
    hasError = true;
  }

  if (privateKeyData.value && hasPassphrase.value && passphrase.value) {
    try {
      parsePrivateKey(privateKeyData.value, passphrase.value);
    } catch {
      setPassphraseError("Incorrect passphrase");
      hasError = true;
    }
  }

  if (hasError) return;

  try {
    const fingerprint = convertToFingerprint(
      privateKeyData.value,
      hasPassphrase.value ? passphrase.value : undefined,
    );

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
