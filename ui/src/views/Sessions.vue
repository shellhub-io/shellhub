<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
    data-test="sessions-title"
  >
    <h1>Sessions</h1>
  </div>
  <div>
    <SessionList v-if="hasSession" data-test="sessions-list" />

    <NoItemsMessage
      v-else
      item="Sessions"
      icon="mdi-history"
      data-test="no-items-message-component"
    >
      <template #content>
        <p>An SSH session is created when a connection is made to any registered device.</p>
        <p>Please follow our guide on
          <a
            rel="noopener noreferrer"
            target="_blank"
            href="https://docs.shellhub.io/user-guides/devices/connecting"
          >how to connect to your devices</a>.
        </p>
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import NoItemsMessage from "../components/NoItemsMessage.vue";
import { useStore } from "../store";
import SessionList from "../components/Sessions/SessionList.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const snackbar = useSnackbar();

onMounted(async () => {
  try {
    store.dispatch("sessions/resetPagePerpage");

    await store.dispatch("sessions/refresh");
  } catch (error: unknown) {
    snackbar.showError("Failed to load the sessions list.");
    handleError(error);
  }
});

const hasSession = computed(
  () => store.getters["sessions/getNumberSessions"] > 0,
);
</script>
