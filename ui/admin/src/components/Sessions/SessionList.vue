<template>
  <div>
    <DataTable
      v-if="loading || sessions.length"
      v-model:items-per-page="itemsPerPage"
      v-model:page="page"
      :headers
      :items="sessions"
      :items-per-page-options="[10, 20, 50, 100]"
      :loading
      :total-count="sessionCount"
      table-name="adminSessions"
      data-test="session-list"
    >
      <template #rows>
        <tr
          v-for="(session, index) in sessions"
          :key="index"
        >
          <td>
            <v-icon
              v-if="session.active"
              color="success"
              icon="mdi-check-circle"
            />
            <v-tooltip
              v-else
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  v-bind="props"
                  icon="mdi-check-circle"
                />
              </template>
              <span>{{ getTimeFromNow(session.last_seen) }}</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip>
              {{ session.uid ? displayOnlyTenCharacters(session.uid) : "—" }}
              <v-tooltip
                activator="parent"
                anchor="bottom"
              >
                {{
                  session.uid
                }}
              </v-tooltip>
            </v-chip>
          </td>
          <td>
            <span
              tabindex="0"
              class="hover"
              data-test="device-link"
              @click="session.device?.uid && redirectToDevice(session.device.uid)"
              @keypress.enter="session.device?.uid && redirectToDevice(session.device.uid)"
            >
              {{ session.device?.name || "Unknown device" }}
            </span>
          </td>
          <td>
            {{ session.username }}
          </td>
          <td>
            <v-tooltip
              v-if="session.authenticated"
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  v-bind="props"
                  icon="mdi-shield-check"
                />
              </template>
              <span>User has been authenticated</span>
            </v-tooltip>
            <v-tooltip
              v-else
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  v-bind="props"
                  color="error"
                  icon="mdi-shield-alert"
                />
              </template>
              <span>User has not been authenticated</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip
              density="compact"
              label
            >
              {{ session.ip_address }}
            </v-chip>
          </td>
          <td>
            <span>{{ formatFullDateTime(session.started_at) }}</span>
          </td>
          <td>
            {{ formatFullDateTime(session.last_seen) }}
          </td>
          <td>
            <v-tooltip
              bottom
              anchor="bottom"
            >
              <template #activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  tabindex="0"
                  icon="mdi-information"
                  data-test="info-button"
                  @click="session.uid && goToSession(session.uid)"
                  @keypress.enter="session.tenant_id && goToSession(session.tenant_id)"
                />
              </template>
              <span>Details</span>
            </v-tooltip>
          </td>
        </tr>
      </template>
    </DataTable>

    <NoItemsMessage
      v-else
      class="mt-2"
      item="Sessions"
      icon="mdi-console"
      data-test="sessions-empty-state"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useSessionsStore from "@admin/store/modules/sessions";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import { getTimeFromNow, formatFullDateTime } from "@/utils/date";
import { displayOnlyTenCharacters } from "@/utils/string";
import handleError from "@/utils/handleError";
import NoItemsMessage from "@/components/NoItemsMessage.vue";

const router = useRouter();
const snackbar = useSnackbar();
const sessionStore = useSessionsStore();
const sessions = computed(() => sessionStore.sessions);
const sessionCount = computed(() => sessionStore.sessionCount);
const itemsPerPage = ref(10);
const loading = ref(false);
const page = ref(1);
const headers = ref([
  {
    text: "Active",
    value: "active",
  },
  {
    text: "Id",
    value: "uid",
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
    value: "started_at",
  },
  {
    text: "Last Seen",
    value: "last_seen",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const fetchSessions = async () => {
  try {
    loading.value = true;
    await sessionStore.fetchSessionList({
      perPage: itemsPerPage.value,
      page: page.value,
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch sessions list.");
  }
  loading.value = false;
};

const redirectToDevice = async (deviceId: string) => {
  await router.push({ name: "deviceDetails", params: { id: deviceId } });
};

const goToSession = async (sessionId: string) => {
  await router.push({ name: "sessionDetails", params: { id: sessionId } });
};

watch([itemsPerPage, page], async () => {
  await fetchSessions();
});

onMounted(async () => {
  await fetchSessions();
});
</script>

<style scoped>
.hover {
  cursor: pointer;
  text-decoration: underline;
}
</style>
