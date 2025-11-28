<template>
  <div class="d-flex pa-0 align-center">
    <h1>Namespace Details</h1>
  </div>
  <v-card
    v-if="namespace.tenant_id"
    class="mt-2 border rounded bg-background"
    elevation="0"
  >
    <v-card-title
      class="pa-4 d-flex align-center justify-space-between bg-v-theme-surface"
    >
      <span class="text-h6 ml-2 d-flex align-center ga-2">
        {{ namespace.name || 'Namespace' }}
        <v-chip
          v-if="namespace.type"
          :color="namespace.type === 'team' ? 'primary' : 'default'"
          data-test="namespace-type-chip"
          class="text-capitalize"
        >
          {{ namespace.type }}
        </v-chip>
      </span>

      <v-menu
        location="bottom"
        scrim
        eager
      >
        <template #activator="{ props }">
          <v-btn
            v-bind="props"
            variant="plain"
            class="border rounded bg-v-theme-background"
            density="comfortable"
            size="default"
            icon="mdi-format-list-bulleted"
            data-test="namespace-actions-menu-btn"
          />
        </template>

        <v-list
          class="bg-v-theme-surface"
          lines="two"
          density="compact"
        >
          <v-list-item
            data-test="namespace-edit-btn"
            @click="namespaceEdit = true"
          >
            <div class="d-flex align-center">
              <v-icon class="mr-2">mdi-pencil</v-icon>
              <v-list-item-title>Edit namespace</v-list-item-title>
            </div>
          </v-list-item>

          <v-list-item
            data-test="namespace-delete-btn"
            @click="namespaceDelete = true"
          >
            <div class="d-flex align-center">
              <v-icon class="mr-2">mdi-delete</v-icon>
              <v-list-item-title>Delete namespace</v-list-item-title>
            </div>
          </v-list-item>

          <div class="px-2 py-1" />
        </v-list>
      </v-menu>
    </v-card-title>

    <v-divider />

    <v-card-text class="pa-4 pt-0">
      <v-row class="py-3">
        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div data-test="namespace-name-field">
            <div class="item-title">Name:</div>
            <p class="text-truncate">{{ namespace.name }}</p>
          </div>

          <div data-test="namespace-tenant-id-field">
            <div class="item-title">Tenant ID:</div>
            <p class="text-truncate">
              <code>{{ namespace.tenant_id }}</code>
            </p>
          </div>

          <div data-test="namespace-owner-field">
            <div class="item-title">Owner:</div>
            <router-link
              :to="{ name: 'userDetails', params: { id: namespace.owner } }"
              class="unstyled-link text-decoration-underline cursor-pointer"
            >
              {{ getOwnerLabel(namespace) }}
            </router-link>
          </div>

          <div data-test="namespace-devices-field">
            <div class="item-title">Total Devices:</div>
            <p>{{ sumDevicesCount(namespace) }}</p>
          </div>

          <div
            class="d-flex align-center flex-wrap mt-2"
            data-test="namespace-devices-breakdown"
          >
            <div
              class="mr-6"
              data-test="namespace-devices-accepted"
            >
              <div class="item-title">Accepted:</div>
              <p>{{ namespace.devices_accepted_count || 0 }}</p>
            </div>
            <div
              class="mr-6"
              data-test="namespace-devices-pending"
            >
              <div class="item-title">Pending:</div>
              <p>{{ namespace.devices_pending_count || 0 }}</p>
            </div>
            <div data-test="namespace-devices-rejected">
              <div class="item-title">Rejected:</div>
              <p>{{ namespace.devices_rejected_count || 0 }}</p>
            </div>
          </div>
        </v-col>

        <v-col
          cols="12"
          md="6"
          class="my-0 py-0"
        >
          <div
            v-if="namespace.created_at"
            data-test="namespace-created-field"
          >
            <div class="item-title">Created:</div>
            <p>{{ formatFullDateTime(namespace.created_at) }}</p>
          </div>

          <div data-test="namespace-max-devices-field">
            <div class="item-title">Max Devices:</div>
            <p>{{ namespace.max_devices === -1 ? 'Unlimited' : namespace.max_devices }}</p>
          </div>

          <div
            v-if="namespace.settings"
            data-test="namespace-session-record-field"
          >
            <div class="item-title">Session Record:</div>
            <v-chip
              size="small"
              :color="namespace.settings.session_record ? 'success' : 'default'"
            >
              {{ namespace.settings.session_record ? 'Enabled' : 'Disabled' }}
            </v-chip>
          </div>

          <div
            v-if="namespace.settings?.connection_announcement"
            data-test="namespace-connection-announcement-field"
          >
            <div class="item-title">Connection Announcement:</div>
            <p>{{ namespace.settings.connection_announcement }}</p>
          </div>
        </v-col>
      </v-row>

      <v-divider class="my-4" />

      <div data-test="namespace-members-section">
        <h3 class="text-h6 mb-3">
          Members ({{ namespace.members?.length || 0 }})
        </h3>

        <v-list
          v-if="namespace.members?.length"
          class="bg-transparent"
          data-test="namespace-members-list"
        >
          <v-list-item
            v-for="(member, index) in namespace.members"
            :key="index"
            class="border rounded mb-2 bg-v-theme-surface"
            data-test="namespace-member-item"
          >
            <v-list-item-title class="d-flex align-center mb-2">
              <router-link
                :to="{ name: 'userDetails', params: { id: namespace.owner } }"
                class="unstyled-link text-decoration-underline cursor-pointer"
              >
                {{ getOwnerLabel(namespace) }}
              </router-link>
              <v-chip
                size="small"
                class="ml-2 text-capitalize"
                data-test="namespace-member-role"
              >
                <v-icon
                  :icon="getRoleIcon(member.role)"
                  class="mr-1"
                  size="small"
                />
                {{ getRoleLabel(member.role) }}
              </v-chip>
              <v-chip
                size="small"
                class="ml-2 text-capitalize"
                data-test="namespace-member-status"
              >
                {{ member.status }}
              </v-chip>
            </v-list-item-title>

            <v-list-item-subtitle class="d-flex flex-column">
              <span
                class="text-caption"
                data-test="namespace-member-id"
              >
                <strong>ID:</strong> <code>{{ member.id }}</code>
              </span>
              <span
                v-if="member.added_at && member.added_at !== '0001-01-01T00:00:00Z'"
                class="text-caption"
                data-test="namespace-member-added"
              >
                <strong>Added:</strong> {{ formatFullDateTime(member.added_at) }}
              </span>
              <span
                v-if="member.expires_at && member.expires_at !== '0001-01-01T00:00:00Z'"
                class="text-caption"
                data-test="namespace-member-expires"
              >
                <strong>Expires:</strong> {{ formatFullDateTime(member.expires_at) }}
              </span>
            </v-list-item-subtitle>
          </v-list-item>
        </v-list>

        <p
          v-else
          class="text-disabled text-center py-4"
        >
          No members found
        </p>
      </div>
    </v-card-text>
  </v-card>

  <v-card
    v-else
    class="mt-2 pa-4 bg-v-theme-surface"
  >
    <p class="text-center">Something is wrong, try again!</p>
  </v-card>

  <NamespaceEdit
    v-if="hasNamespace"
    v-model="namespaceEdit"
    :namespace="namespace"
    @update="fetchNamespaceDetails"
  />

  <NamespaceDelete
    v-if="hasNamespace"
    v-model="namespaceDelete"
    :tenant="namespace.tenant_id"
    :name="namespace.name || 'Namespace'"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import useNamespacesStore from "@admin/store/modules/namespaces";
