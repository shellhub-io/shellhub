<template>
  <v-table class="bg-v-theme-surface">
    <thead>
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          :class="head.align ? `text-${head.align}` : 'text-center'"
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
    <tbody v-if="members">
      <slot name="rows">
        <tr v-for="(member, i) in members" :key="i">
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
            >This member was added on {{ formatDate(member.added_at) }}</v-tooltip>
          </td>

          <td class="text-end">
            <v-menu
              location="bottom"
              scrim
              eager
              v-if="!isNamespaceOwner(member.role)"
            >
              <template v-slot:activator="{ props }">
                <v-chip v-bind="props" density="comfortable" size="small">
                  <v-icon>mdi-dots-horizontal</v-icon>
                </v-chip>
              </template>
              <v-list class="bg-v-theme-surface" lines="two" density="compact">
                <v-tooltip
                  location="bottom"
                  class="text-center"
                  :disabled="hasAuthorizationEditMember()"
                >
                  <template v-slot:activator="{ props }">
                    <div :v-bind="props">
                      <NamespaceMemberEdit
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
                      <NamespaceMemberDelete
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
import { formatDate } from "../../utils/formateDate";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import NamespaceMemberDelete from "./NamespaceMemberDelete.vue";
import NamespaceMemberEdit from "./NamespaceMemberEdit.vue";
import { INotificationsError } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const props = defineProps({
  namespace: {
    type: Object,
    required: true,
  },
});

const headers = [
  {
    text: "Email",
    value: "email",
    align: "start",
    sortable: false,
  },
  {
    text: "Role",
    value: "role",
    align: "center",
    sortable: false,
  },
  {
    text: "Status",
    value: "status",
    align: "center",
    sortable: false,
  },
  {
    text: "Actions",
    value: "actions",
    align: "end",
    sortable: false,
  },
];
const store = useStore();
const tenant = computed(() => store.getters["auth/tenant"]);
const members = computed(() => props.namespace.members);
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
</script>
