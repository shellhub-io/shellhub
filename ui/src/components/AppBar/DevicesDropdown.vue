<template>
  <v-icon
    @click="toggleDrawer"
    color="primary"
    aria-label="Open devices menu"
    icon="mdi-developer-board"
    data-test="devices-icon"
    class="ml-3 mr-2"
  />

  <Teleport to="body">
    <v-navigation-drawer
      v-model="drawerOpen"
      location="right"
      temporary
      :width="drawerWidth"
      class="bg-v-theme-surface"
      data-test="devices-drawer"
    >
    <v-card
      class="bg-v-theme-surface h-100"
      flat
      data-test="devices-card"
    >
      <!-- Header with Stats -->
      <v-card-title class="text-h6 pb-3 border-b">
        Device Management
      </v-card-title>

      <v-card-text class="pa-4">
        <v-row dense class="mb-4">
          <v-col cols="3">
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

          <v-col cols="3">
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

          <v-col cols="3">
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

          <v-col cols="3">
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

        <!-- Tab Buttons -->
        <v-btn-toggle
          v-model="activeTab"
          mandatory
          color="primary"
          variant="outlined"
          divided
          class="mb-3 w-100"
          data-test="tab-toggle"
        >
          <v-btn value="pending" data-test="pending-tab" class="flex-1-1">
            <v-icon icon="mdi-clock-alert" size="small" class="mr-2" />
            Pending Approval
            <v-chip
              v-if="stats.pending_devices > 0"
              color="warning"
              size="x-small"
              class="ml-2"
            >
              {{ stats.pending_devices }}
            </v-chip>
          </v-btn>
          <v-btn value="recent" data-test="recent-tab" class="flex-1-1">
            <v-icon icon="mdi-history" size="small" class="mr-2" />
            Recent Activity
          </v-btn>
        </v-btn-toggle>

        <v-window v-model="activeTab">
          <!-- Pending Tab -->
          <v-window-item value="pending">
            <!-- Pending Devices List -->
            <v-card
              variant="text"
              class="overflow-y-auto border"
              :max-height="isMobile ? 'calc(100vh - 300px)' : '400px'"
            >
              <template v-if="pendingDevices.length > 0">
                <v-list density="compact" class="bg-v-theme-surface pa-0">
                  <template v-for="(device, index) in pendingDevices" :key="device.uid">
                    <v-divider v-if="index > 0" />
                    <v-list-item class="px-3 py-3">
                      <template v-slot:prepend>
                        <v-icon icon="mdi-devices" color="primary" size="small" />
                      </template>

                      <v-list-item-title class="text-body-2 font-weight-medium mb-1">
                        {{ device.name }}
                      </v-list-item-title>

                      <v-list-item-subtitle class="text-caption">
                        <span class="font-mono">{{ device.identity?.mac || device.uid }}</span>
                        <span class="mx-1">•</span>
                        <span>{{ getDeviceIP(device) }}</span>
                      </v-list-item-subtitle>

                      <template v-slot:append>
                        <span class="text-caption text-medium-emphasis">
                          {{ formatTimeAgo(device.status_updated_at) }}
                        </span>
                      </template>

                      <div class="d-flex gap-2 mt-2">
                        <v-btn
                          color="success"
                          variant="flat"
                          size="small"
                          prepend-icon="mdi-check-circle"
                          @click="handleAccept(device.uid)"
                          :data-test="`accept-${device.uid}`"
                        >
                          Accept
                        </v-btn>
                        <v-btn
                          color="error"
                          variant="tonal"
                          size="small"
                          @click="handleReject(device.uid)"
                          :data-test="`reject-${device.uid}`"
                        >
                          Reject
                        </v-btn>
                        <v-btn
                          icon="mdi-dots-vertical"
                          variant="text"
                          size="small"
                          :to="`/devices/${device.uid}`"
                        />
                      </div>
                    </v-list-item>
                  </template>
                </v-list>
              </template>

              <template v-else>
                <div class="pa-8 text-center">
                  <v-icon icon="mdi-check-circle" size="64" color="success" class="opacity-50 mb-3" />
                  <p class="text-body-2 text-medium-emphasis">No pending devices</p>
                  <p class="text-caption text-disabled mt-1">All devices have been approved</p>
                </div>
              </template>
            </v-card>
          </v-window-item>

          <!-- Recent Activity Tab -->
          <v-window-item value="recent">
            <v-card
              variant="outlined"
              class="overflow-y-auto"
              :max-height="isMobile ? 'calc(100vh - 400px)' : '400px'"
            >
              <v-list density="compact" class="pa-0">
                <template v-for="(device, index) in recentDevices" :key="device.uid">
                  <v-divider v-if="index > 0" />
                  <v-list-item class="px-3 py-2" :to="`/devices/${device.uid}`">
                    <template v-slot:prepend>
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

                    <template v-slot:append>
                      <span class="text-caption text-medium-emphasis">
                        {{ device.online ? 'Active now' : formatTimeAgo(device.last_seen) }}
                      </span>
                    </template>
                  </v-list-item>
                </template>
              </v-list>
            </v-card>
          </v-window-item>
        </v-window>
      </v-card-text>

      <!-- Footer -->
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
        >
          View All Devices
        </v-btn>
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

const { mobile, thresholds } = useDisplay();
const statsStore = useStatsStore();
const devicesStore = useDevicesStore();
const snackbar = useSnackbar();

const isMobile = computed(() => mobile.value);
const drawerWidth = computed(() => thresholds.value.sm);
const drawerOpen = ref(false);
const activeTab = ref("pending");
const pendingDevicesList = ref<any[]>([]);
const recentDevicesList = ref<any[]>([]);

const stats = computed(() => statsStore.stats);
const hasPendingDevices = computed(() => stats.value.pending_devices > 0);
const offlineDevices = computed(() =>
  stats.value.registered_devices - stats.value.online_devices
);

const pendingDevices = computed(() => pendingDevicesList.value);
const recentDevices = computed(() => recentDevicesList.value);

const toggleDrawer = () => {
  drawerOpen.value = !drawerOpen.value;
  console.log("Drawer toggled:", drawerOpen.value);
};

const getDeviceIP = (device: any) => {
  return device.info?.public_ip || device.info?.private_ip || "N/A";
};

const formatTimeAgo = (date: string | Date) => {
  try {
    return moment(date).fromNow();
  } catch {
    return "Unknown";
  }
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
    await devicesStore.fetchDeviceList({ status: "pending", page: 1, perPage: 100 });
    pendingDevicesList.value = [...devicesStore.devices];
  } catch (error: unknown) {
    handleError(error);
  }
};

const fetchRecentDevices = async () => {
  try {
    await devicesStore.fetchDeviceList({ status: "accepted", page: 1, perPage: 10 });
    recentDevicesList.value = [...devicesStore.devices]
      .sort((a, b) => new Date(b.last_seen).getTime() - new Date(a.last_seen).getTime())
      .slice(0, 10);
  } catch (error: unknown) {
    handleError(error);
  }
};

onBeforeMount(async () => {
  await fetchStats();
  await fetchPendingDevices();
  await fetchRecentDevices();
});

defineExpose({ fetchStats, fetchPendingDevices, drawerOpen });
</script>

<style scoped>
.v-window {
  overflow: visible;
}
</style>