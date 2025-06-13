<template>
  <v-chip
    class="ma-2"
    :color="statusChipAttributes.color"
    variant="text"
    :prepend-icon="statusChipAttributes.icon"
  >
    {{ statusChipAttributes.label }}
  </v-chip>
</template>

<script setup lang="ts">
import { UserStatus } from "@admin/interfaces/IUser";
import { computed } from "vue";

const { status } = defineProps<{
  status: UserStatus;
}>();

const validStatuses = ["confirmed", "invited", "not-confirmed"];

const safeStatus = computed(() => validStatuses.includes(status) ? status : "not-confirmed");

const statusChipAttributes = computed(() => ({
  confirmed: { color: "success", icon: "mdi-checkbox-marked-circle", label: "Confirmed" },
  invited: { color: "warning", icon: "mdi-email-alert", label: "Invited" },
  "not-confirmed": { color: "error", icon: "mdi-alert-circle", label: "Not Confirmed" },
}[safeStatus.value]));
</script>
