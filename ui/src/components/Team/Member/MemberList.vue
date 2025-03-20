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
    <tbody v-if="namespace" data-test="member-table-rows">
      <slot name="rows">
        <tr v-for="(member, i) in namespace" :key="i" class="text-center">
          <td>
            <v-icon> mdi-account </v-icon>
            {{ member.email }}
          </td>

          <td class="text-center">
            {{ member.role }}
          </td>

          <td class="text-center">
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
                        :notHasAuthorization="!hasAuthorizationEditMember()"
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
import { useStore } from "@/store";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import MemberDelete from "./MemberDelete.vue";
import MemberEdit from "./MemberEdit.vue";
import { INotificationsError } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";

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

const store = useStore();
const tenant = computed(() => store.getters["auth/tenant"]);
const namespace = computed(() => store.getters["namespaces/get"].members);

const hasAuthorizationEditMember = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.editMember,
    );
  }
  return false;
};
const hasAuthorizationRemoveMember = () => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.removeMember,
    );
  }
  return false;
};
const getNamespace = async () => {
  try {
    await store.dispatch("namespaces/get", tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
        handleError(error);
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.namespaceLoad,
      );
      handleError(error);
    }
  }
};
const refresh = () => {
  getNamespace();
};

const isNamespaceOwner = (role: string) => role === "owner";

defineExpose({ namespace });
</script>
