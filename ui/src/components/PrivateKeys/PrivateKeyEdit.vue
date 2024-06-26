<template>
  <div>
    <v-list-item @click="showDialog = true" v-bind="$props" data-test="privatekey-edit-btn">
      <div class="d-flex align-center">
        <div data-test="privatekey-icon" class="mr-2">
          <v-icon>mdi-pencil</v-icon>
        </div>

        <v-list-item-title data-test="privatekey-title">
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <v-dialog v-model="showDialog" width="520" transition="dialog-bottom-transition">
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
              :messages="supportedKeys"
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
            <v-btn color="primary" @click="close" data-test="pk-edit-cancel-btn">
              Cancel
            </v-btn>
            <v-btn color="primary" type="submit" data-test="pk-edit-save-btn" :disabled="!!keyLocalDataError || !!nameError">
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
import { ref, PropType, onMounted } from "vue";
import * as yup from "yup";
import { useStore } from "../../store";
import { IPublicKey } from "../../interfaces/IPublicKey";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "../../utils/handleError";
import { parsePrivateKeySsh } from "../../utils/validate";

const props = defineProps({
  show: {
    type: Boolean,
    required: false,
  },
  keyObject: {
    type: Object as PropType<Partial<IPublicKey>>,
    required: true,
  },
  style: {
    type: [String, Object],
    default: undefined,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();
const {
  value: keyLocal,
  errorMessage: keyLocalDataError,
  setErrors: setKeyLocalDataError,
} = useField<string>("privateKey", yup.string().required(), {
  initialValue: props.keyObject.data,
});

const {
  value: name,
  errorMessage: nameError,
} = useField<string>("name", yup.string().required(), {
  initialValue: props.keyObject.name ?? "", // Ensure name is a string
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

const supportedKeys = ref(
  "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
);

const setPrivateKey = () => {
  keyLocal.value = props.keyObject.data ?? "";
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
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.privateKeyEditing,
      );
      update();
    } catch (error: unknown) {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.privateKeyEditing,
      );
    }
  }
};

defineExpose({ keyLocal, isValid, name, update, edit, handleError, setPrivateKey });
</script>
