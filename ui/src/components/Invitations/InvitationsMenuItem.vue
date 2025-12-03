<template>
  <v-list-item class="mb-2">
    <v-card variant="tonal">
      <v-card-title class="d-flex align-center pa-3">
        <v-btn
          variant="text"
          size="small"
          :icon="isExpanded ? 'mdi-chevron-up' : 'mdi-chevron-down'"
          class="mr-2"
          @click="isExpanded = !isExpanded"
        />
        <div class="flex-grow-1">
          <div class="text-body-1 font-weight-medium">{{ invitation.namespace.name }}</div>
          <div class="d-flex align-center ga-1 text-body-2 text-capitalize text-medium-emphasis">
            <v-icon
              size="x-small"
              :icon="getRoleIcon(invitation.role)"
            />
            <span>{{ getRoleLabel(invitation.role) }}</span>
          </div>
        </div>
        <div class="d-flex ga-1">
          <InvitationDecline
            v-slot="{ openDialog }"
            :tenant="invitation.namespace.tenant_id"
            :namespace-name="invitation.namespace.name"
            :on-success="handleSuccess"
          >
            <v-btn
              variant="text"
              color="error"
              icon="mdi-close"
              @click="openDialog"
            />
          </InvitationDecline>
          <InvitationAccept
            v-slot="{ openDialog }"
            :tenant="invitation.namespace.tenant_id"
            :namespace-name="invitation.namespace.name"
            :role="invitation.role"
            :on-success="handleSuccess"
          >
            <v-btn
              variant="text"
              color="success"
              icon="mdi-check"
              @click.stop="openDialog"
            />
          </InvitationAccept>
        </div>
      </v-card-title>
      <v-expand-transition>
        <div v-show="isExpanded">
          <v-divider />
          <v-card-text class="text-caption text-medium-emphasis">
            Invited by {{ invitation.invited_by }} at {{ formatFullDateTime(invitation.created_at) }}
          </v-card-text>
        </div>
      </v-expand-transition>
    </v-card>
  </v-list-item>
</template>

<script setup lang="ts">
import { ref } from "vue";
import type { IInvitation } from "@/interfaces/IInvitation";
import type { Role } from "@/interfaces/INamespace";
import { formatFullDateTime } from "@/utils/date";
import InvitationAccept from "./InvitationAccept.vue";
import InvitationDecline from "./InvitationDecline.vue";

interface Emits {
  (e: "update"): void;
}

const emit = defineEmits<Emits>();
defineProps<{ invitation: IInvitation }>();

const isExpanded = ref(false);

const roleConfig: Record<string, { label: string; icon: string }> = {
  owner: { label: "owner", icon: "mdi-crown" },
  administrator: { label: "admin", icon: "mdi-shield-account" },
  operator: { label: "operator", icon: "mdi-account-cog" },
  observer: { label: "observer", icon: "mdi-eye" },
};

const getRoleLabel = (role: Role): string => roleConfig[role]?.label || role;

const getRoleIcon = (role: Role): string => roleConfig[role]?.icon || "mdi-account";

const handleSuccess = () => { emit("update"); };

defineExpose({ handleSuccess });
</script>
