<template>
  <v-menu
    scrim
    location="bottom end"
    :offset="4"
  >
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        size="medium"
        color="primary"
        icon
        data-test="user-menu-btn"
      >
        <UserIcon
          size="1.5rem"
          :email="userEmail"
          data-test="user-icon"
        />
      </v-btn>
    </template>

    <v-card
      :width="$vuetify.display.thresholds.sm / 2"
      border
    >
      <v-list class="bg-v-theme-surface pa-0">
        <div class="pa-6 text-center">
          <UserIcon
            size="4rem"
            :email="userEmail"
            class="mb-4"
            data-test="user-icon-large"
          />
          <div class="text-h6 font-weight-medium mb-1">
            {{ primaryLabel }}
          </div>
          <div
            v-if="showSecondary"
            class="text-body-2 text-medium-emphasis"
          >
            {{ userEmail }}
          </div>
        </div>

        <v-divider />

        <div>
          <v-list-item
            v-for="item in menuItems"
            :key="item.title"
            :value="item"
            :data-test="item.title"
            :prepend-icon="item.icon"
            @click="emit('select', item)"
          >
            <v-list-item-title class="font-weight-medium">
              {{ item.title }}
            </v-list-item-title>
          </v-list-item>
        </div>

        <v-divider />

        <v-list-item @click="emit('toggle-dark-mode')">
          <template #prepend>
            <v-icon
              :icon="isDarkMode ? 'mdi-brightness-6' : 'mdi-brightness-6'"
              size="small"
            />
          </template>
          <v-list-item-title class="font-weight-medium">
            {{ isDarkMode ? "Dark Mode" : "Light Mode" }}
          </v-list-item-title>
          <template #append>
            <v-switch
              :model-value="isDarkMode"
              data-test="dark-mode-switch"
              color="primary"
              density="comfortable"
              false-icon="mdi-weather-sunny"
              true-icon="mdi-weather-night"
              hide-details
              readonly
            />
          </template>
        </v-list-item>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { computed } from "vue";
import UserIcon from "@/components/User/UserIcon.vue";

type MenuItem = {
  title: string;
  icon: string;
  [key: string]: unknown;
};

const props = defineProps<{
  userEmail: string;
  displayName?: string;
  menuItems: MenuItem[];
  isDarkMode: boolean;
}>();

const emit = defineEmits<{
  (e: "select", item: MenuItem): void;
  (e: "toggle-dark-mode"): void;
}>();

const primaryLabel = computed(() => props.displayName || props.userEmail || "User");
const showSecondary = computed(() => !!props.userEmail && !!props.displayName && props.userEmail !== props.displayName);
</script>
