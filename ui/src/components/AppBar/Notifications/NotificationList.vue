<template>
  <v-list @click.stop class="pa-0" density="compact">
    <v-list-subheader>Pending Devices</v-list-subheader>
    <v-divider />

    <v-list-item
      class="pr-0"
      v-for="notification in notifications"
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
            @update="$emit('update')"
          />
        </v-list-item-action>
      </template>
    </v-list-item>
  </v-list>
</template>

<script setup lang="ts">
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";

defineProps<{
  notifications: Array<{
    uid: string;
    name: string;
  }>,
}>();
</script>