import NamespaceEdit from "@admin/components/Namespace/NamespaceEdit.vue";
import NamespaceDelete from "@admin/components/Namespace/NamespaceDelete.vue";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import { formatFullDateTime } from "@/utils/date";

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const route = useRoute();

const loading = ref(false);
const namespaceEdit = ref(false);
const namespaceDelete = ref(false);
const namespace = ref({} as IAdminNamespace);
const hasNamespace = computed(() => !!namespace.value.tenant_id);
const namespaceId = computed(() => route.params.id);

const fetchNamespaceDetails = async () => {
  try {
    loading.value = true;
    namespace.value = await namespacesStore.fetchNamespaceById(namespaceId.value as string);
  } catch (error) {
    snackbar.showError("Failed to fetch namespace details.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  await fetchNamespaceDetails();
});

const sumDevicesCount = (namespace: IAdminNamespace) => {
  const { devices_accepted_count: acceptedCount, devices_pending_count: pendingCount, devices_rejected_count: rejectedCount } = namespace;
  return (acceptedCount + pendingCount + rejectedCount) || 0;
};

const roleConfig: Record<string, { label: string; icon: string; color: string }> = {
  owner: { label: "owner", icon: "mdi-crown", color: "warning" },
  administrator: { label: "admin", icon: "mdi-shield-account", color: "error" },
  operator: { label: "operator", icon: "mdi-account-cog", color: "primary" },
  observer: { label: "observer", icon: "mdi-eye", color: "info" },
};

const getRoleLabel = (role: string) => {
  if (!role) return "member";
  return roleConfig[role]?.label || role;
};

const getRoleIcon = (role: string) => {
  if (!role) return "mdi-account";
  return roleConfig[role]?.icon || "mdi-account";
};

const getOwnerLabel = (namespace: IAdminNamespace) => {
  const owner = namespace.members?.find(
    (member) => member.id === namespace.owner,
  );

  return owner?.email || namespace.owner || "";
};

defineExpose({ namespace });
</script>

<style scoped>
.unstyled-link {
  all: unset;
}
.unstyled-link:focus {
  outline: none;
}
</style>
