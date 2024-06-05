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
        :max-width="!openVersion ? '400' : '650'"
        v-bind="$attrs"
      >
        <v-card data-test="namespaceAdd-card" class="bg-v-theme-surface">
          <div v-if="!openVersion">
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
          </div>
          <div v-else>
            <v-card-title class="bg-primary">
              Add a namespace using the CLI
            </v-card-title>

            <v-card-text class="mt-4 mb-0 pb-1 mb-4">
              <p class="text-body-2">
                In the Community Edition of ShellHub, namespaces must be added using the administration CLI.
                For detailed instructions on how to add namespaces, please refer to the documentation at the ShellHub Administration Guide.
              </p>
              <div id="cli-instructions" class="mt-3 text-body-2">
                <p class="text-caption mb-0 mt-3" data-test="openContentSecond-text">
                  Check the
                  <a
                    :href="'https://docs.shellhub.io/self-hosted/administration'"
                    target="_blank"
                    rel="noopener noreferrer"
                  >ShellHub Administration Guide</a
                  >
                  for more information.
                </p>
              </div>
            </v-card-text>
          </div>
        </v-card>
      </v-dialog>
    </v-list-item-title>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "../../utils/handleError";
import { envVariables } from "@/envVariables";

const props = defineProps({
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
});
const emit = defineEmits(["update"]);
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
  emit("update");
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

const openVersion = computed(() => !envVariables.isCloud || !envVariables.isEnterprise);
</script>
