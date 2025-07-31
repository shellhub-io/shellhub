<template>
  <div>
    <v-list-item
      @click="showDialog = true"
      :disabled="!hasAuthorization"
      data-test="member-edit-btn"
    >
      <div class="d-flex align-center">
        <div class="mr-2">
          <v-icon> mdi-pencil </v-icon>
        </div>

        <v-list-item-title data-test="member-edit-title">
          Edit
        </v-list-item-title>
      </div>
    </v-list-item>

    <BaseDialog v-model="showDialog">
      <v-card class="bg-v-theme-surface" data-test="member-edit-dialog">
        <v-card-title class="text-h5 pa-4 bg-primary" data-test="member-edit-dialog-title">
          Update member role
        </v-card-title>
        <v-divider />

        <v-card-text class="mt-4 mb-0 pb-1">
          <RoleSelect
            v-model="newRole"
            data-test="role-select"
          />
        </v-card-text>

        <v-card-actions>
          <v-spacer />
          <v-btn variant="text" data-test="close-btn" @click="close()">
            Close
          </v-btn>

          <v-btn
            color="primary"
            variant="text"
            data-test="edit-btn"
            @click="editMember()"
          >
            Edit
          </v-btn>
        </v-card-actions>
      </v-card>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import axios from "axios";
import { BasicRole, INamespaceMember } from "@/interfaces/INamespace";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/BaseDialog.vue";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import useAuthStore from "@/store/modules/auth";

const { member, hasAuthorization } = defineProps<{
  member: INamespaceMember;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const store = useStore();
const authStore = useAuthStore();
const snackbar = useSnackbar();
const showDialog = ref(false);
const newRole = ref(member.role as BasicRole);

const close = () => {
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const handleEditMemberError = (error: unknown) => {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status;
    switch (status) {
      case 400:
        snackbar.showError("The user isn't linked to the namespace.");
        break;
      case 403:
        snackbar.showError("You don't have permission to assign a role to the user.");
        break;
      case 404:
        snackbar.showError("The username doesn't exist.");
        break;
      default:
        snackbar.showError("Failed to update user role.");
    }
  } else snackbar.showError("Failed to update user role.");

  handleError(error);
};

const editMember = async () => {
  try {
    await store.dispatch("namespaces/editUser", {
      user_id: member.id,
      tenant_id: authStore.tenantId,
      role: newRole.value,
    });

    snackbar.showSuccess("Successfully updated user role.");
    update();
  } catch (error: unknown) {
    handleEditMemberError(error);
  }
};

</script>
