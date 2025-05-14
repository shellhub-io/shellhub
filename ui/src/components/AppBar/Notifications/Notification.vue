<template>
  <v-menu :close-on-content-click="true">
    <template v-slot:activator="{ props }">
      <v-badge
        :model-value="showNotifications"
        :content="notificationCount"
        offset-y="-5"
        location="top right"
        color="success"
        size="x-small"
        data-test="notifications-badge"
        class="ml-3 mr-2"
      >
        <v-icon
          v-bind="props"
          color="primary"
          aria-label="Open notifications menu"
          icon="mdi-bell"
        />
      </v-badge>
    </template>

    <v-card
      v-if="showNotifications"
      data-test="notifications-card"
    >
      <NotificationList
        :notifications
        @update="fetchNotifications"
      />

      <v-btn
        to="/devices/pending"
        variant="tonal"
        link
        block
        size="small"
        data-test="show-btn"
      >
        Show all Pending Devices
      </v-btn>
    </v-card>

    <v-card
      v-else
      data-test="empty-card"
      class="pa-2 bg-v-theme-surface"
    >
      <v-card-subtitle>{{ emptyCardMessage }}</v-card-subtitle>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { computed, onBeforeMount } from "vue";
import { useStore } from "@/store";
import { authorizer, actions } from "@/authorizer";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import NotificationList from "./NotificationList.vue";

const store = useStore();
const snackbar = useSnackbar();
const notifications = computed(() => store.getters["notifications/list"]);
const notificationCount = computed(() => store.getters["notifications/getNumberNotifications"]);
const canViewNotifications = computed(() => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.notification.view);
});
const showNotifications = computed(() => notificationCount.value > 0 && canViewNotifications.value);
const emptyCardMessage = computed(() => (
  canViewNotifications.value ? "You don't have notifications" : "You don't have permission to view notifications"
));

const fetchNotifications = async () => {
  try {
    await store.dispatch("notifications/fetch");
  } catch (error: unknown) {
    if (canViewNotifications.value) {
      snackbar.showError("Failed to load notifications.");
      handleError(error);
    }
  }
};

onBeforeMount(async () => {
  await fetchNotifications();
});
</script>
