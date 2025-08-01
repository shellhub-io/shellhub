<template>
  <div class="d-flex pa-0 align-center">
    <h1>Session Details</h1>
  </div>
  <v-card class="mt-2 bg-v-theme-surface" v-if="!sessionIsEmpty" data-test="sessionDetails-card">
    <v-card-title class="pa-4 d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-icon v-if="session.active" color="success" size="small" data-test="sessionActive-icon">
          mdi-check-circle
        </v-icon>
        <v-tooltip location="bottom" v-else>
          <template v-slot:activator="{ props }">
            <v-icon v-bind="props" size="small" data-test="sessionInactive-icon"> mdi-check-circle </v-icon>
          </template>
          <span>{{ getTimeFromNow(session.last_seen) }}</span>
        </v-tooltip>
        <span class="ml-2" v-if="session.device" data-test="sessionDeviceName">{{ session.device.name }}</span>
      </div>

      <div>
        <v-menu location="bottom" scrim eager>
          <template v-slot:activator="{ props }">
            <v-btn
              v-bind="props"
              variant="plain"
              class="border rounded bg-v-theme-background"
              density="comfortable"
              size="default"
              icon="mdi-format-list-bulleted"
            />
          </template>
          <v-list class="bg-v-theme-surface" lines="two" density="compact">
            <v-tooltip
              location="bottom"
              class="text-center"
              :disabled="hasAuthorizationPlay()"
              data-test="sessionPlay-tooltip"
            >
              <template v-slot:activator="{ props }">
                <div v-bind="props">
                  <SessionPlay
                    v-if="session.authenticated && session.recorded"
                    :authenticated="session.authenticated"
                    :uid="session.uid"
                    :recorded="session.recorded"
                    data-test="session-play-component"
                  />
                </div>
              </template>
              <span> You don't have this kind of authorization. </span>
            </v-tooltip>

            <v-tooltip
              location="bottom"
              class="text-center"
              :disabled="hasAuthorizationRemoveRecord()"
              data-test="session-close-tooltip"
            >
              <template v-slot:activator="{ props }">
                <div v-bind="props">
                  <SessionClose
                    v-if="session.active"
                    :uid="session.uid"
                    :device="session.device"
                    :hasAuthorization="hasAuthorizationRemoveRecord()"
                    @update="refreshSessions"
                    data-test="session-close-component"
                  />
                </div>
              </template>
              <span> You don't have this kind of authorization. </span>
            </v-tooltip>

            <v-tooltip
              location="bottom"
              class="text-center"
              :disabled="hasAuthorizationRemoveRecord()"
              data-test="session-delete-tooltip"
            >
              <template v-slot:activator="{ props }">
                <div v-bind="props">
                  <SessionDelete
                    v-if="session.uid"
                    :uid="session.uid"
                    :hasAuthorization="hasAuthorizationRemoveRecord()"
                    @update="refreshSessions"
                    data-test="session-delete-record-component"
                  />

                </div>
              </template>
              <span> You don't have this kind of authorization. </span>
            </v-tooltip>
          </v-list>
        </v-menu>
      </div>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <div>
        <div class="text-overline mt-3">uid:</div>
        <div data-test="sessionUid-field">
          <p>{{ session.uid }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">user:</div>
        <div data-test="sessionUser-field">
          <p>{{ session.username }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">Authenticated:</div>
        <div data-test="sessionAuthenticated-field">
          <v-tooltip location="bottom" v-if="session.authenticated">
            <template v-slot:activator="{ props }">
              <v-icon v-bind="props" color="success"> mdi-shield-check </v-icon>
            </template>
            <span>User has been authenticated</span>
          </v-tooltip>
          <v-tooltip bottom v-else>
            <template v-slot:activator="{ props }">
              <v-icon v-bind="props" color="error"> mdi-shield-alert </v-icon>
            </template>
            <span>User has not been authenticated</span>
          </v-tooltip>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">IP address:</div>
        <div data-test="sessionIpAddress-field">
          <code class="bg-tabs pa-1">
            {{ session.ip_address }}
          </code>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">Started:</div>
        <div data-test="sessionStartedAt-field">
          <p>{{ formatFullDateTime(session.started_at) }}</p>
        </div>
      </div>

      <div>
        <div>Last seen:</div>
        <div data-test="sessionLastSeen-field">
          <p>{{ formatFullDateTime(session.last_seen) }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "../store";
import { formatFullDateTime, getTimeFromNow } from "..//utils/date";
import hasPermission from "..//utils/permission";
import { ISession } from "../interfaces/ISession";
import { authorizer, actions } from "../authorizer";
import SessionDelete from "../components/Sessions/SessionDelete.vue";
import SessionClose from "../components/Sessions/SessionClose.vue";
import SessionPlay from "../components/Sessions/SessionPlay.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const store = useStore();
const route = useRoute();
const snackbar = useSnackbar();
const sessionId = computed(() => route.params.id);
const session = ref({} as ISession);

onMounted(async () => {
  try {
    await store.dispatch("sessions/get", sessionId.value);
    session.value = store.getters["sessions/get"];
  } catch (error: unknown) {
    snackbar.showError("Failed to load session details.");
    handleError(error);
  }
});

const sessionIsEmpty = computed(
  () => store.getters["sessions/get"]
        && store.getters["sessions/get"].lenght === 0,
);

const refreshSessions = async () => {
  try {
    await store.dispatch("sessions/get", sessionId.value);
    session.value = store.getters["sessions/get"];
  } catch (error: unknown) {
    snackbar.showError("Failed to load session details.");
    handleError(error);
  }
};

const hasAuthorizationRemoveRecord = () => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.session.removeRecord);
};

const hasAuthorizationPlay = () => {
  const role = store.getters["auth/role"];
  return !!role && hasPermission(authorizer.role[role], actions.session.play);
};
</script>
