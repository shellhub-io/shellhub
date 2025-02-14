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
              <span>{{ lastSeen(session.last_seen) }}</span>
            </v-tooltip>
          </td>
          <td>
            <v-chip>
              {{ displayOnlyTenCharacters(session.uid) }}
              <v-tooltip activator="parent" anchor="bottom">{{
                session.uid
              }}</v-tooltip>
            </v-chip>
          </td>
          <td>
            <span
              @click="redirectToDevice(session.device.uid)"
              @keypress.enter="redirectToDevice(session.device.uid)"
              tabindex="0"
              class="hover"
            >
              {{ session.device.name }}
            </span>
          </td>
          <td>
            {{ session.username }}
          </td>
          <td class="d-flex justify-center align-center">
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
            <span>{{ formatDate(session.started_at) }}</span>
          </td>
          <td>
            {{ formatDate(session.last_seen) }}
          </td>
          <td>
            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="goToSession(session.uid)"
                  @keypress.enter="goToSession(session.tenant_id)"
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

<script lang="ts">
import { computed, defineComponent, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useStore } from "../../store";
import DataTable from "../DataTable.vue";
import { lastSeen, formatDate } from "../../hooks/formateDate";
import displayOnlyTenCharacters from "../../hooks/string";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const router = useRouter();
    const itemsPerPage = ref(10);
    const loading = ref(false);
    const page = ref(1);

    const getSessions = async (perPagaeValue: number, pageValue: number) => {
      try {
        loading.value = true;

        const hasSessions = await store.dispatch("sessions/fetch", {
          perPage: perPagaeValue,
          page: pageValue,
        });

        if (!hasSessions) {
          page.value--;
        }

        loading.value = false;
      } catch (error) {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.sessionList);
      }
    };

    onMounted(async () => {
      try {
        loading.value = true;
        getSessions(itemsPerPage.value, page.value);
      } catch (error) {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.sessionList);
      } finally {
        loading.value = false;
      }
    });

    const next = async () => {
      await getSessions(itemsPerPage.value, ++page.value);
    };

    const prev = async () => {
      try {
        if (page.value > 1) await getSessions(itemsPerPage.value, --page.value);
      } catch (error) {
        store.dispatch("snackbar/setSnackbarErrorDefault");
      }
    };

    const changeItemsPerPage = async (newItemsPerPage: number) => {
      itemsPerPage.value = newItemsPerPage;
    };

    watch(itemsPerPage, async () => {
      await getSessions(itemsPerPage.value, page.value);
    });

    const sessions = computed(() => store.getters["sessions/sessions"]);
    const numberSessions = computed(() => store.getters["sessions/numberSessions"]);

    const redirectToDevice = (deviceId: string) => {
      router.push({ name: "deviceDetails", params: { id: deviceId } });
    };

    const goToSession = (sessionId: string) => {
      router.push({ name: "sessionDetails", params: { id: sessionId } });
    };

    return {
      headers: [
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
      ],
      itemsPerPage,
      loading,
      page,
      sessions,
      numberSessions,
      lastSeen,
      displayOnlyTenCharacters,
      redirectToDevice,
      formatDate,
      goToSession,
      changeItemsPerPage,
      next,
      prev,
    };
  },
  components: { DataTable },
});
</script>

<style scoped>
.hover {
  cursor: pointer;
  text-decoration: underline;
}
</style>
