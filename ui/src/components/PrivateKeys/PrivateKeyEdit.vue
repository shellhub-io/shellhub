<template>
  <div>
    <v-list-item @click="showDialog = true" data-test="privatekey-edit-btn">
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
              :error-messages="keyLocalDataError"
              :update:modelValue="validatePrivateKeyData"
              @change="validatePrivateKeyData"
              variant="underlined"
              data-test="private-key-field"
              rows="5"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn @click="close" data-test="pk-edit-cancel-btn">
              Cancel
            </v-btn>
            <v-btn color="primary" type="submit" data-test="pk-edit-save-btn" :disabled="!!keyLocalDataError || !!nameError">
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
import { ref, onMounted } from "vue";
import * as yup from "yup";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import { parsePrivateKeySsh } from "@/utils/validate";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import { IPrivateKey } from "@/interfaces/IPrivateKey";

const { privateKey } = defineProps<{ privateKey: IPrivateKey }>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const snackbar = useSnackbar();
const {
  value: keyLocal,
  errorMessage: keyLocalDataError,
  setErrors: setKeyLocalDataError,
} = useField<string>("privateKey", yup.string().required(), {
  initialValue: privateKey.data,
});

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: privateKey.name ?? "", // Ensure name is a string
});

const isValid = ref(true);

const validatePrivateKeyData = () => {
  try {
    parsePrivateKeySsh(keyLocal.value);
    isValid.value = true;
    keyLocalDataError.value = "";
  } catch (err: unknown) {
    const typedErr = err as { name: string };
    if (typedErr.name === "KeyEncryptedError") {
      setKeyLocalDataError("Private key with passphrase is not supported");
    } else {
      setKeyLocalDataError("Invalid private key data");
    }
    isValid.value = false;
  }
};

const setPrivateKey = () => {
  keyLocal.value = privateKey.data ?? "";
};

onMounted(() => {
  setPrivateKey();
});

const close = () => {
  setPrivateKey();
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (!nameError.value && isValid.value) {
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    const keySend = { name: name.value, data: keyLocal.value };

    try {
      await store.dispatch("privateKey/edit", keySend);
      snackbar.showSuccess("Private key updated successfully.");
      update();
    } catch (error: unknown) {
      snackbar.showError("Failed to update private key.");
    }
  }
};

defineExpose({ keyLocal, isValid, name, update, edit, handleError, setPrivateKey });
</script>
