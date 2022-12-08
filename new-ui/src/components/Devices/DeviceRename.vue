<template>
  <v-list-item v-bind="$props, $attrs" @click="showDialog = true">
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="rename-icon" color="white"> mdi-pencil </v-icon>
      </div>

      <v-list-item-title data-test="rename-title"> Rename </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="500" v-model="showDialog">
    <v-card class="bg-v-theme-surface" data-test="deviceRename-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="text-title">
        Rename Device
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <v-text-field
          v-model="editName"
          label="Hostname"
          :error-messages="editNameError"
          :messages="messages"
          require
          variant="underlined"
          data-test="hostname-field"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn data-test="close-btn" variant="text" @click="showDialog = false">
          Close
        </v-btn>

        <v-btn
          data-test="rename-btn"
          color="primary darken-1"
          variant="text"
          @click="rename()"
        >
          Rename
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import { useField } from "vee-validate";
import * as yup from "yup";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

export default defineComponent({
  props: {
    uid: {
      type: String,
      required: true,
    },

    name: {
      type: String,
      default: true,
    },
  },
  emits: ["new-hostname"],
  setup(props, ctx) {
    const showDialog = ref(false);
    const messages = ref(
      "Examples: (foobar, foo-bar-ba-z-qux, foo-example, 127-0-0-1)"
    );
    const store = useStore();

    const {
      value: editName,
      errorMessage: editNameError,
      setErrors: setEditNameError,
    } = useField<string | undefined>("name", yup.string().required(), {
      initialValue: props.name,
    });

    const rename = async () => {
      try {
        await store.dispatch("devices/rename", {
          uid: props.uid,
          name: { name: editName.value },
        });

        ctx.emit("new-hostname", editName.value);
        close();
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.deviceRename
        );
      } catch (error: any) {
        if (error.response.status === 400) {
          setEditNameError("nonStandardCharacters");
        } else if (error.response.status === 409) {
          setEditNameError("The name already exists in the namespace");
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.deviceRename
          );
        }
      }
    };

    const close = () => {
      setEditNameError("");
      showDialog.value = false;
    };

    return {
      editName,
      editNameError,
      messages,
      showDialog,
      rename,
    };
  },
});
</script>
