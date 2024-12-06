<template>
  <v-dialog v-model="show" max-width="700" @click:outside="close">
    <v-card data-test="password-change-card" class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary" data-test="title">
        Change Connection Announcement
      </v-card-title>

      <v-card-text class="mt-4 mb-3 pb-1">
        <div class="mt-4 pl-4 pr-4">
          <v-textarea
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
            auto-grow
            max-rows="25"
          />
        </div>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="close">
          Cancel
        </v-btn>

        <v-btn
          color="primary"
          variant="text"
          data-test="change-connection-btn"
          :disabled="!!connectionAnnouncementError"
          @click="updateAnnouncement()"
        >
          Save Announcement
        </v-btn>

      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from "vue";
import { useField } from "vee-validate";
import axios from "axios";
import * as yup from "yup";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const store = useStore();
const namespace = computed(() => store.getters["namespaces/get"]);
const tenant = computed(() => store.getters["auth/tenant"]);
const show = defineModel({ default: false });
const emit = defineEmits(["update"]);

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

const close = () => {
  connectionAnnouncement.value = namespace.value.settings.connection_announcement;
  show.value = false;
};

watch(namespace, (ns) => {
  connectionAnnouncement.value = ns.settings.connection_announcement;
});

onMounted(() => {
  if (!store.getters["auth/isLoggedIn"]) return;
  store.dispatch("namespaces/get", tenant.value);
});

const handleUpdateNameError = (error: unknown): void => {
  if (axios.isAxiosError(error)) {
    switch (error.response?.status) {
      case 400:
        setConnectionAnnouncementError("This message is not valid");
        break;
      default:
        store.dispatch(
          "snackbar/showSnackbarErrorAction",
          INotificationsError.namespaceEdit,
        );
        handleError(error);
    }
  }

  store.dispatch(
    "snackbar/showSnackbarErrorAction",
    INotificationsError.namespaceEdit,
  );
  handleError(error);
};

const updateAnnouncement = async () => {
  try {
    await store.dispatch("namespaces/put", {
      id: tenant.value,
      settings: {
        connection_announcement: connectionAnnouncement.value,
      },
    });

    await store.dispatch("namespaces/fetch", {
      page: 1,
      perPage: 10,
      filter: "",
    });

    emit("update");
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceEdit,
    );
    show.value = false;
  } catch (error) {
    handleUpdateNameError(error);
  }
};

defineExpose({ show });
</script>
