<template>
  <v-icon
    color="primary"
    aria-label="Open devices menu"
    icon="mdi-developer-board"
    data-test="devices-icon"
    class="ml-3 mr-2"
    @click="toggleDrawer"
  />

  <Teleport to="body">
    <v-navigation-drawer
      v-model="isDrawerOpen"
      location="right"
      temporary
      :width="thresholds.sm"
      class="bg-v-theme-surface"
      data-test="devices-drawer"
    >
      <v-card
        class="bg-v-theme-surface h-100"
        flat
        data-test="devices-card"
      >
        <v-card-title class="text-h6 py-3">
          Device Management
        </v-card-title>
        <v-card-text class="pa-4 pt-0">
          <v-row
            dense
            class="mb-4"
          >
            <v-col
              cols="6"
              sm="3"
            >
              <v-card
                class="pa-3 text-center"
                variant="tonal"
                data-test="total-devices-card"
              >
                <div class="text-h4 font-weight-bold">
                  {{ stats.registered_devices }}
                </div>
                <div class="text-caption text-medium-emphasis">
                  Total
                </div>
              </v-card>
            </v-col>
            <v-col
              cols="6"
              sm="3"
            >
              <v-card
                class="pa-3 text-center"
                variant="tonal"
                data-test="online-devices-card"
              >
                <div class="text-h4 font-weight-bold">
                  {{ stats.online_devices }}
                </div>
                <div class="text-caption text-medium-emphasis">
                  Online
                </div>
              </v-card>
            </v-col>
            <v-col
              cols="6"
              sm="3"
            >
              <v-card
                class="pa-3 text-center"
                variant="tonal"
                data-test="pending-devices-card"
              >
                <div class="text-h4 font-weight-bold">
                  {{ stats.pending_devices }}
                </div>
                <div class="text-caption text-medium-emphasis">
                  Pending
                </div>
              </v-card>
            </v-col>
            <v-col
              cols="6"
              sm="3"
            >
              <v-card
                class="pa-3 text-center"
                variant="tonal"
                data-test="offline-devices-card"
              >
                <div class="text-h4 font-weight-bold">
                  {{ offlineDevices }}
                </div>
                <div class="text-caption text-medium-emphasis">
                  Offline
                </div>
              </v-card>
            </v-col>
          </v-row>
          <v-btn-toggle
            v-model="activeTab"
            mandatory
            color="primary"
            variant="outlined"
            divided
            class="mb-3 w-100"
            data-test="tab-toggle"
          >
            <v-btn
              value="pending"
              data-test="pending-tab"
              class="w-50"
            >
              <v-icon
                icon="mdi-clock-alert"
                :size="smAndUp ? 'small' : 'large'"
                class="mr-2"
              />
              <span v-if="smAndUp">Pending Approval</span>
              <v-chip
                v-if="stats.pending_devices > 0"
                color="warning"
                size="x-small"
                class="ml-2"
              >
                {{ stats.pending_devices }}
              </v-chip>
            </v-btn>
            <v-btn
              value="recent"
              data-test="recent-tab"
              class="w-50"
            >
              <v-icon
                icon="mdi-history"
                :size="smAndUp ? 'small' : 'large'"
                class="mr-2"
              />
              <span v-if="smAndUp">Recent Activity</span>
            </v-btn>
          </v-btn-toggle>
          <v-window
            v-model="activeTab"
            class="overflow-visible"
          >
            <v-window-item value="pending">
              <v-card
                variant="text"
                class="overflow-y-auto border"
              >
                <v-list
                  v-if="pendingDevicesList.length > 0"
                  density="compact"
                  class="bg-v-theme-surface pa-0"
                >
                  <template
                    v-for="(device, index) in pendingDevicesList"
                    :key="device.uid"
                  >
                    <v-divider v-if="index > 0" />
                    <v-list-item
                      class="px-3 py-3"
                      data-test="pending-device-item"
                    >
                      <template #prepend>
                        <v-icon
                          icon="mdi-devices"
                          color="primary"
                          size="small"
                          class="mr-n3 ml-1"
                        />
                      </template>
                      <v-list-item-title class="text-body-2 font-weight-medium mb-1">
                        {{ device.name }}
                      </v-list-item-title>
                      <v-list-item-subtitle class="text-caption">
                        <span class="font-mono">{{ device.identity?.mac || device.uid }}</span>
                        <span class="mx-1">•</span>
                        <span>{{ device.remote_addr }}</span>
                      </v-list-item-subtitle>
                      <template
                        v-if="smAndUp"
                        #append
                      >
                        <span class="text-caption text-medium-emphasis">
                          {{ formatTimeAgo(device.status_updated_at) }}
                        </span>
                      </template>
                      <div class="d-flex align-center ga-2 mt-1">
                        <v-btn
                          color="success"
                          variant="flat"
                          size="small"
                          prepend-icon="mdi-check-circle"
                          :data-test="`accept-${device.uid}`"
                          @click="handleAccept(device.uid)"
                        >
                          Accept
                        </v-btn>
                        <v-btn
                          color="error"
                          variant="tonal"
                          size="small"
                          prepend-icon="mdi-cancel"
                          :data-test="`reject-${device.uid}`"
                          @click="handleReject(device.uid)"
                        >
                          Reject
                        </v-btn>
                        <v-btn
                          icon="mdi-dots-vertical"
                          variant="text"
                          size="small"
                          :active="false"
                          :to="`/devices/${device.uid}`"
                        />
                      </div>
                    </v-list-item>
                  </template>
                </v-list>
                <div
                  v-else
                  class="pa-8 text-center"
                >
                  <v-icon
                    icon="mdi-check-circle"
                    size="64"
                    color="success"
                    class="opacity-50 mb-3"
                  />
                  <p class="text-body-2 text-medium-emphasis">
                    No pending devices
                  </p>
                  <p class="text-caption text-disabled mt-1">
                    All devices have been approved
                  </p>
                </div>
              </v-card>
            </v-window-item>
            <v-window-item value="recent">
              <v-card
                variant="text"
                class="overflow-y-auto border"
              >
                <v-list
                  v-if="recentDevicesList.length > 0"
                  density="compact"
                  class="pa-0"
                >
                  <template
                    v-for="(device, index) in recentDevicesList"
                    :key="device.uid"
                  >
                    <v-divider v-if="index > 0" />
                    <v-list-item
                      class="px-3 py-2"
                      :to="`/devices/${device.uid}`"
                    >
                      <template #prepend>
                        <v-badge
                          :color="device.online ? 'success' : 'grey'"
                          dot
                          inline
                          class="mr-2"
                        />
                      </template>
                      <v-list-item-title class="text-body-2 font-weight-medium">
                        {{ device.name }}
                      </v-list-item-title>
                      <v-list-item-subtitle class="text-caption font-mono">
                        {{ device.identity?.mac || device.uid }}
                      </v-list-item-subtitle>
                      <template #append>
                        <span class="text-caption text-medium-emphasis">
                          {{ device.online ? 'Active now' : formatTimeAgo(device.last_seen) }}
                        </span>
                      </template>
                    </v-list-item>
                  </template>
                </v-list>
                <div
                  v-else
                  class="pa-8 text-center"
                >
                  <v-icon
                    icon="mdi-history"
                    size="64"
                    color="primary"
                    class="opacity-50 mb-3"
                  />
                  <p class="text-body-2 text-medium-emphasis">
                    No recent activity
                  </p>
                </div>
              </v-card>
            </v-window-item>
          </v-window>
        </v-card-text>
        <v-divider />
        <v-card-actions class="pa-3">
          <v-btn
            to="/devices"
            variant="text"
            color="primary"
            block
            size="small"
            append-icon="mdi-arrow-right"
            data-test="view-all-devices-btn"
            text="View all devices"
            :active="false"
          />
        </v-card-actions>
      </v-card>
    </v-navigation-drawer>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeMount, ref } from "vue";
