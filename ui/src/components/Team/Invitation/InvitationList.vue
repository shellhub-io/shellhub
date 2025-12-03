<template>
  <div>
    <DataTable
      v-model:page="page"
      v-model:items-per-page="itemsPerPage"
      :headers
      :items="invitations"
      :total-count="invitationCount"
      :loading
      :items-per-page-options="[10, 20, 50, 100]"
      data-test="invitations-list"
    >
      <template #rows>
        <tr
          v-for="invitation in invitations"
          :key="invitation.user.id"
        >
          <td class="text-center">
            <v-icon icon="mdi-email" />
            {{ invitation.user.email }}
          </td>

          <td class="text-center text-capitalize">{{ invitation.role }}</td>

          <td class="text-center">
            <v-chip
              :color="getStatusColor(invitation.status)"
              variant="tonal"
              size="small"
              class="text-capitalize font-weight-medium"
              :text="invitation.status"
            />
          </td>

          <td class="text-center">{{ formatShortDateTime(invitation.created_at) }}</td>

          <td class="text-center">
            <span
              v-if="isInvitationPending(invitation) && isInvitationExpired(invitation.expires_at)"
              class="text-error font-weight-medium"
            >
              <v-icon
                icon="mdi-alert-circle"
                size="small"
                class="mr-1"
              />
              Expired at {{ formatShortDateTime(invitation.expires_at) }}
            </span>
            <template v-else-if="isInvitationPending(invitation)">
              {{ formatShortDateTime(invitation.expires_at) }}
            </template>
            <template v-else>
              {{ formatShortDateTime(invitation.status_updated_at) }}
            </template>
          </td>

          <td class="text-center">
            <v-menu
              v-if="isInvitationPending(invitation) || invitation.status === 'cancelled'"
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
                  data-test="invitation-actions"
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
                  :disabled="canSendInvitation"
                >
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <InvitationResend
                        :invitation="invitation"
                        :has-authorization="canSendInvitation"
                        @update="getInvitations"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
                <v-tooltip
                  v-if="isInvitationPending(invitation)"
                  location="bottom"
                  class="text-center"
                  :disabled="canEditInvitation"
                >
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <InvitationEdit
                        :invitation="invitation"
                        :has-authorization="canEditInvitation"
                        @update="getInvitations"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>

                <v-tooltip
                  v-if="isInvitationPending(invitation)"
                  location="bottom"
                  class="text-center"
                  :disabled="canCancelInvitation"
                >
                  <template #activator="{ props }">
                    <div v-bind="props">
                      <InvitationCancel
                        :invitation="invitation"
                        :has-authorization="canCancelInvitation"
                        :is-remove="false"
                        @update="getInvitations"
                      />
                    </div>
                  </template>
                  <span> You don't have this kind of authorization. </span>
                </v-tooltip>
              </v-list>
            </v-menu>
            <v-tooltip
              v-else
              location="bottom"
              activator="parent"
              :text="`${invitation.status === 'accepted' ? 'Accepted' : 'Rejected'} invitations cannot be resent, edited or cancelled`"
            />
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import axios from "axios";
import { formatShortDateTime } from "@/utils/date";
import hasPermission from "@/utils/permission";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import InvitationEdit from "./InvitationEdit.vue";
import InvitationCancel from "./InvitationCancel.vue";
import InvitationResend from "./InvitationResend.vue";
import useAuthStore from "@/store/modules/auth";
import useInvitationsStore from "@/store/modules/invitations";
import { IInvitation } from "@/interfaces/IInvitation";
import { getInvitationStatusFilter, orderInvitationsByCreatedAt, isInvitationExpired } from "@/utils/invitations";

const props = defineProps<{ statusFilter: IInvitation["status"] }>();

const authStore = useAuthStore();
const invitationsStore = useInvitationsStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const tenant = computed(() => authStore.tenantId);
const invitations = computed(() => orderInvitationsByCreatedAt(invitationsStore.namespaceInvitations));
const invitationCount = computed(() => invitationsStore.invitationCount);
const filter = computed(() => getInvitationStatusFilter(props.statusFilter));
const canEditInvitation = hasPermission("namespace:editInvitation");
const canCancelInvitation = hasPermission("namespace:cancelInvitation");
const canSendInvitation = hasPermission("namespace:addMember");

const dateColumnHeader = computed(() => {
  if (props.statusFilter === "pending") return "Expires At";
  const capitalizedStatus = props.statusFilter.charAt(0).toUpperCase() + props.statusFilter.slice(1);
  return `${capitalizedStatus} At`;
});

const headers = computed(() => [
  {
    text: "Email",
    value: "email",
  },
  {
    text: "Role",
    value: "role",
  },
  {
    text: "Status",
    value: "status",
  },
  {
    text: "Created At",
    value: "created_at",
  },
  {
    text: dateColumnHeader.value,
    value: "date_column",
  },
  {
    text: "Actions",
    value: "actions",
  },
]);

const statusColorMap: Record<IInvitation["status"], string> = {
  accepted: "success",
  rejected: "error",
  pending: "warning",
  cancelled: "grey",
};

const getStatusColor = (status: IInvitation["status"]) => statusColorMap[status] || "grey";

const isInvitationPending = (invitation: IInvitation) => invitation.status === "pending";

const getInvitations = async () => {
  try {
    loading.value = true;

    await invitationsStore.fetchNamespaceInvitationList(
      tenant.value,
      page.value,
      itemsPerPage.value,
      filter.value,
    );
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      if (error.response?.status === 403) {
        snackbar.showError("You don't have permission to access this resource.");
      }
    } else {
      snackbar.showError("Failed to load the invitation list.");
      handleError(error);
    }
  } finally {
    loading.value = false;
  }
};

watch(filter, async () => { page.value = 1; await getInvitations(); });

watch([page, itemsPerPage], async () => { await getInvitations(); });

onMounted(async () => { await getInvitations(); });

defineExpose({ getInvitations });
</script>
