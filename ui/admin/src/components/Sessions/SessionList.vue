<template>
  <div>
    <DataTable
      :headers
      :items="sessions"
      v-model:itemsPerPage="itemsPerPage"
      v-model:page="page"
      :itemsPerPageOptions="[10, 20, 50, 100]"
      :loading
      :totalCount="sessionCount"
      data-test="session-list"
    >
      <template v-slot:rows>
        <tr v-for="(session, index) in sessions" :key="index">
          <td>
            <v-icon v-if="session.active" color="success" icon="mdi-check-circle" />
            <v-tooltip anchor="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" icon="mdi-check-circle" />
              </template>
              <span>{{ getTimeFromNow(session.last_seen) }}</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip>
              {{ session.uid ? displayOnlyTenCharacters(session.uid) : "—" }}
              <v-tooltip activator="parent" anchor="bottom">{{
                session.uid
              }}</v-tooltip>
            </v-chip>
          </td>
          <td>
            <span
              @click="session.device?.uid && redirectToDevice(session.device.uid)"
              @keypress.enter="session.device?.uid && redirectToDevice(session.device.uid)"
              tabindex="0"
              class="hover"
            >
              {{ session.device?.name || "Unknown device" }}
            </span>
          </td>
          <td>
            {{ session.username }}
          </td>
          <td>
            <v-tooltip anchor="bottom" v-if="session.authenticated">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" icon="mdi-shield-check" />
              </template>
              <span>User has been authenticated</span>
            </v-tooltip>
            <v-tooltip anchor="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="error" icon="mdi-shield-alert" />
              </template>
              <span>User has not been authenticated</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip density="compact" label>{{ session.ip_address }}</v-chip>
          </td>
          <td>
            <span>{{ formatFullDateTime(session.started_at) }}</span>
          </td>
          <td>
            {{ formatFullDateTime(session.last_seen) }}
          </td>
          <td>
            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="session.uid && goToSession(session.uid)"
                  @keypress.enter="session.tenant_id && goToSession(session.tenant_id)"
                  tabindex="0"
                  icon="mdi-information"
                />
              </template>
              <span>Details</span>
            </v-tooltip>
          </td>
        </tr>
      </template>
    </DataTable>
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

const redirectToDevice = (deviceId: string) => {
  router.push({ name: "deviceDetails", params: { id: deviceId } });
};

const goToSession = (sessionId: string) => {
  router.push({ name: "sessionDetails", params: { id: sessionId } });
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
