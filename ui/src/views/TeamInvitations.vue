<template>
  <PageHeader
    icon="mdi-email"
    title="Invitations"
    overline="Team Management"
    description="View and manage pending invitations to this namespace. Track invitation status and send new invitations to collaborators."
    icon-color="primary"
    data-test="title"
  >
    <template #actions>
      <MemberInvite @update="handleNewInvitation" />
    </template>
  </PageHeader>

  <div
    class="mt-2"
    data-test="invitation-list"
  >
    <InvitationList ref="invitationListRef" />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import InvitationList from "@/components/Team/Invitation/InvitationList.vue";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import PageHeader from "@/components/PageHeader.vue";

const invitationListRef = ref<InstanceType<typeof InvitationList> | null>(null);

const handleNewInvitation = () => {
  invitationListRef.value?.setStatusFilterToPending();
  invitationListRef.value?.getInvitations();
};
</script>
