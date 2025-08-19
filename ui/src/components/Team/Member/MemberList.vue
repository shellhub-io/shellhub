<template>
  <v-table class="bg-background border rounded" data-test="member-table">
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
            @click="$emit('clickSortableIcon', head.value)"
            @keypress.enter="$emit('clickSortableIcon', head.value)"
            tabindex="0"
            class="hover"
          >
            {{ head.text }}
            <v-tooltip
              activator="parent"
              anchor="top"
            >Sort by {{ head.text }}</v-tooltip
            >
          </span>
          <span v-else> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="members" data-test="member-table-rows">
      <slot name="rows">
        <tr v-for="member in members" :key="member.id" class="text-center">
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
            >This member was added on {{ formatFullDateTime(member.added_at) }}</v-tooltip>
          </td>

          <td class="text-center">
            <v-menu
              location="bottom"
              scrim
              eager
              v-if="!isNamespaceOwner(member.role)"
            >
              <template v-slot:activator="{ props }">
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
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationEditMember()"
                >
                  <template v-slot:activator="{ props }">
                    <div :v-bind="props">
                      <MemberEdit
                        :member="member"
                        @update="refresh"
                        :hasAuthorization="hasAuthorizationEditMember()"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationRemoveMember()"
                >
                  <template v-slot:activator="{ props }">
                    <div :v-bind="props">
                      <MemberDelete
                        :member="member"
                        @update="refresh"
                        :hasAuthorization="hasAuthorizationRemoveMember()"
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
            >No one can modify the owner of this namespace.</v-tooltip>
          </td>
        </tr>
      </slot>
    </tbody>
    <div v-else class="pa-4 text-subtitle-2">
      <p>No data available</p>
    </div>
  </v-table>
</template>

<script setup lang="ts">
import { computed } from "vue";
import axios, { AxiosError } from "axios";
import { formatFullDateTime } from "@/utils/date";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
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

const authStore = useAuthStore();
const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const tenant = authStore.tenantId;
const members = computed(() => namespacesStore.currentNamespace.members);

const hasAuthorizationEditMember = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.editMember);
};

const hasAuthorizationRemoveMember = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.namespace.removeMember);
};

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
