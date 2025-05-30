<template>
  <div>
    <DataTable
      :headers="headers"
      :items="sessions"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :totalCount="numberSessions"
      :page="page"
      :actualPage="page"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
      data-test="session-list"
    >
      <template v-slot:rows>
        <tr v-for="(session, index) in sessions" :key="index">
          <td>
            <v-icon v-if="session.active" color="success">
              mdi-check-circle
            </v-icon>
            <v-tooltip anchor="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props"> mdi-check-circle </v-icon>
              </template>
              <span>{{ getTimeFromNow(session.last_seen) }}</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip>
              {{ session.uid ? displayOnlyTenCharacters(session.uid) : '—' }}
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
              {{ session.device?.name || 'Unknown device' }}
            </span>
          </td>
          <td>
            {{ session.username }}
          </td>
          <td class="text-center">
            <v-tooltip anchor="bottom" v-if="session.authenticated">
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props">mdi-shield-check </v-icon>
              </template>
              <span>User has been authenticated</span>
            </v-tooltip>
            <v-tooltip anchor="bottom" v-else>
              <template v-slot:activator="{ props }">
                <v-icon v-bind="props" color="error"> mdi-shield-alert </v-icon>
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
                >mdi-information
                </v-icon>
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
import DataTable from "../DataTable.vue";
import { getTimeFromNow, formatFullDateTime } from "../../hooks/date";
import displayOnlyTenCharacters from "../../hooks/string";

const router = useRouter();
const snackbar = useSnackbar();
const sessionStore = useSessionsStore();

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
const itemsPerPage = ref(10);
const loading = ref(false);
const page = ref(1);

const getSessions = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;

    const hasSessions = await sessionStore.fetch({
      perPage: perPageValue,
      page: pageValue,
    });

    if (!hasSessions) {
      page.value--;
    }

    loading.value = false;
  } catch (error) {
    snackbar.showError("Failed to fetch sessions list.");
  }
};

onMounted(async () => {
  try {
    loading.value = true;
    getSessions(itemsPerPage.value, page.value);
  } catch (error) {
    snackbar.showError("Failed to fetch sessions list.");
  } finally {
    loading.value = false;
  }
});

const next = async () => {
  await getSessions(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  if (page.value > 1) await getSessions(itemsPerPage.value, --page.value);
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, async () => {
  await getSessions(itemsPerPage.value, page.value);
});

const sessions = computed(() => sessionStore.getSessions);
const numberSessions = computed(() => sessionStore.getNumberSessions);

const redirectToDevice = (deviceId: string) => {
  router.push({ name: "deviceDetails", params: { id: deviceId } });
};

const goToSession = (sessionId: string) => {
  router.push({ name: "sessionDetails", params: { id: sessionId } });
};
</script>

<style scoped>
.hover {
  cursor: pointer;
  text-decoration: underline;
}
</style>
