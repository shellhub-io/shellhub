<template>
  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this user and all its namespace associated data.
    This action will also bill usage for namespace subscriptions if any and might take some time to finish."
    icon="mdi-alert-circle"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Close"
    @confirm="remove"
    @cancel="showDialog = false"
  />
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import handleError from "@/utils/handleError";

const props = defineProps<{
  id: string;
  redirect?: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = defineModel<boolean>({ required: true });
const router = useRouter();
const snackbar = useSnackbar();
const usersStore = useUsersStore();

const remove = async () => {
  try {
    await usersStore.deleteUser(props.id);
    snackbar.showSuccess("User removed successfully.");
    if (props.redirect) await router.push("/users");
    await usersStore.fetchUsersList();
    showDialog.value = false;
    emit("update");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to remove the user.");
  }
};

defineExpose({ showDialog });
</script>
