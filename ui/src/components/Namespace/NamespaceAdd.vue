<template>
  <div class="ml-2 mr-2">
    <v-btn
      v-if="!firstNamespace"
      block
      :size="isSmall ? 'small' : 'default'"
      color="primary"
      @click="showDialog = true"
    >
      Add Namespace
    </v-btn>

    <v-list-item-title>
      <v-dialog
        v-model="showDialog"
        @click:outside="update"
        min-width="350"
        max-width="450"
        v-bind="$attrs"
      >
        <v-card data-test="namespaceAdd-card" class="bg-v-theme-surface">
          <v-card-title class="text-headline bg-primary">
            Enter Namespace
          </v-card-title>

          <v-card-text>
            <v-text-field
              v-model="namespaceName"
              label="Username"
              :error-messages="namespaceNameError"
              required
              variant="underlined"
              data-test="username-text"
            />
          </v-card-text>

          <v-card-actions>
            <v-spacer />
            <v-btn data-test="close-btn" @click="update"> Close </v-btn>

            <v-btn color="primary" data-test="add-btn" @click="addNamespace">
              Add
            </v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-list-item-title>
  </div>
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "../../utils/handleError";

export default defineComponent({
  inheritAttrs: false,
  props: {
    firstNamespace: {
      type: Boolean,
      default: false,
    },
    show: {
      type: Boolean,
      default: false,
    },
    isSmall: {
      type: Boolean,
      default: false,
    },
    enableSwitchIn: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["update"],
  setup(props, ctx) {
    const store = useStore();

    const showDialog = ref(false);

    const {
      value: namespaceName,
      errorMessage: namespaceNameError,
      setErrors: setNamespaceNameError,
      resetField: resetNamespaceName,
    } = useField<string>(
      "namespaceName",
      yup
        .string()
        .required()
        .min(3)
        .max(30)
        .matches(/^[^.]*$/, "The name must not contain dots"),
      {
        initialValue: "",
      },
    );

    const switchIn = async (tenant: string) => {
      try {
        await store.dispatch("namespaces/switchNamespace", {
          tenant_id: tenant,
        });
        window.location.reload();
      } catch (error: unknown) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceSwitch,
        );
        handleError(error);
      }
    };

    const close = () => {
      showDialog.value = false;
      resetNamespaceName();
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const addNamespace = async () => {
      if (!namespaceNameError.value) {
        try {
          // const tenant = localStorage.getItem("tenant");

          const response = await store.dispatch("namespaces/post", namespaceName.value);

          if (props.firstNamespace || props.enableSwitchIn) {
            await switchIn(response.data.tenant_id);
            close();
          } else {
            await store.dispatch("namespaces/fetch", {
              page: 1,
              perPage: 30,
            });
            update();
          }

          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.namespaceCreating,
          );
        } catch (error: unknown) {
          if (axios.isAxiosError(error)) {
            const axiosError = error as AxiosError;
            if (axiosError.response?.status === 400) {
              setNamespaceNameError(
                "Your namespace should be 3-30 characters long",
              );
            } else if (axiosError.response?.status === 403) {
              setNamespaceNameError("Update your plan to create more namespaces");
            } else if (axiosError.response?.status === 409) {
              setNamespaceNameError("namespace already exists");
            }
          } else {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.namespaceCreating,
            );
            handleError(error);
          }
        }
      }
    };

    return {
      showDialog,
      addNamespace,
      update,
      namespaceName,
      namespaceNameError,
    };
  },
});
</script>
