<template>
  <PageHeader
    icon="mdi-home"
    :title="hasNamespace ? namespace.name : 'No Active Namespace'"
    overline="Home"
    :description="activeNamespaceDescription"
    icon-color="primary"
    class="mb-6"
  >
    <template #actions>
      <div v-if="hasNamespace">
        <v-btn
          to="/settings/namespace"
          color="primary"
          variant="elevated"
          text="Settings"
          data-test="namespace-settings-btn"
        />
      </div>
      <div v-else>
        <div class="text-overline text-medium-emphasis mb-2">Create your first namespace</div>
        <v-btn
          color="primary"
          variant="elevated"
          prepend-icon="mdi-plus"
          data-test="create-namespace-home-btn"
          text="Create Namespace"
          class="mb-2"
          @click="showNamespaceAdd = true"
        />
        <div class="text-caption text-medium-emphasis">
          You need to create or join a namespace to start managing your devices and remote connections.
        </div>
      </div>
    </template>
  </PageHeader>

  <div v-if="hasNamespace">
    <v-row>
      <v-col
        cols="12"
        class="d-flex align-center mb-2"
      >
        <v-icon
          class="mr-2"
          icon="mdi-developer-board"
        />
        <h2 class="text-h6">Devices</h2>
      </v-col>
    </v-row>
    <v-row>
      <v-col
        cols="12"
        md="3"
      >
        <StatCard
          title="Accepted Devices"
          :stat="totalDevices"
          icon="mdi-check"
          button-label="View all devices"
          path="/devices"
          data-test="accepted-devices-card"
        />
      </v-col>
      <v-col
        cols="12"
        md="3"
      >
        <StatCard
          title="Online Devices"
          :stat="onlineDevices"
          icon="mdi-lan-connect"
          button-label="View Online Devices"
          path="/devices"
          data-test="online-devices-card"
        />
      </v-col>
      <v-col
        cols="12"
        md="3"
      >
        <StatCard
          title="Pending Devices"
          :stat="pendingDevices"
          icon="mdi-clock-outline"
          button-label="Approve Devices"
          path="/devices/pending"
          data-test="pending-devices-card"
        />
      </v-col>
      <v-col
        cols="12"
        md="3"
      >
        <v-card class="pa-6 bg-transparent text-center h-100 border border-dashed">
          <v-avatar
            color="surface-variant"
            size="64"
            class="mb-4"
            theme="dark"
          >
            <v-icon
              size="40"
              color="primary"
              icon="mdi-developer-board"
            />
          </v-avatar>
          <v-card-title class="text-h6 font-weight-bold mb-2">Add a new device</v-card-title>
          <v-card-subtitle class="text-body-2 text-medium-emphasis mb-4 text-wrap">
            Register new devices to this namespace and start managing remote connections
          </v-card-subtitle>
          <DeviceAdd />
        </v-card>
      </v-col>
    </v-row>
  </div>

  <NamespaceAdd v-model="showNamespaceAdd" />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import useNamespacesStore from "@/store/modules/namespaces";
import useDevicesStore from "@/store/modules/devices";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";
import StatCard from "@/components/StatCard.vue";
import NamespaceAdd from "@/components/Namespace/NamespaceAdd.vue";
import PageHeader from "@/components/PageHeader.vue";

const namespacesStore = useNamespacesStore();
const devicesStore = useDevicesStore();

const totalDevices = computed(() => devicesStore.totalDevicesCount);
const onlineDevices = computed(() => devicesStore.onlineDevicesCount);
const pendingDevices = computed(() => devicesStore.pendingDevicesCount);

const namespace = computed(() => namespacesStore.currentNamespace);
const hasNamespace = computed(() => namespacesStore.namespaceList.length !== 0);
const showNamespaceAdd = ref(false);

const activeNamespaceDescription = computed(() => (
  hasNamespace.value
    ? "This is your active namespace. All devices, sessions and configurations are isolated within this namespace."
    : `A namespace is a logical grouping that isolates your devices, sessions, and configurations from others.
          Each namespace has its own unique Tenant ID used to register devices.
          You can create multiple namespaces to organize different projects, teams, or environments.`
));
</script>
