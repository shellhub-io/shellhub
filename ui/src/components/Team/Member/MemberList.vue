<template>
  <v-table
    class="bg-background border rounded"
    data-test="member-table"
  >
    <thead class="bg-v-theme-background">
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          class="text-center"
          data-test="member-table-headers"
        >
          <span
            v-if="head.sortable"
            tabindex="0"
            class="hover"
            @click="emit('clickSortableIcon', head.value)"
            @keypress.enter="emit('clickSortableIcon', head.value)"
          >
            {{ head.text }}
            <v-tooltip
              activator="parent"
              anchor="top"
            >Sort by {{ head.text }}</v-tooltip>
          </span>
          <span v-else> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody
      v-if="members"
      data-test="member-table-rows"
    >
      <slot name="rows">
        <tr
          v-for="member in members"
          :key="member.id"
          class="text-center"
        >
          <td>
            <v-icon> mdi-account </v-icon>
            {{ member.email }}
          </td>

          <td class="text-center text-capitalize">
            {{ member.role }}
          </td>

          <td class="text-center text-capitalize">
            {{ member.status }}
            <v-tooltip
              v-if="member.added_at !== '0001-01-01T00:00:00Z'"
              activator="parent"
              location="bottom"
            >
              This member was added on {{ formatFullDateTime(member.added_at) }}
            </v-tooltip>
          </td>

          <td class="text-center">
            <v-menu
              v-if="!isNamespaceOwner(member.role)"
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
                  data-test="namespace-member-actions"
                />
              </template>
              <v-list
                class="bg-v-theme-surface"
                lines="two"
                density="compact"
              >
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="canEditMember"
                >
                  <template #activator="{ props }">
                    <div :v-bind="props">
                      <MemberEdit
                        :member="member"
                        :has-authorization="canEditMember"
                        @update="refresh"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="canRemoveMember"
                >
                  <template #activator="{ props }">
                    <div :v-bind="props">
                      <MemberDelete
                        :member="member"
                        :has-authorization="canRemoveMember"
                        @update="refresh"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
              </v-list>
            </v-menu>
            <v-tooltip
              v-else
              activator="parent"
              location="top"
            >
              No one can modify the owner of this namespace.
            </v-tooltip>
          </td>
        </tr>
      </slot>
    </tbody>
    <div
      v-else
      class="pa-4 text-subtitle-2"
    >
      <p>No data available</p>
    </div>
  </v-table>
</template>

<script setup lang="ts">
import { computed } from "vue";
import axios, { AxiosError } from "axios";
import { formatFullDateTime } from "@/utils/date";
import hasPermission from "@/utils/permission";
import MemberDelete from "./MemberDelete.vue";
import MemberEdit from "./MemberEdit.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useAuthStore from "@/store/modules/auth";
import useNamespacesStore from "@/store/modules/namespaces";

const headers = [
  {
    text: "Email",
    value: "email",
    sortable: false,
  },
  {
    text: "Role",
    value: "role",
    sortable: false,
  },
  {
    text: "Status",
    value: "status",
    sortable: false,
  },
  {
    text: "Actions",
    value: "actions",
    sortable: false,
  },
];
const emit = defineEmits<{
  (e: "clickSortableIcon", value: string): void;
}>();
const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const tenant = authStore.tenantId;
const members = computed(() => namespacesStore.currentNamespace.members);

const canEditMember = hasPermission("namespace:editMember");

const canRemoveMember = hasPermission("namespace:removeMember");

const getNamespace = async () => {
  try {
    await namespacesStore.fetchNamespace(tenant);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You don't have permission to view this namespace.");
        handleError(error);
      }
    } else {
      snackbar.showError("Failed to fetch namespace members.");
      handleError(error);
    }
  }
};
const refresh = async () => { await getNamespace(); };

const isNamespaceOwner = (role: string) => role === "owner";
</script>
