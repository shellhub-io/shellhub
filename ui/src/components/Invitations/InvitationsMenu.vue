<template>
  <v-badge
    :model-value="pendingInvitationsCount > 0"
    :content="pendingInvitationsCount"
    offset-y="-4"
    offset-x="-4"
    location="top right"
    color="success"
    size="x-small"
    data-test="invitations-menu-badge"
    :class="{ 'mr-1': pendingInvitationsCount > 0 }"
  >
    <v-icon
      color="primary"
      aria-label="Open pending invitations menu"
      icon="mdi-email"
      data-test="invitations-menu-icon"
      @click="toggleDrawer"
    />
  </v-badge>

  <Teleport to="body">
    <v-navigation-drawer
      v-model="isDrawerOpen"
      location="right"
      temporary
      :width="thresholds.sm"
      class="bg-v-theme-surface"
      data-test="invitations-drawer"
    >
      <v-card
        class="bg-v-theme-surface h-100"
        flat
        data-test="invitations-card"
      >
        <v-card-title class="text-h6 py-3">Pending Invitations</v-card-title>

        <v-card-text class="pa-4 pt-0">
          <div
            v-if="isLoading"
            class="d-flex justify-center align-center"
            style="min-height: 200px"
            data-test="loading-state"
          >
            <v-progress-circular indeterminate />
          </div>

          <v-list
            v-else-if="pendingInvitationsList.length > 0"
            density="compact"
            class="bg-v-theme-surface pa-0"
            data-test="invitations-list"
          >
            <InvitationsMenuItem
              v-for="(invitation, index) in pendingInvitationsList"
              :key="index"
              :invitation="invitation"
              @update="fetchInvitations"
            />
          </v-list>

          <div
            v-else
            class="d-flex flex-column justify-center align-center text-center"
            style="min-height: 200px"
            data-test="empty-state"
          >
            <v-icon
              size="64"
              color="medium-emphasis"
              class="mb-4"
              icon="mdi-email-check-outline"
            />
            <div class="text-body-2 text-medium-emphasis">No pending invitations</div>
          </div>
        </v-card-text>
      </v-card>
    </v-navigation-drawer>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeMount, nextTick } from "vue";
import { useDisplay } from "vuetify";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import InvitationsMenuItem from "./InvitationsMenuItem.vue";
import useInvitationsStore from "@/store/modules/invitations";

const { thresholds } = useDisplay();
const invitationsStore = useInvitationsStore();
const snackbar = useSnackbar();

const isDrawerOpen = defineModel<boolean>({ required: true });
const isLoading = ref(false);
const pendingInvitationsList = computed(() => invitationsStore.pendingInvitations);
const pendingInvitationsCount = computed(() => pendingInvitationsList.value.length);

const toggleDrawer = async () => {
  isDrawerOpen.value = !isDrawerOpen.value;
  await nextTick();
  if (isDrawerOpen.value) { await fetchInvitations(); }
};

const fetchInvitations = async () => {
  try {
    isLoading.value = true;
    await invitationsStore.fetchUserPendingInvitationList();
  } catch (error: unknown) {
    snackbar.showError("Failed to fetch pending invitations");
    handleError(error);
  } finally {
    isLoading.value = false;
  }
};

onBeforeMount(async () => {
  await fetchInvitations();
});

defineExpose({
  isDrawerOpen,
  pendingInvitationsList,
  isLoading,
  toggleDrawer,
  fetchInvitations,
});
</script>
