<template>
  <v-list-item
    :active="active"
    :title="namespace.name"
    @click="handleClick"
  >
    <template #prepend>
      <div class="mr-3">
        <NamespaceChip :name="namespace.name" />
      </div>
    </template>

    <v-list-item-subtitle>
      <div class="d-flex align-center text-caption text-capitalize">
        <div
          v-if="userRole"
          class="d-flex align-center ga-1 w-50"
        >
          <v-icon
            size="x-small"
            :icon="roleIcon"
          />
          <span>{{ roleLabel }}</span>
        </div>
        <div class="d-flex align-center ga-1 w-50">
          <v-icon
            size="x-small"
            :icon="namespaceTypeIcon"
          />
          <span>{{ namespaceType }}</span>
        </div>
      </div>
    </v-list-item-subtitle>
  </v-list-item>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { INamespace } from "@/interfaces/INamespace";
import NamespaceChip from "./NamespaceChip.vue";

interface Props {
  namespace: INamespace;
  active?: boolean;
  userId: string;
}

interface Emits {
  (e: "select", tenantId: string): void;
}

const props = withDefaults(defineProps<Props>(), {
  active: false,
});

const emit = defineEmits<Emits>();

const userRole = computed(() => {
  const member = props.namespace.members?.find((m) => m.id === props.userId);
  return member?.role || null;
});

const roleConfig: Record<string, { label: string; icon: string }> = {
  owner: { label: "owner", icon: "mdi-crown" },
  administrator: { label: "admin", icon: "mdi-shield-account" },
  operator: { label: "operator", icon: "mdi-account-cog" },
  observer: { label: "observer", icon: "mdi-eye" },
};

const roleLabel = computed(() => {
  if (!userRole.value) return "";
  return roleConfig[userRole.value]?.label || userRole.value;
});

const roleIcon = computed(() => {
  if (!userRole.value) return "";
  return roleConfig[userRole.value]?.icon || "mdi-account";
});

const namespaceType = computed(() => props.namespace.type || "team");

const namespaceTypeIcon = computed(() => namespaceType.value === "personal" ? "mdi-account" : "mdi-account-group");

const handleClick = () => {
  if (!props.active) {
    emit("select", props.namespace.tenant_id);
  }
};
</script>
