<template>
  <div>
    <v-list-item @click="open" data-test="privatekey-edit-btn">
      <div class="d-flex align-center">
        <div data-test="privatekey-icon" class="mr-2">
          <v-icon>mdi-pencil</v-icon>
        </div>

        <v-list-item-title data-test="privatekey-title">
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary">
          Edit Private Key
        </v-card-title>
        <form @submit.prevent="edit" class="mt-3">
          <v-card-text>
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
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn @click="close" data-test="pk-edit-cancel-btn">
              Cancel
            </v-btn>
            <v-btn
              color="primary"
              type="submit"
              data-test="pk-edit-save-btn"
              :disabled="!!keyLocalError || !!nameError || (hasPassphrase && !!passphraseError)"
            >
              Save
            </v-btn>
          </v-card-actions>
        </form>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref } from "vue";
import * as yup from "yup";
import handleError from "@/utils/handleError";
import { createKeyFingerprint, parsePrivateKeySsh, validateKey } from "@/utils/validate";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
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
    parsePrivateKeySsh(keyLocal.value, passphrase.value || undefined);
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
  hasPassphrase.value = privateKey.hasPassphrase;
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

  if (!validateKey("private", keyLocal.value, passphrase.value || undefined)) {
    setKeyLocalError("Invalid private key data");
    return;
  }

  snackbar.showError("Failed to update private key.");
};

const edit = async () => {
  if (hasValidationError()) return;

  try {
    const fingerprint = createKeyFingerprint(keyLocal.value, passphrase.value);
    await privateKeysStore.editPrivateKey({
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
defineExpose({ keyLocal, name, hasPassphrase, update, edit, handleError, initializeFormData });
</script>
