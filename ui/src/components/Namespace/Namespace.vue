<template>
  <NamespaceAdd v-model="showAddDialog" />

  <v-menu
    :close-on-content-click="false"
    scrim
    location="bottom"
    :offset="4"
  >
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        variant="outlined"
        class="text-none px-2 border-thin"
        height="auto"
      >
        <div class="d-flex align-center ga-2">
          <NamespaceChip :name="currentNamespace.name" />
          <span class="text-body-1">{{ currentNamespace.name || 'No Namespace' }}</span>
          <v-icon size="small">mdi-chevron-down</v-icon>
        </div>
      </v-btn>
    </template>

    <v-card :width="$vuetify.display.thresholds.sm / 2" border>
      <v-list class="bg-v-theme-surface">
        <div class="d-flex align-center justify-space-between pr-4">
          <v-list-subheader>Active Namespace</v-list-subheader>
          <v-btn
            @click="showAddDialog = true"
            variant="flat"
            color="primary"
            prepend-icon="mdi-plus-circle"
            size="small"
          >
            Create
          </v-btn>
        </div>

        <NamespaceListItem
          v-if="currentNamespace.tenant_id"
          :namespace="currentNamespace"
          :active="true"
          :user-id="userId"
          @select="handleNamespaceSwitch"
        />

        <div v-if="currentNamespace.tenant_id" class="px-4 pb-2 pt-3">
          <div class="text-caption text-grey mb-1">Tenant ID</div>
          <div class="d-flex align-center ga-2 pa-2 border-thin rounded text-caption">
            <span class="flex-1-1 text-truncate">{{ currentNamespace.tenant_id }}</span>
            <CopyWarning :copied-item="'Tenant ID'">
              <template #default="{ copyText }">
                <v-icon
                  @click="copyText(currentNamespace.tenant_id)"
                  size="small"
                  class="cursor-pointer"
                >
                  mdi-content-copy
                </v-icon>
              </template>
            </CopyWarning>
          </div>
        </div>

        <template v-if="otherNamespaces.length > 0">
          <v-divider class="my-2" />
          <v-list-subheader>Switch Namespace</v-list-subheader>

          <template v-for="(ns, index) in otherNamespaces" :key="ns.tenant_id">
            <NamespaceListItem
              :namespace="ns"
              :active="false"
              :user-id="userId"
              @select="handleNamespaceSwitch"
            />
            <v-divider v-if="index < otherNamespaces.length - 1" />
          </template>
        </template>

        <template v-if="hasNamespaces && showAdminPanel">
          <v-divider class="my-2" />

          <div class="px-4 py-2">
            <v-btn
              @click="navigateToAdminPanel"
              variant="tonal"
              color="primary"
              prepend-icon="mdi-shield-crown"
              block
            >
              ShellHub Admin
            </v-btn>
          </div>
        </template>
      </v-list>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useRouter } from "vue-router";
import NamespaceAdd from "./NamespaceAdd.vue";
import NamespaceChip from "./NamespaceChip.vue";
import NamespaceListItem from "./NamespaceListItem.vue";
import useNamespaceManager from "./composables/useNamespaceManager";
import useAuthStore from "@/store/modules/auth";
import CopyWarning from "@/components/User/CopyWarning.vue";

defineOptions({
  inheritAttrs: false,
});

const router = useRouter();
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

// TODO: Implement super admin detection
const showAdminPanel = computed(() => true);

const otherNamespaces = computed(() => namespaceList.value.filter((ns) => ns.tenant_id !== currentNamespace.value.tenant_id));

const handleNamespaceSwitch = async (tenantId: string) => {
  await switchNamespace(tenantId);
};

const navigateToAdminPanel = () => {
  router.push("/admin");
};

onMounted(async () => {
  await loadCurrentNamespace();
});
</script>
