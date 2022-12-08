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
            <v-tooltip activator="parent" anchor="top"
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
            {{ member.username }}
          </td>

          <td class="text-center">
            {{ member.role }}
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
                  :disabled="hasAuthorizationRemoveMember()"
                >
                  <template v-slot:activator="{ props }">
                    <NamespaceMemberEdit
                      :member="member"
                      :v-bind="props"
                      @update="refresh"
                      :notHasAuthorization="!hasAuthorizationEditMember()"
                    />
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  location="bottom"
                  :disabled="hasAuthorizationRemoveMember()"
                >
                  <template v-slot:activator="{ props }">
                    <NamespaceMemberDelete
                      :v-bind="props"
                      :member="member"
                      @update="refresh"
                      :hasAuthorization="hasAuthorizationRemoveMember()"
                    />
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

<script lang="ts">
import { defineComponent, ref, onMounted, computed } from "vue";
import { useStore } from "../../store";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import NamespaceMemberDelete from "./NamespaceMemberDelete.vue";
import NamespaceMemberEdit from "./NamespaceMemberEdit.vue";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  props: {
    namespace: {
      type: Object,
      required: true,
    },
  },
  setup(props) {
    const store = useStore();
    const tenant = computed(() => store.getters["auth/tenant"]);
    const members = computed(() => props.namespace.members);
    const hasAuthorizationEditMember = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.namespace["editMember"]
        );
      }
      return false;
    };
    const hasAuthorizationRemoveMember = () => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.namespace["removeMember"]
        );
      }
      return false;
    };
    const getNamespace = async () => {
      try {
        await store.dispatch("namespaces/get", tenant.value);
      } catch (error: any) {
        if (error.response.status === 403) {
          store.dispatch("snackbar/showSnackbarErrorAssociation");
        } else {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceLoad
          );
        }
      }
    };
    const refresh = () => {
      getNamespace();
    };

    const isNamespaceOwner = (role: string) => {
      return role === "owner";
    };

    return {
      headers: [
        {
          text: "Username",
          value: "username",
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
          text: "Actions",
          value: "actions",
          align: "end",
          sortable: false,
        },
      ],
      tenant,
      members,
      hasAuthorizationEditMember,
      hasAuthorizationRemoveMember,
      isNamespaceOwner,
      refresh,
    };
  },
  components: { NamespaceMemberDelete, NamespaceMemberEdit },
});
</script>
