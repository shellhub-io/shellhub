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
            :update:modelValue="validatePrivateKeyData"
            @change="validatePrivateKeyData"
            variant="underlined"
            data-test="private-key-field"
            rows="5"
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
            :disabled="!!privateKeyDataError || !!nameError"
          >
            Save
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "@/store";
import { parsePrivateKeySsh, validateKey } from "@/utils/validate";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";

const emit = defineEmits(["update"]);
const store = useStore();
const snackbar = useSnackbar();
const showDialog = defineModel({ default: false });

const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
  resetField: resetName,
} = useField<string>("name", yup.string().required(), {
  initialValue: "",
});

const {
  value: privateKeyData,
  errorMessage: privateKeyDataError,
  setErrors: setPrivateKeyDataError,
  resetField: resetPrivateKeyData,
} = useField<string>("privateKeyData", yup.string().required(), {
  initialValue: "",
});

const hasError = () => {
  if (name.value === "") {
    setNameError("Name is required");
    return true;
  }

  if (privateKeyData.value === "") {
    setPrivateKeyDataError("Private key data is required");
    return true;
  }

  if (!validateKey("private", privateKeyData.value)) {
    setPrivateKeyDataError("Not is a valid private key");
    return true;
  }

  return false;
};

const validatePrivateKeyData = () => {
  try {
    parsePrivateKeySsh(privateKeyData.value);
    return true;
  } catch (err: unknown) {
    const typedErr = err as {name: string};
    if (typedErr.name === "KeyEncryptedError") {
      setPrivateKeyDataError("Private key with passphrase is not supported");
    } else {
      setPrivateKeyDataError("Invalid private key data");
    }
    return false;
  }
};

const resetFields = () => {
  resetName();
  resetPrivateKeyData();
};

const close = () => {
  resetFields();
  showDialog.value = false;
};

const create = async () => {
  if (!hasError()) {
    try {
      await store.dispatch("privateKey/set", {
        name: name.value,
        data: privateKeyData.value,
      });
      snackbar.showSuccess("Private key created successfully.");
      emit("update");
      close();
    } catch (error) {
      const pkError = error as Error;
      switch (pkError.message) {
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
    }
  }
};

defineExpose({ privateKeyDataError, nameError, showDialog });
</script>
