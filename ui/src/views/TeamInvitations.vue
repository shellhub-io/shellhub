<template>
  <div
    class="d-flex pa-0 align-center"
    data-test="title"
  >
    <h1>Invitations</h1>

    <v-select
      v-model="statusFilter"
      :items="statusOptions"
      label="Filter by status"
      variant="outlined"
      density="comfortable"
      hide-details
      class="mx-4"
      style="max-width: 200px;"
      data-test="invitation-status-select"
    />
    <v-spacer />

    <MemberInvite @update="handleNewInvitation" />
  </div>

  <div
    class="mt-2"
    data-test="invitation-list"
  >
    <InvitationList
      ref="invitationListRef"
      :status-filter="statusFilter"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { IInvitation } from "@/interfaces/IInvitation";
import InvitationList from "@/components/Team/Invitation/InvitationList.vue";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";

const invitationListRef = ref<InstanceType<typeof InvitationList> | null>(null);
const statusFilter = ref<IInvitation["status"]>("pending");

const statusOptions: { title: string; value: IInvitation["status"] }[] = [
  { title: "Pending", value: "pending" },
  { title: "Cancelled", value: "cancelled" },
  { title: "Accepted", value: "accepted" },
  { title: "Rejected", value: "rejected" },
];

const handleNewInvitation = () => {
  statusFilter.value = "pending";
  invitationListRef.value?.getInvitations();
};
</script>
