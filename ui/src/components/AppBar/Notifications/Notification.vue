<template>
  <v-menu :close-on-content-click="true">
    <template v-slot:activator="{ props }">
      <v-badge
        v-if="showNotifications"
        :content="notificationCount"
        offset-y="-5"
        location="top right"
        color="success"
        size="x-small"
        data-test="notifications-badge"
        class="ml-2 mr-2"
      >
        <v-icon
          v-bind="props"
          color="primary"
          aria-label="Open notifications menu"
        >
          mdi-bell
        </v-icon>
      </v-badge>
      <v-icon
        v-bind="props"
        v-else
        class="ml-2 mr-1"
        color="primary"
        aria-label="Open notifications menu"
      >
        mdi-bell
      </v-icon>
    </template>

    <v-card
      v-if="showNotifications"
      data-test="notifications-card"
      offset-x="20"
    >
      <v-list @click.stop class="pa-0" density="compact">
        <v-list-subheader>Pending Devices</v-list-subheader>
        <v-divider />

        <v-list-item
          class="pr-0"
          v-for="notification in notificationList"
          :key="notification.uid"
        >
          <template v-slot:prepend>
            <v-list-item-title>
              <router-link
                :to="{ name: 'DeviceDetails', params: { id: notification.uid } }"
                :data-test="notification.uid + '-field'"
              >
                {{ notification.name }}
              </router-link>
            </v-list-item-title>
          </template>

          <template v-slot:append>
            <v-list-item-action class="ma-0">
              <DeviceActionButton
                :uid="notification.uid"
                :name="notification.name"
                variant="device"
                :notification-status="true"
                :show="true"
                action="accept"
                :data-test="notification.uid + '-btn'"
                @update="fetchNotifications"
              />
            </v-list-item-action>
          </template>
        </v-list-item>
      </v-list>

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
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

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
