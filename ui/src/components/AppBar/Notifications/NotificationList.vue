<template>
  <v-list @click.stop class="pa-0" density="compact">
    <v-list-subheader>Pending Devices</v-list-subheader>
    <v-divider />

    <v-list-item
      v-for="notification in notifications"
      :key="notification.id"
    >
      <template v-slot:prepend>
        <v-list-item-title>
          <router-link
            :to="{ name: 'DeviceDetails', params: { id: notification.data.uid } }"
            :data-test="notification.data.uid + '-field'"
          >
            {{ notification.data.name }}
          </router-link>
        </v-list-item-title>
      </template>

      <template v-slot:append>
        <v-list-item-action>
          <DeviceActionButton
            :uid="notification.data.uid"
            :name="notification.data.name"
            :variant="notification.type"
            :isInNotification="true"
            :show="true"
            action="accept"
            :data-test="notification.data.uid + '-btn'"
          />
        </v-list-item-action>
      </template>
    </v-list-item>
  </v-list>
</template>

<script setup lang="ts">
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import { INotification } from "@/interfaces/INotification";

defineProps<{
  notifications: Array<INotification>,
}>();
</script>
