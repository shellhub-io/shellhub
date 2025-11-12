<template>
  <v-container
    class="h-100 d-flex align-center justify-center"
    fluid
  >
    <v-card
      :max-width="thresholds.md"
      class="pa-3 pa-sm-6 border"
    >
      <v-card-text class="text-center">
        <v-avatar
          color="warning"
          size="80"
          class="mb-6"
        >
          <v-icon
            size="48"
            icon="mdi-shield-lock-outline"
          />
        </v-avatar>

        <h1 class="text-h4 font-weight-bold mb-4">Admin Access Required</h1>

        <p class="text-body-1 text-medium-emphasis mb-6">
          You don't have administrator privileges to access the Admin Console.
          This area is restricted to system administrators only.
        </p>

        <v-divider class="my-6" />

        <div class="text-left mb-6">
          <h2 class="text-h6 font-weight-medium mb-3">What you can do:</h2>
          <v-list density="compact">
            <v-list-item
              v-for="(item, index) in actionItems"
              :key="index"
            >
              <template #prepend>
                <v-icon
                  color="primary"
                  icon="mdi-check-circle"
                />
              </template>
              <v-list-item-title class="text-wrap">{{ item }}</v-list-item-title>
            </v-list-item>
          </v-list>
        </div>

        <v-alert
          color="primary"
          variant="tonal"
          class="mb-6 text-left"
        >
          <template #prepend>
            <v-icon icon="mdi-information" />
          </template>
          <div class="text-body-2">
            If you believe you should have admin access, please contact your system administrator.
          </div>
        </v-alert>
      </v-card-text>

      <v-card-actions class="justify-center flex-column flex-sm-row px-6 pb-6 ga-3">
        <v-btn
          prepend-icon="mdi-logout"
          size="large"
          text="Logout"
          variant="tonal"
          class="w-100 w-sm-50 w-md-33"
          @click="logout"
        />
        <v-btn
          color="primary"
          size="large"
          prepend-icon="mdi-arrow-left"
          variant="flat"
          text="Go to ShellHub"
          class="w-100 w-sm-50 w-md-33"
          @click="goToMainApp"
        />
      </v-card-actions>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import useAuthStore from "@admin/store/modules/auth";
import { useDisplay } from "vuetify";

const { thresholds } = useDisplay();

const authStore = useAuthStore();

const actionItems = [
  "Return to the main ShellHub application",
  "Contact your system administrator for admin access",
  "Manage your devices, sessions, and namespaces in the main app",
];

const goToMainApp = () => {
  window.location.href = "/";
};

const logout = () => {
  authStore.logout();
  window.location.href = "/login";
};
</script>
