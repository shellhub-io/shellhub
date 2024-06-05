<template>
  <v-row v-bind="$attrs">
    <v-col data-test="namespace-title">
      <h3>Namespace</h3>
    </v-col>

    <v-spacer />

    <v-col md="auto" class="ml-auto">
      <v-tooltip
        location="bottom"
        class="text-center"
        :disabled="hasAuthorizationRenameNamespace()"
      >
        <template v-slot:activator="{ props }">
          <div v-bind="props">
            <v-btn
              v-if="editBtn"
              :disabled="!hasAuthorizationRenameNamespace() || (!!nameError || !!connectionAnnouncementError)"
              color="primary"
              @click="editBtn = false;"
              data-test="edit-btn"
            >
              Edit Namespace
            </v-btn>
            <v-btn
              v-else
              :disabled="!!nameError || !!connectionAnnouncementError"
              color="primary"
              @click="editNamespace"
              data-test="save-btn"
            >
              Save Namespace
            </v-btn>
          </div>
        </template>
        <span> You don't have this kind of authorization. </span>
      </v-tooltip>
    </v-col>
  </v-row>

  <div class="mt-4 mb-2">
    <v-text-field
      :disabled="validateInput"
      v-model="name"
      class="ml-3"
      label="Name"
      :error-messages="nameError"
      variant="underlined"
      required
      data-test="name-text"
    />
  </div>

  <div class="mb-2">
    <v-textarea
      :disabled="validateInput"
      v-model="connectionAnnouncement"
      label="Connection Announcement"
      :error-messages="connectionAnnouncementError"
      data-test="connectionAnnouncement-text"
      variant="underlined"
      hint="A connection announcement is a custom message written
      during a session when a connection is established on a device
      within the namespace."
      persistent-hint
      required
    />
  </div>

</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useField } from "vee-validate";
import axios, { AxiosError } from "axios";
import * as yup from "yup";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const namespace = computed(() => store.getters["namespaces/get"]);
const tenant = computed(() => store.getters["auth/tenant"]);
const editBtn = ref(true);
const validateInput = computed(() => editBtn.value === true);
const {
  value: name,
  errorMessage: nameError,
  setErrors: setNameError,
} = useField<string>(
  "name",
  yup
    .string()
    .min(3, "Your namespace should be 3-30 characters long")
    .max(30, "Your namespace should be 3-30 characters long")
    .required()
    .matches(/^[^.]*$/, "The name must not contain dots"),
  {
    initialValue: "",
  },
);

const {
  value: connectionAnnouncement,
  errorMessage: connectionAnnouncementError,
  setErrors: setConnectionAnnouncementError,
} = useField<string>(
  "Connection Announcement",
  yup
    .string()
    .max(4096, "Your message should be 1-4096 characters long"),
  {
    initialValue: "",
  },
);

watch(namespace, (ns) => {
  name.value = ns.name;
  connectionAnnouncement.value = ns.settings.connection_announcement;
});

onMounted(() => {
  if (!store.getters["auth/isLoggedIn"]) return;
  store.dispatch("namespaces/get", tenant.value);
});

const hasAuthorizationRenameNamespace = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(authorizer.role[role], actions.namespace.rename);
  }

  return false;
};

const editNamespace = async () => {
  if (!nameError.value) {
    try {
      await store.dispatch("namespaces/put", {
        id: tenant.value,
        name: name.value,
        settings: {
          connection_announcement: connectionAnnouncement.value,
        },
      });
      await store.dispatch("namespaces/fetch", {
        page: 1,
        perPage: 10,
        filter: "",
      });
      await store.dispatch("namespaces/get", tenant.value);
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.namespaceEdit,
      );
      editBtn.value = true;
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        if (axiosError.response?.status === 400) {
          setNameError("This name is not valid");
          setConnectionAnnouncementError("This message is not valid");
        } else if (axiosError.response?.status === 409) {
          setNameError("name used already");
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceEdit,
          );
          handleError(error);
        }
      } else {
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceEdit,
        );
        handleError(error);
      }
    }
  }
};
</script>
