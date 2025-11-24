<template>
  <h1>Session Details</h1>
  <v-card
    v-if="session.uid"
    class="mt-2 border rounded bg-background"
  >
    <v-card-title class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface">
      <div class="d-flex align-center ml-2 ga-3">
        <v-tooltip
          location="bottom"
          :text="session.active ? 'Active' : 'Inactive'"
        >
          <template #activator="{ props }">
            <v-icon
              v-bind="props"
              :color="session.active ? 'success' : 'white'"
              data-test="active-icon"
              icon="mdi-check-circle"
            />
          </template>
        </v-tooltip>
        <h2 class="text-h6">{{ session.uid }}</h2>
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="session-uid-field">
            <h3 class="item-title">UID:</h3>
            <p class="text-truncate">{{ session.uid }}</p>
          </div>

          <div
            v-if="session.device"
            data-test="session-device-field"
          >
            <h3 class="item-title">Device:</h3>
            <router-link
              :to="{ name: 'deviceDetails', params: { id: session.device.uid } }"
              class="text-white"
            >
              {{ session.device.name || session.device.uid }}
            </router-link>
          </div>

          <div data-test="session-username-field">
            <h3 class="item-title">Username:</h3>
            <p>{{ session.username }}</p>
          </div>

          <div data-test="session-ip-field">
            <h3 class="item-title">IP Address:</h3>
            <code>{{ session.ip_address }}</code>
          </div>

          <div data-test="session-type-field">
            <h3 class="item-title">Type:</h3>
            <p class="text-capitalize">{{ session.type }}</p>
          </div>

          <div data-test="session-terminal-field">
            <h3 class="item-title">Terminal:</h3>
            <p>{{ session.term === "none" ? "-" : session.term }}</p>
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="session-namespace-field">
            <h3 class="item-title">Namespace:</h3>
            <router-link
              :to="{ name: 'namespaceDetails', params: { id: session.tenant_id } }"
              class="text-white"
            >
              {{ session.device.namespace }}
            </router-link>
          </div>

          <div data-test="session-authenticated-field">
            <h3 class="item-title">Authenticated:</h3>
            <p>{{ session.authenticated ? 'Yes' : 'No' }}</p>
          </div>

          <div data-test="session-recorded-field">
            <h3 class="item-title">Recorded:</h3>
            <p>{{ session.recorded ? 'Yes' : 'No' }}</p>
          </div>

          <div data-test="session-started-field">
            <h3 class="item-title">Started At:</h3>
            <p>{{ formatFullDateTime(session.started_at) }}</p>
          </div>

          <div data-test="session-last-seen-field">
            <h3 class="item-title">Last Seen:</h3>
            <p>{{ formatFullDateTime(session.last_seen) }}</p>
          </div>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { IAdminSession } from "@admin/interfaces/ISession";
import useSessionsStore from "@admin/store/modules/sessions";
import useSnackbar from "@/helpers/snackbar";
import { formatFullDateTime } from "@/utils/date";

const route = useRoute();
const snackbar = useSnackbar();
const sessionStore = useSessionsStore();
const sessionId = computed(() => route.params.id);
const session = ref({} as IAdminSession);

onMounted(async () => {
  try {
    session.value = await sessionStore.fetchSessionById(sessionId.value as string);
  } catch {
    snackbar.showError("Failed to get session details.");
  }
});
</script>

<style lang="scss" scoped>
.item-title {
  margin-top: 0.75rem;
  // Vuetify's text-overline styles
  font-size: 0.75rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1666666667em;
  line-height: 2.667;
}
</style>
