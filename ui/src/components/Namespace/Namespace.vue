<template>
  <NamespaceAdd v-model="showAddDialog" />

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
        :class="['text-none p-2 border-e-thin', { 'border-s-thin': $vuetify.display.mdAndDown }]"
        density="comfortable"
        size="x-large"
        stacked
      >
        <div class="d-flex align-center ga-2">
          <NamespaceChip :name="currentNamespace.name" />
          <span class="text-body-1">{{ currentNamespace.name || 'No Namespace' }}</span>
          <v-icon size="x-small">
            mdi-chevron-down
          </v-icon>
        </div>
      </v-btn>
    </template>

    <v-card
      :width="$vuetify.display.thresholds.sm / 2"
      border
    >
      <v-list class="bg-v-theme-surface">
        <v-list-subheader>Active Namespace</v-list-subheader>

        <NamespaceListItem
          v-if="currentNamespace.tenant_id"
          :namespace="currentNamespace"
          :active="true"
          :user-id="userId"
          @select="handleNamespaceSwitch"
        />

        <div
          v-if="currentNamespace.tenant_id"
          class="px-4 pb-2 pt-3"
        >
          <div class="text-caption text-grey mb-1">
            Tenant ID
          </div>
          <div class="d-flex align-center ga-2 pa-2 border-thin rounded text-caption">
            <span class="flex-1-1 text-truncate">{{ currentNamespace.tenant_id }}</span>
            <CopyWarning :copied-item="'Tenant ID'">
              <template #default="{ copyText }">
                <v-icon
                  size="small"
                  class="cursor-pointer"
                  @click="copyText(currentNamespace.tenant_id)"
                >
                  mdi-content-copy
                </v-icon>
              </template>
            </CopyWarning>
          </div>
        </div>

        <template v-if="otherNamespaces.length > 0 || (hasNamespaces && showAdminPanel)">
          <v-divider class="my-2" />
          <v-list-subheader>Switch Namespace</v-list-subheader>

          <template
            v-for="ns in otherNamespaces"
            :key="ns.tenant_id"
          >
            <NamespaceListItem
              :namespace="ns"
              :active="false"
              :user-id="userId"
              @select="handleNamespaceSwitch"
            />
            <v-divider />
          </template>

          <v-list-item
            v-if="hasNamespaces && showAdminPanel"
            lines="two"
            @click="navigateToAdminPanel"
          >
            <template #prepend>
              <v-avatar
                size="48"
                color="primary"
                rounded="rounded"
                variant="tonal"
                class="border border-primary border-opacity-100"
              >
                <v-icon>mdi-shield-crown</v-icon>
              </v-avatar>
            </template>
            <v-list-item-title>Admin Console</v-list-item-title>
            <v-list-item-subtitle>
              <div class="d-flex align-center text-caption text-capitalize">
                <div class="d-flex align-center ga-1 flex-1-0">
                  <v-icon size="x-small">
                    mdi-shield-crown
                  </v-icon>
                  <span>super admin</span>
                </div>
                <div class="d-flex align-center ga-1 flex-1-0">
                  <v-icon size="x-small">
                    mdi-server
                  </v-icon>
                  <span>instance</span>
                </div>
              </div>
            </v-list-item-subtitle>
          </v-list-item>
        </template>

        <v-divider />

        <div class="px-4 pt-4 pb-2">
          <v-btn
            variant="flat"
            color="primary"
            prepend-icon="mdi-plus-circle"
            block
            @click="showAddDialog = true"
          >
            Create Namespace
          </v-btn>
        </div>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import NamespaceAdd from "./NamespaceAdd.vue";
import NamespaceChip from "./NamespaceChip.vue";
import NamespaceListItem from "./NamespaceListItem.vue";
import useNamespaceManager from "./composables/useNamespaceManager";
import useAuthStore from "@/store/modules/auth";
import CopyWarning from "@/components/User/CopyWarning.vue";
import { envVariables } from "@/envVariables";

defineOptions({
  inheritAttrs: false,
});

const authStore = useAuthStore();
const {
  currentNamespace,
  namespaceList,
  hasNamespaces,
  switchNamespace,
  loadCurrentNamespace,
} = useNamespaceManager();

const showAddDialog = ref(false);
const userId = computed(() => authStore.id);

const showAdminPanel = computed(() => envVariables.isEnterprise && !envVariables.isCloud);

const otherNamespaces = computed(() => namespaceList.value.filter((ns) => ns.tenant_id !== currentNamespace.value.tenant_id));

const handleNamespaceSwitch = async (tenantId: string) => {
  await switchNamespace(tenantId);
};

const navigateToAdminPanel = () => { window.location.href = "/admin"; };

onMounted(async () => {
  await loadCurrentNamespace();
});
</script>