import { useDisplay } from "vuetify";
import useStatsStore from "@/store/modules/stats";
import useDevicesStore from "@/store/modules/devices";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import moment from "moment";
import { IDevice } from "@/interfaces/IDevice";

const { smAndUp, thresholds } = useDisplay();
const statsStore = useStatsStore();
const devicesStore = useDevicesStore();
const snackbar = useSnackbar();

const isDrawerOpen = ref(false);
const activeTab = ref<"pending" | "recent">("pending");
const pendingDevicesList = ref<IDevice[]>([]);
const recentDevicesList = ref<IDevice[]>([]);
const stats = computed(() => statsStore.stats);
const offlineDevices = computed(() => stats.value.registered_devices - stats.value.online_devices);
const toggleDrawer = () => { isDrawerOpen.value = !isDrawerOpen.value; };

const formatTimeAgo = (date: string | Date) => {
  if (!date) return "Unknown";
  return moment(date).fromNow();
};

const handleAccept = async (uid: string) => {
  try {
    await devicesStore.acceptDevice(uid);
    await fetchStats();
    await fetchPendingDevices();
    snackbar.showSuccess("Device accepted successfully");
  } catch (error: unknown) {
    snackbar.showError("Failed to accept device");
    handleError(error);
  }
};

const handleReject = async (uid: string) => {
  try {
    await devicesStore.rejectDevice(uid);
    await fetchStats();
    await fetchPendingDevices();
    snackbar.showSuccess("Device rejected successfully");
  } catch (error: unknown) {
    snackbar.showError("Failed to reject device");
    handleError(error);
  }
};

const fetchStats = async () => {
  try {
    await statsStore.fetchStats();
  } catch (error: unknown) {
    snackbar.showError("Failed to load device statistics");
    handleError(error);
  }
};

const fetchPendingDevices = async () => {
  try {
    await devicesStore.fetchDeviceList({ status: "pending", perPage: 100 });
    pendingDevicesList.value = [...devicesStore.devices];
  } catch (error: unknown) {
    handleError(error);
  }
};

const fetchRecentDevices = async () => {
  try {
    await devicesStore.fetchDeviceList({ status: "accepted" });
    recentDevicesList.value = [...devicesStore.devices]
      .sort((a, b) => new Date(b.last_seen).getTime() - new Date(a.last_seen).getTime());
  } catch (error: unknown) {
    handleError(error);
  }
};

onBeforeMount(async () => {
  await fetchStats();
  await fetchPendingDevices();
  await fetchRecentDevices();
});

defineExpose({
  toggleDrawer,
  formatTimeAgo,
  isDrawerOpen,
  handleAccept,
  handleReject,
  activeTab,
  pendingDevicesList,
  recentDevicesList,
  stats,
  offlineDevices,
});
</script>
