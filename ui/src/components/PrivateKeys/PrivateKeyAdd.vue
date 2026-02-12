<template>
  <FormDialog
    v-model="showDialog"
    title="New Private Key"
    icon="mdi-key"
    confirm-text="Save"
    cancel-text="Cancel"
    :confirm-disabled
    confirm-data-test="private-key-save-btn"
    cancel-data-test="private-key-cancel-btn"
    data-test="private-key-dialog"
    @close="close"
    @cancel="close"
    @confirm="create"
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
        role="status"
        aria-live="polite"
      />

      <v-text-field
        v-model="name"
        :error-messages="nameError"
        label="Key name"
        placeholder="Name used to identify the private key"
        data-test="name-field"
        class="my-3"
        hide-details="auto"
      />

      <FileTextComponent
        v-model="privateKeyData"
        v-model:error-message="privateKeyDataError"
        class="mb-2"
        enable-paste
        allow-extensionless
        textarea-label="Private key data"
        description-text="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
        :validator="encryptionAwareValidator"
        :invalid-message="ftcInvalidMessage"
        data-test="private-key-field"
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
        autocomplete="current-password"
        data-test="passphrase-field"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import { convertToFingerprint, parsePrivateKey } from "@/utils/sshKeys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import FileTextComponent from "@/components/Fields/FileTextComponent.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const emit = defineEmits(["update"]);
const privateKeysStore = usePrivateKeysStore();
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });

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

const privateKeyData = ref("");
const privateKeyDataError = ref("");

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
    const { name } = err as { name?: string };
    if (name === "KeyEncryptedError") {
      encryptedDetected.value = true;
      hasPassphrase.value = true;
      return true;
    }
    ftcInvalidMessage.value = "Invalid private key data";
    return false;
  }
};

const onPrivateKeyInput = () => {
  const text = (privateKeyData.value || "").trim();

  if (!text) {
    hasPassphrase.value = false;
    encryptedDetected.value = false;
    privateKeyDataError.value = "";
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
    privateKeyDataError.value = "";
    setPassphraseError("");
  } catch (err) {
    const e = err as { name?: string };
    if (e.name === "KeyEncryptedError") {
      hasPassphrase.value = true;
      encryptedDetected.value = true;
      privateKeyDataError.value = "";
      if (!passphrase.value) {
        setPassphraseError("Passphrase for this private key is required");
      }
    } else {
      hasPassphrase.value = false;
      encryptedDetected.value = false;
      privateKeyDataError.value = "Invalid private key data";
      setPassphraseError("");
    }
  }
};

const onPrivateKeyFileProcessed = () => {
  onPrivateKeyInput();
};

const onFileTextModeChanged = () => {
  privateKeyData.value = "";
  resetPassphrase();
  hasPassphrase.value = false;
  encryptedDetected.value = false;
  privateKeyDataError.value = "";
  setPassphraseError("");
};

watch(privateKeyData, (val) => {
  const text = (val || "").trim();
  if (!text) {
    resetPassphrase();
    hasPassphrase.value = false;
    encryptedDetected.value = false;
    setPassphraseError("");
    privateKeyDataError.value = "";
  }
});

const confirmDisabled = computed(() => {
  const nameReady = Boolean(name.value && name.value.trim());
  const keyReady = Boolean(privateKeyData.value && privateKeyData.value.trim());
  const passReady = hasPassphrase.value
    ? Boolean(passphrase.value && passphrase.value.trim())
    : true;
  const anyError = Boolean(
    nameError.value
    || privateKeyDataError.value
    || (hasPassphrase.value && passphraseError.value),
  );

  return !(nameReady && keyReady && passReady) || anyError;
});

const resetFields = () => {
  resetName();
  privateKeyData.value = "";
  privateKeyDataError.value = "";
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
      privateKeyDataError.value = "Private key data is already used";
      break;
    case "name":
      setNameError("Name is already used");
      break;
    case "private_key":
      privateKeyDataError.value = "Private key data is already used";
      break;
    default:
      snackbar.showError("Failed to create private key.");
      handleError(error);
  }
};

const create = () => {
  let hasError = false;

  if (!name.value || !name.value.trim()) {
    setNameError("Name is required");
    hasError = true;
  }

  if (!privateKeyData.value || !privateKeyData.value.trim()) {
    privateKeyDataError.value = "Private key data is required";
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
</script>
