<template>
  <div>
    <v-tooltip v-bind="$attrs" class="text-center" location="bottom">
      <template v-slot:activator="{ props }">
        <div v-bind="props">
          <v-btn
            @click="dialog = !dialog"
            color="primary"
            tabindex="0"
            variant="elevated"
            aria-label="Dialog Add Private Key"
            @keypress.enter="dialog = !dialog"
            data-test="private-key-dialog-btn"
          >
            Add Private Key
          </v-btn>
        </div>
      </template>
    </v-tooltip>

    <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
      <v-card class="bg-v-theme-surface">
        <v-card-title class="text-h5 pa-3 bg-primary" data-test="card-title">
          New Private Key
        </v-card-title>
        <form @submit.prevent="create" class="mt-3">
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
              v-model="privateKeyData"
              label="Private key data"
              required
              :messages="supportedKeys"
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
              color="primary"
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
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { ref } from "vue";
import * as yup from "yup";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { parsePrivateKeySsh, validateKey } from "../../utils/validate";
import { IPrivateKeyError } from "../../interfaces/IPrivateKey";
import handleError from "../../utils/handleError";

const emit = defineEmits(["update"]);
const store = useStore();
const dialog = ref(false);
const supportedKeys = ref(
  "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
);

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
  dialog.value = false;
};

const create = async () => {
  if (!hasError()) {
    try {
      await store.dispatch("privateKey/set", {
        name: name.value,
        data: privateKeyData.value,
      });
      store.dispatch(
        "snackbar/showSnackbarSuccessNotRequest",
        INotificationsSuccess.privateKeyCreating,
      );
      emit("update");
      close();
    } catch (error) {
      const pkError = error as IPrivateKeyError;
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
          store.dispatch(
            "snackbar/showSnackbarErrorNotRequest",
            INotificationsError.privateKeyCreating,
          );
          handleError(error);
        }
      }
    }
  }
};

defineExpose({ privateKeyDataError, nameError });
</script>
