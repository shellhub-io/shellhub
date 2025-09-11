<template>
  <div class="d-flex pa-0 align-center">
    <h1>Session Details</h1>
  </div>
  <v-card class="mt-2 border rounded bg-background" v-if="session.uid" data-test="session-details-card" elevation="0">
    <v-card-title class="bg-v-theme-surface pa-4 d-flex align-center justify-space-between">
      <div class="d-flex align-center">
        <v-tooltip location="bottom" :disabled="session.active">
          <template v-slot:activator="{ props }">
            <v-icon
              v-bind="props"
              :color="session.active ? 'success' : 'white'"
              size="small"
              data-test="session-active-icon"
              icon="mdi-check-circle" />
          </template>
          <span>{{ getTimeFromNow(session.last_seen) }}</span>
        </v-tooltip>
        <DeviceLink v-if="session.device" :device-uid="session.device.uid" :device-name="session.device.name" class="ml-2" />
      </div>

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
          <SessionPlay
            :uid="session.uid"
            :recorded="session.recorded"
            :authenticated="session.authenticated"
            v-slot="{ loading, disabled, openDialog }"
          >
            <div>
              <v-list-item @click="openDialog" :loading :disabled>
                <div class="d-flex align-center">
                  <v-icon icon="mdi-play" class="mr-2" />
                  <v-list-item-title>
                    Play Session
                  </v-list-item-title>
                </div>
              </v-list-item>
            </div>
          </SessionPlay>
          <v-tooltip
            location="bottom"
            class="text-center"
            :disabled="canRemoveSessionRecord"
            data-test="session-close-tooltip"
          >
            <template v-slot:activator="{ props }">
              <div v-bind="props">
                <SessionClose
                  v-if="session.active"
                  :uid="session.uid"
                  :device="session.device"
                  :hasAuthorization="canRemoveSessionRecord"
                  @update="getSession"
                  data-test="session-close-component"
                />
              </div>
            </template>
            <span> You don't have this kind of authorization. </span>
          </v-tooltip>

          <v-tooltip
            location="bottom"
            class="text-center"
            :disabled="canRemoveSessionRecord"
            data-test="session-delete-tooltip"
          >
            <template v-slot:activator="{ props }">
              <div v-bind="props">
                <SessionDelete
                  :uid="session.uid"
                  :hasAuthorization="canRemoveSessionRecord"
                  @update="getSession"
                  data-test="session-delete-record-component"
                />

              </div>
            </template>
            <span> You don't have this kind of authorization. </span>
          </v-tooltip>
        </v-list>
      </v-menu>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col cols="12" md="6" class="my-0 py-0">
          <div data-test="session-uid-field">
            <div class="item-title">UID:</div>
            <p class="text-truncate">{{ session.uid }}</p>
          </div>

          <div data-test="session-user-field">
            <div class="item-title">User:</div>
            <p>{{ session.username }}</p>
          </div>

          <div data-test="session-authenticated-field">
            <div class="item-title">Authenticated:</div>
            <v-tooltip location="bottom">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" :color="authenticatedTooltipAttrs.color" :icon="authenticatedTooltipAttrs.icon" />
              </template>
              <span>{{ authenticatedTooltipAttrs.text }}</span>
            </v-tooltip>
          </div>
        </v-col>

        <v-col cols="12" md="6" class="my-0 py-0">
          <div data-test="session-ip-address-field">
            <div class="item-title">IP address:</div>
            <code class="bg-tabs pa-1">{{ session.ip_address }}</code>
          </div>

          <div data-test="session-started-at-field">
            <div class="item-title">Started:</div>
            <p>{{ formatFullDateTime(session.started_at) }}</p>
          </div>

          <div data-test="session-last-seen-field">
            <div class="item-title">Last seen:</div>
            <p>{{ formatFullDateTime(session.last_seen) }}</p>
          </div>
        </v-col>
      </v-row>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useRoute } from "vue-router";
import { formatFullDateTime, getTimeFromNow } from "@/utils/date";
import hasPermission from "@/utils/permission";
import SessionDelete from "@/components/Sessions/SessionDelete.vue";
import SessionClose from "@/components/Sessions/SessionClose.vue";
import SessionPlay from "@/components/Sessions/SessionPlay.vue";
import DeviceLink from "@/components/Devices/DeviceLink.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useSessionsStore from "@/store/modules/sessions";

const sessionsStore = useSessionsStore();
const route = useRoute();
const snackbar = useSnackbar();
const sessionId = computed(() => route.params.id as string);
const session = computed(() => sessionsStore.session);
const canRemoveSessionRecord = hasPermission("session:removeRecord");
const authenticatedTooltipAttrs = computed(() => session.value.authenticated
  ? { color: "success", icon: "mdi-shield-check", text: "User has been authenticated" }
  : { color: "error", icon: "mdi-shield-alert", text: "User has not been authenticated" });

const getSession = async () => {
  try {
    await sessionsStore.getSession(sessionId.value);
  } catch (error: unknown) {
    snackbar.showError("Failed to load session details.");
    handleError(error);
  }
};

onMounted(async () => {
  await getSession();
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
