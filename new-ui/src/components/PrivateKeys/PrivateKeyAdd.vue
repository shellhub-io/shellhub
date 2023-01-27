<template>
  <v-tooltip v-bind="$attrs" class="text-center" location="bottom" :disabled="hasAuthorization">
    <template v-slot:activator="{ props }">
      <div v-bind="props">
        <v-btn
          @click="dialog = !dialog"
          color="primary"
          tabindex="0"
          variant="elevated"
          aria-label="Dialog Add Private Key"
          :disabled="!hasAuthorization"
          @keypress.enter="dialog = !dialog"
          :size="size"
          data-test="private-key-add-btn"
        >
          Add Private Key
        </v-btn>
      </div>
    </template>
    <span> You don't have this kind of authorization. </span>
  </v-tooltip>

  <v-dialog v-model="dialog" width="520" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
        New Private Key
      </v-card-title>
      <form @submit.prevent="create" class="mt-3">
        <v-card-text>
          <v-text-field
            v-model="name"
            :error-messages="nameError"
            label="Name"
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
            data-test="data-field"
            rows="5"
          />
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="primary"
            text
            @click="close"
            data-test="device-add-cancel-btn"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            text
            type="submit"
            data-test="device-add-save-btn"
          >
            Save
          </v-btn>
        </v-card-actions>
      </form>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { useField } from "vee-validate";
import { computed, defineComponent, ref } from "vue";
import * as yup from "yup";
import { actions, authorizer } from "../../authorizer";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { validateKey } from "../../utils/validate";

export default defineComponent({
  props: {
    size: {
      type: String,
      default: "default",
      required: false,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();
    const dialog = ref(false);
    const supportedKeys = ref(
      "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.",
    );

    const {
      value: name,
      errorMessage: nameError,
      setErrors: setnameError,
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
        setnameError("Name is required");
        return true;
      }

      if (privateKeyData.value === "") {
        setPrivateKeyDataError("Public key data is required");
        return true;
      }

      if (!validateKey("private", privateKeyData.value)) {
        setPrivateKeyDataError("Not is a valid private key");
        return true;
      }

      return false;
    };

    const validatePrivateKeyData = () => {
      const isValid = validateKey("private", privateKeyData.value);
      if (!isValid) {
        setPrivateKeyDataError("Not is a valid private key");
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
          ctx.emit("update");
          close();
        } catch (error: any) {
          switch (true) {
            case error.message === "both": {
              setnameError("Name is already used");
              setPrivateKeyDataError("Public key data is already used");
              break;
            }
            case error.message === "name": {
              setnameError("Name is already used");
              break;
            }
            case error.message === "private_key": {
              setPrivateKeyDataError("Public key data is already used");
              break;
            }
            default: {
              store.dispatch(
                "snackbar/showSnackbarErrorNotRequest",
                INotificationsError.privateKeyCreating,
              );
              throw new Error(error);
            }
          }
        }
      }
    };

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.publicKey.create,
        );
      }
      return false;
    });

    return {
      dialog,
      name,
      nameError,
      privateKeyData,
      privateKeyDataError,
      supportedKeys,
      hasAuthorization,
      validatePrivateKeyData,
      create,
      close,
    };
  },
});
</script>
