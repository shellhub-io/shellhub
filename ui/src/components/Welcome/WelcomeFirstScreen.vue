<template>
  <div class="pa-6">
    <div class="text-center mb-6">
      <v-avatar
        size="64"
        color="primary"
        class="mb-4"
      >
        <v-icon
          size="32"
          color="white"
          icon="mdi-rocket-launch"
        />
      </v-avatar>
      <h2
        class="text-h4 mb-2"
        data-test="welcome-name"
      >
        Welcome, <span class="text-primary">{{ name }}</span>!
      </h2>
      <p class="text-subtitle-1 text-medium-emphasis">
        Let's get you started with ShellHub
      </p>
    </div>

    <p class="text-body-1 my-5 text-justify">
      ShellHub is a modern SSH server that eliminates the complexity of remote device access.
      Connect to your Linux devices effortlessly without worrying about network configurations,
      firewalls, or IP addresses. Everything is automated and secure.
    </p>

    <v-card
      v-for="feature in features"
      :key="feature.title"
      variant="outlined"
      color="primary"
      class="pa-4 ma-3"
    >
      <div class="d-flex align-center h-100">
        <v-icon
          size="32"
          color="primary"
          class="mr-4"
          :icon="feature.icon"
        />
        <div class="text-high-emphasis">
          <v-card-title class="text-h6 mb-1 pa-0 text-high-emphasis">
            {{ feature.title }}
          </v-card-title>
          <v-card-subtitle class="mb-0 pl-0 text-medium-emphasis">
            {{ feature.description }}
          </v-card-subtitle>
        </div>
      </div>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import useAuthStore from "@/store/modules/auth";

interface FeatureCard {
  icon: string;
  title: string;
  description: string;
}

const authStore = useAuthStore();
const name = computed(() => authStore.name || authStore.username);

const features: FeatureCard[] = [
  {
    icon: "mdi-monitor",
    title: "Remote Access",
    description: "Access your Linux devices from anywhere via CLI or web interface",
  },
  {
    icon: "mdi-shield-check",
    title: "Secure Connection",
    description: "Bypass firewalls and NAT with secure, encrypted connections",
  },
  {
    icon: "mdi-cogs",
    title: "Easy Setup",
    description: "Automated access process with seamless device management",
  },
];

defineExpose({ name });
</script>
