<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:itemsPerPage="itemsPerPage"
      :headers
      :items="sessions"
      :totalCount="sessionCount"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      data-test="sessions-list"
    >
      <template v-slot:rows>
        <tr v-for="(session, index) in sessions" :key="index">
          <td class="text-center">
            <SessionPlay
              :authenticated="session.authenticated"
              :uid="session.uid"
              :recorded="session.recorded"
              data-test="session-play-component"
            />
          </td>

          <td class="text-center" v-if="session.device">
            <p
              @click="redirectDevice(session.device.uid)"
              @keyup="redirectDevice(session.device.uid)"
              tabindex="0"
              class="link"
            >
              <v-btn
                class="text-none border rounded bg-v-theme-background"
                prepend-icon="mdi-developer-board"
                variant="plain"
                density="default"
                size="default"
              >
                {{ session.device.name }}
              </v-btn>
            </p>
          </td>

          <td class="text-center">
            <v-tooltip location="bottom" v-if="!session.authenticated">
              <template v-slot:activator="{ props }">
                <span v-bind="props">{{ session.username }}</span>
              </template>
              <span v-if="!session.authenticated">Unauthorized</span>
            </v-tooltip>
            <span v-if="session.authenticated">{{ session.username }}</span>
          </td>

          <td class="text-center">
            <v-tooltip location="bottom" v-if="session.authenticated">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="success">
                  mdi-shield-check
                </v-icon>
              </template>
              <span>User has been authenticated</span>
            </v-tooltip>
            <v-tooltip bottom v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="error"> mdi-shield-alert </v-icon>
              </template>
              <span>User has not been authenticated</span>
            </v-tooltip>
          </td>

          <td class="text-center">
            <v-code class="bg-tabs">
              {{ session.ip_address }}
            </v-code>
          </td>
          <td class="text-center">
            <span>{{ formatShortDateTime(session.started_at) }}</span>
          </td>

          <td class="text-center">
            <span>{{ formatShortDateTime(session.last_seen) }}</span>
          </td>

          <td class="text-center">
            <v-menu location="bottom" scrim eager>
              <template v-slot:activator="{ props }">
                <v-btn
                  v-bind="props"
                  variant="plain"
                  class="border rounded bg-v-theme-background"
                  density="comfortable"
                  size="default"
                  icon="mdi-format-list-bulleted"
                  data-test="session-list-actions"
                />
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-list-item @click="redirectToSession(session.uid)">
                  <div class="d-flex align-center">
                    <div class="mr-2">
                      <v-icon> mdi-information </v-icon>
                    </div>

                    <v-list-item-title data-test="mdi-information-list-item">
                      Details
                    </v-list-item-title>
                  </div>
                </v-list-item>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveRecord()"
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
              </v-list>
            </v-menu>
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import axios, { AxiosError } from "axios";
import { useRouter } from "vue-router";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import { formatShortDateTime } from "@/utils/date";
import DataTable from "../DataTable.vue";
import SessionClose from "./SessionClose.vue";
import SessionPlay from "./SessionPlay.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useSessionsStore from "@/store/modules/sessions";

const headers = [
  {
    text: "Recorded",
    value: "recorded",
  },
  {
    text: "Device",
    value: "device",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Authenticated",
    value: "authenticated",
  },
  {
    text: "IP Address",
    value: "ip_address",
  },
  {
    text: "Started",
    value: "started",
  },
  {
    text: "Last Seen",
    value: "last_seen",
  },
  {
    text: "Actions",
    value: "actions",
  },
];
const authStore = useAuthStore();
const sessionsStore = useSessionsStore();
const router = useRouter();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const sessions = computed(() => sessionsStore.sessions);
const sessionCount = computed(
  () => sessionsStore.sessionCount,
);

const getSessions = async () => {
  try {
    loading.value = true;
    await sessionsStore.fetchSessionList({
      page: page.value,
      perPage: itemsPerPage.value,
    });
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You don't have permission to access this resource.");
      }
    } else {
      snackbar.showError("Failed to load the session list.");
      handleError(error);
    }
  } finally {
    loading.value = false;
  }
};

const hasAuthorizationRemoveRecord = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.session.removeRecord);
};

watch([page, itemsPerPage], async () => {
  await getSessions();
});

const redirectToSession = (sessionUid: string) => {
  router.push({ name: "SessionDetails", params: { id: sessionUid } });
};

const redirectDevice = (deviceUid: string) => {
  router.push({ name: "DeviceDetails", params: { identifier: deviceUid } });
};

const refreshSessions = async () => {
  await getSessions();
};

onMounted(async () => {
  await getSessions();
});
</script>

<style scoped>
.hover-text {
  cursor: pointer;
  animation: fadeIn 0.5s;
}

.hover-text:hover,
.hover-text:focus {
  text-decoration: underline;
}

.link {
  text-decoration: underline;
  cursor: pointer;
}
</style>
