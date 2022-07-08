<template>
  <v-list-item @click="showDialog = true" v-bind="$props">
    <div class="d-flex align-center">
      <div data-test="privatekey-icon" class="mr-2">
        <v-icon color="white"> mdi-pencil </v-icon>
      </div>

      <v-list-item-title data-test="privatekey-title">
        Edit
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog v-model="showDialog"  width="520" transition="dialog-bottom-transition">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-3 bg-primary">
        Edit Private Key
      </v-card-title>
      <form @submit.prevent="edit" class="mt-3">
        <v-card-text>
          <v-text-field
            v-model="name"
            label="Key name"
            placeholder="Name used to identify the public key"
            :error-messages="nameError"
            required
            variant="underlined"
            data-test="name-field"
          />

          <v-textarea
            v-model="keyLocal.data"
            class="mt-5"
            label="Public key data"
            readonly
            :messages="supportedKeys"
            variant="underlined"
            data-test="data-field"
            rows="5"
            :style="{ cursor: 'not-allowed' }"
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
import {
  defineComponent,
  ref,
  watch,
  onMounted,
  computed,
  nextTick,
  onUpdated,
} from "vue";
import { useStore } from "../../store";
import * as yup from "yup";
import { IPublicKey } from "../../interfaces/IPublicKey";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
// import { validateKey } from "../../utils/validate";

export default defineComponent({
  props: {
    show: {
      type: Boolean,
      required: false,
    },
    keyObject: {
      type: Object,
      required: true,
      default: Object as unknown as IPublicKey,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const showDialog = ref(false);
    const store = useStore();
    const keyLocal = ref<Partial<IPublicKey>>({
      name: "",
      username: "",
      data: "",
    });
    const supportedKeys = ref(
      "Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats."
    );

    const { value: name, errorMessage: nameError } = useField<
      string | undefined
    >("name", yup.string().required(), {
      initialValue: props.keyObject.name,
    });

    watch(name, () => {
      keyLocal.value.name = name.value;
    });

    onMounted(() => {
      setPrivateKey();
    });

    const setPrivateKey = () => {
      keyLocal.value = { ...props.keyObject };
    };

    const edit = async () => {
      if (!nameError.value) {
        // @ts-ignore
        let keySend = { ...keyLocal.value, data: btoa(keyLocal.value.data) };

        try {
          await store.dispatch("privateKey/edit", keySend);
          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.privateKeyEditing
          );
          update();
        } catch {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.publicKeyEditing
          );
        }
      }
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const close = () => {
      setPrivateKey();
      showDialog.value = false;
    };

    return {
      showDialog,
      keyLocal,
      name,
      nameError,
      supportedKeys,
      close,
      edit,
    };
  },
});
</script>
