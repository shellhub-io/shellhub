<template>
  <v-select
    v-model="selectedRole"
    :items="roles"
    label="Role"
    required
    hide-details
    data-test="role-select"
  >
    <template #item="{ props, item }">
      <v-list-item v-bind="props">
        <v-list-item-subtitle class="description-text">
          {{ item.raw.description }}
        </v-list-item-subtitle>
      </v-list-item>
    </template>
  </v-select>
</template>

<script setup lang="ts">
import { BasicRole } from "@/interfaces/INamespace";

const roles = [
  {
    title: "Administrator",
    value: "administrator",
    // eslint-disable-next-line vue/max-len
    description: "Full access to the namespace, can perform all actions except managing billing.\nThis includes user and device management, but excludes billing-related operations.",
  },
  {
    title: "Operator",
    value: "operator",
    // eslint-disable-next-line vue/max-len
    description: "Can manage and operate devices, but has limited administrative privileges.\nOperators cannot change billing or ownership settings.",
  },
  {
    title: "Observer",
    value: "observer",
    description: "Can view device details and sessions but cannot make any changes.\nObservers have read-only access to monitor activity.",
  },
];

const selectedRole = defineModel<BasicRole>({ required: true });
</script>

<style scoped lang="scss">
.description-text {
  white-space: normal;
  word-break: break-word;
  max-width: 600px;
  display: block;
}
</style>
