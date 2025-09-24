<template>
  <div>
    <v-list-item @click="open" data-test="privatekey-edit-btn">
      <div class="d-flex align-center">
        <div data-test="privatekey-icon" class="mr-2">
          <v-icon>mdi-pencil</v-icon>
        </div>
        <v-list-item-title data-test="privatekey-title">Edit</v-list-item-title>
      </div>
    </v-list-item>

    <FormDialog
      v-model="showDialog"
      @close="close"
      @cancel="close"
      @confirm="edit"
      title="Edit Private Key"
      icon="mdi-key"
      confirm-text="Save"
      cancel-text="Cancel"
      :confirm-disabled="confirmDisabled"
      confirm-data-test="pk-edit-save-btn"
      cancel-data-test="pk-edit-cancel-btn"
      data-test="private-key-edit-dialog"
    >
      <div class="px-6 pt-4">
        <v-text-field
          v-model="name"
          :error-messages="nameError"
          label="Key name"
          placeholder="Name used to identify the private key"
          variant="underlined"
          data-test="name-field"
        />

        <v-textarea
          v-model="keyLocal"
          label="Private key data"
          required
          messages="Supports RSA, DSA, ECDSA (NIST P-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
          :error-messages="keyLocalError"
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
      </div>
    </FormDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref, computed } from "vue";
import * as yup from "yup";
import handleError from "@/utils/handleError";
import { convertToFingerprint, isKeyValid, parsePrivateKey } from "@/utils/sshKeys";
import useSnackbar from "@/helpers/snackbar";
import FormDialog from "../FormDialog.vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import usePrivateKeysStore from "@/store/modules/private_keys";

const { privateKey } = defineProps<{ privateKey: IPrivateKey }>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const privateKeysStore = usePrivateKeysStore();
const snackbar = useSnackbar();
const hasPassphrase = ref(privateKey.hasPassphrase || false);

const {
  value: keyLocal,
  errorMessage: keyLocalError,
  setErrors: setKeyLocalError,
} = useField<string>("keyLocal", yup.string().required(), {
  initialValue: privateKey.data,
});

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
} = useField<string>("name", yup.string().required("Name is required"), {
  initialValue: privateKey.name ?? "",
});

const {
  value: passphrase,
  errorMessage: passphraseError,
  setErrors: setPassphraseError,
  resetField: resetPassphrase,
} = useField<string>("passphrase", yup.string().required("Passphrase is required"), {
  initialValue: "",
});

const validatePrivateKeyData = () => {
  try {
    parsePrivateKey(keyLocal.value, passphrase.value || undefined);
    setKeyLocalError("");
    hasPassphrase.value = false;
  } catch (err: unknown) {
    const typedErr = err as { name: string };
    if (typedErr.name === "KeyEncryptedError") {
      hasPassphrase.value = true;
      return;
    }
    setKeyLocalError("Invalid private key data");
  }
};

const initializeFormData = () => {
  name.value = privateKey.name ?? "";
  keyLocal.value = privateKey.data ?? "";
  setKeyLocalError("");
};

const open = () => {
  showDialog.value = true;
  initializeFormData();
};

const close = () => {
  resetPassphrase();
  hasPassphrase.value = privateKey.hasPassphrase || false;
  showDialog.value = false;
};

const hasValidationError = () => {
  if (name.value === "") {
    setNameError("Name is required");
    return true;
  }
  if (keyLocal.value === "") {
    setKeyLocalError("Private key data is required");
    return true;
  }
  if (hasPassphrase.value && passphrase.value === "") {
    setPassphraseError("Passphrase is required");
    return true;
  }
  // If present, ensure it's a valid private key
  if (!isKeyValid("private", keyLocal.value, passphrase.value || undefined)) {
    setKeyLocalError("Invalid private key data");
    return true;
  }
  return false;
};

const update = () => {
  emit("update");
  close();
};

const handleEditError = (error: Error) => {
  if (error.name === "KeyParseError") {
    setPassphraseError("Passphrase is incorrect or missing.");
    return;
  }
  if (!isKeyValid("private", keyLocal.value, passphrase.value || undefined)) {
    setKeyLocalError("Invalid private key data");
    return;
  }
  snackbar.showError("Failed to update private key.");
  handleError(error);
};

const edit = () => {
  if (hasValidationError()) return;

  try {
    const fingerprint = convertToFingerprint(keyLocal.value, passphrase.value);
    privateKeysStore.editPrivateKey({
      id: privateKey.id,
      name: name.value,
      data: keyLocal.value,
      hasPassphrase: hasPassphrase.value,
      fingerprint,
    });
    snackbar.showSuccess("Private key updated successfully.");
    update();
  } catch (error) {
    handleEditError(error as Error);
  }
};

const confirmDisabled = computed(() => Boolean(
  keyLocalError.value
    || nameError.value
    || (hasPassphrase.value && passphraseError.value),
));

defineExpose({ keyLocal, name, hasPassphrase, update, edit, handleError, initializeFormData });
</script>
