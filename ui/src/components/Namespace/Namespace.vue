<template>
  <NamespaceAdd v-model="showAddDialog" />
  <NamespaceInstructions v-model="showInstructionsDialog" />

  <v-menu
    :close-on-content-click="false"
    scrim
    location="bottom center"
    :offset="4"
  >
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        variant="text"
        :class="['text-none p-2 border-e-thin', { 'border-s-thin': mdAndDown }]"
        density="comfortable"
        size="x-large"
        stacked
      >
        <div class="d-flex align-center ga-2">
          <AdminConsoleItem
            v-if="isAdminContext"
            compact
          />
          <template v-else>
            <NamespaceChip :name="currentNamespace.name" />
            <span class="text-body-1">{{ currentNamespace.name || 'No Namespace' }}</span>
          </template>
          <v-icon
            size="x-small"
            icon="mdi-chevron-down"
          />
        </div>
      </v-btn>
    </template>

    <v-card
      :width="thresholds.sm / 2"
      class="border"
    >
      <v-list class="bg-v-theme-surface">
        <AdminConsoleItem v-if="isAdminContext" />
        <template v-else-if="currentNamespace.tenant_id">
          <v-list-subheader>Active Namespace</v-list-subheader>
          <NamespaceListItem
            :namespace="currentNamespace"
            :active="true"
            :user-id="userId"
            @select="handleNamespaceSwitch"
          />
          <div class="px-4 pb-2 pt-3">
            <div class="text-caption text-grey mb-1">Tenant ID</div>
            <div class="d-flex align-center ga-2 pa-2 border-thin rounded text-caption">
              <span class="flex-1-1 text-truncate">{{ currentNamespace.tenant_id }}</span>
              <CopyWarning :copied-item="'Tenant ID'">
                <template #default="{ copyText }">
                  <v-icon
                    size="small"
                    class="cursor-pointer"
                    icon="mdi-content-copy"
                    @click="copyText(currentNamespace.tenant_id)"
                  />
                </template>
              </CopyWarning>
            </div>
          </div>
        </template>

        <template v-if="availableNamespaces.length > 0 || (hasNamespaces && showAdminButton)">
          <v-divider class="my-2" />
          <v-list-subheader>{{ isAdminContext ? 'Available Namespaces' : 'Switch Namespace' }}</v-list-subheader>

          <template
            v-for="(namespace, index) in availableNamespaces"
            :key="namespace.tenant_id"
          >
            <NamespaceListItem
              :namespace="namespace"
              :active="false"
              :user-id="userId"
              @select="handleNamespaceSwitch"
            />
            <v-divider v-if="index < availableNamespaces.length - 1 || (!isAdminContext && showAdminButton)" />
          </template>
        </template>
        <AdminConsoleItem
          v-if="showAdminButton && !isAdminContext"
          @click="navigateToAdminPanel"
        />

        <template v-if="!isAdminContext">
          <v-divider />
          <div class="px-4 pt-4 pb-2">
            <v-btn
              variant="flat"
              color="primary"
              prepend-icon="mdi-plus-circle"
              block
              text="Create Namespace"
              @click="showAddDialog = true"
            />
          </div>
        </template>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import axios from "axios";
import { useDisplay } from "vuetify";
import NamespaceAdd from "./NamespaceAdd.vue";
import NamespaceInstructions from "./NamespaceInstructions.vue";
import NamespaceChip from "./NamespaceChip.vue";
import NamespaceListItem from "./NamespaceListItem.vue";
import AdminConsoleItem from "./AdminConsoleItem.vue";
import useNamespacesStore from "@/store/modules/namespaces";
import useAuthStore from "@/store/modules/auth";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import CopyWarning from "@/components/User/CopyWarning.vue";
import { envVariables } from "@/envVariables";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps<{ isAdminContext?: boolean }>();

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const { mdAndDown, thresholds } = useDisplay();

const showAddDialog = ref(false);

const currentNamespace = computed(() => namespacesStore.currentNamespace);
const namespaceList = computed(() => namespacesStore.namespaceList);
const hasNamespaces = computed(() => namespacesStore.namespaceList.length > 0);
const showInstructionsDialog = ref(false);
const userId = computed(() => authStore.id || localStorage.getItem("id") || "");

const showAdminButton = computed(() => {
  if (props.isAdminContext) return true;
  return envVariables.isEnterprise && !envVariables.isCloud && Boolean(authStore.isAdmin);
});

const availableNamespaces = computed(() => {
  const namespaces = namespaceList.value.filter((ns) => ns.tenant_id !== currentNamespace.value.tenant_id);
  if (props.isAdminContext && currentNamespace.value.tenant_id) namespaces.push(currentNamespace.value);
  return namespaces;
});

const navigateToAdminPanel = () => { window.location.href = "/admin"; };

const handleNamespaceSwitch = async (tenantId: string) => {
  try {
    await namespacesStore.switchNamespace(tenantId);
    if (props.isAdminContext && tenantId) window.location.href = "/";
    else window.location.reload();
  } catch (error: unknown) {
    snackbar.showError("Failed to switch namespace");
    handleError(error);
  }
};

const loadCurrentNamespace = async () => {
  const currentTenantId = localStorage.getItem("tenant") || "";

  try {
    await namespacesStore.fetchNamespaceList({ perPage: 30 });
    await namespacesStore.fetchNamespace(currentTenantId);
  } catch (error: unknown) {
    if (!axios.isAxiosError(error)) {
      snackbar.showError("Failed to load namespace");
      handleError(error);
      return;
    }

    // Namespace not found, try to switch to first available
    if (error.response?.status === 404) {
      const firstNamespace = namespaceList.value[0];
      if (firstNamespace) await handleNamespaceSwitch(firstNamespace.tenant_id);
      return;
    }

    // Server error with no tenant - ignore
    if (error.response?.status === 500 && !currentTenantId) {
      return;
    }

    snackbar.showError("Failed to load namespace");
    handleError(error);
  }
};

onMounted(async () => {
  await loadCurrentNamespace();
  showInstructionsDialog.value = !hasNamespaces.value && !props.isAdminContext;
});
</script>
