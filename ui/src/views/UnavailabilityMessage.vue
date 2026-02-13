<template>
  <div class="d-flex-col text-center">
    <v-icon
      color="error"
      size="80"
      class="mt-4 mb-6"
      data-test="unavailability-icon"
      :icon="message.icon"
    />

    <h1
      class="text-h4 font-weight-bold mb-4"
      data-test="unavailability-title"
      role="status"
      aria-live="polite"
    >
      {{ message.title }}
    </h1>

    <p
      class="text-body-1 mb-6"
      data-test="unavailability-description"
    >
      {{ message.description }}
    </p>

    <v-btn
      color="primary"
      size="large"
      :loading="retrying"
      :disabled="retrying"
      data-test="retry-button"
      :text="message.actionText"
      @click="handleRetry"
    />

    <div
      v-if="message.supportLink"
      class="my-4"
    >
      <v-btn
        variant="text"
        :href="message.supportLink.url"
        target="_blank"
        rel="noopener noreferrer"
        data-test="support-link"
      >
        {{ message.supportLink.text }}
        <v-icon
          end
          size="small"
          icon="mdi-open-in-new"
        />
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import useUsersStore from "@/store/modules/users";
import { useUnavailabilityMessage } from "@/composables/useUnavailabilityMessage";
import useSnackbar from "@/helpers/snackbar";

const router = useRouter();
const snackbar = useSnackbar();
const usersStore = useUsersStore();
const retrying = ref(false);

const message = useUnavailabilityMessage();

const handleRetry = async () => {
  retrying.value = true;
  try {
    await usersStore.checkHealth();
    const redirect = router.currentRoute.value.query.redirect as string;
    await router.push(redirect || "/");
  } catch {
    snackbar.showError("Still unable to connect to the API. Please try again later.");
  } finally {
    retrying.value = false;
  }
};
</script>
