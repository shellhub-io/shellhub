<template>
  <div v-if="!hasError">
    <v-card
      class="bg-transparent mb-6"
      elevation="0"
      rounded="0"
    >
      <v-row>
        <v-col
          cols="12"
          md="6"
        >
          <div class="d-flex align-start">
            <v-avatar
              color="primary"
              size="48"
              class="mr-4"
            >
              <v-icon
                size="32"
                icon="mdi-home"
              />
            </v-avatar>
            <div>
              <div class="text-overline text-medium-emphasis mb-1">Home</div>
              <div class="text-h5 font-weight-bold mb-2">
                {{ hasNamespace ? namespace.name : "No Active Namespace" }}
              </div>
              <div class="text-body-2 text-medium-emphasis">{{ activeNamespaceDescription }}</div>
            </div>
          </div>
        </v-col>
        <v-col
          cols="12"
          md="6"
        >
          <v-card
            class="pa-4"
            variant="tonal"
          >
            <div v-if="hasNamespace">
              <div class="text-overline text-medium-emphasis mb-2">TENANT ID</div>
              <div class="d-flex align-center justify-space-between">
                <code
                  class="text-primary"
                  data-test="tenant-info-text"
                >{{ namespace.tenant_id }}</code>
                <CopyWarning copied-item="Tenant ID">
                  <template #default="{ copyText }">
                    <v-btn
                      data-test="copy-tenant-btn"
                      color="primary"
                      variant="elevated"
                      size="small"
                      prepend-icon="mdi-content-copy"
                      @click="copyText(namespace.tenant_id)"
                    >
                      Copy
                    </v-btn>
                  </template>
                </CopyWarning>
              </div>
              <div class="text-caption text-medium-emphasis mt-2">Use this ID to register new devices to this namespace</div>
            </div>
            <div v-else>
              <div class="text-overline text-medium-emphasis">Create your first namespace</div>
              <div class="my-2">
                <v-btn
                  color="primary"
                  variant="elevated"
                  size="small"
                  prepend-icon="mdi-plus"
                  data-test="create-namespace-home-btn"
                  text="Create Namespace"
                  @click="showNamespaceAdd = true"
                />
              </div>
              <div class="text-caption text-medium-emphasis">
                You need to create or join a namespace to start managing your devices and remote connections.
              </div>
            </div>
          </v-card>
        </v-col>
      </v-row>
    </v-card>

    <div v-if="hasNamespace">
      <v-row>
        <v-col
          cols="12"
          class="d-flex align-center mb-2"
        >
          <v-icon
            class="mr-2"
            icon="mdi-devices"
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
  </div>
  <v-card
    v-else
    data-test="home-failed"
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">
      Something is wrong, try again!
    </p>
  </v-card>

  <NamespaceAdd v-model="showNamespaceAdd" />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import useNamespacesStore from "@/store/modules/namespaces";
import useDevicesStore from "@/store/modules/devices";
import DeviceAdd from "@/components/Devices/DeviceAdd.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";
import StatCard from "@/components/StatCard.vue";
import NamespaceAdd from "@/components/Namespace/NamespaceAdd.vue";

const namespacesStore = useNamespacesStore();
const devicesStore = useDevicesStore();
const hasError = ref(false);

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

defineExpose({ hasError });
</script>
