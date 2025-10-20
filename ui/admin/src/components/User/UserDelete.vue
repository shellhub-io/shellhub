<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon tag="a" dark v-bind="props" @click="showDialog = true">mdi-delete </v-icon>
    </template>
    <span>Remove</span>
  </v-tooltip>

  <BaseDialog v-model="showDialog">
    <v-card>
      <v-card-title class="lighten-2 text-center mt-2"> Are you sure? </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-3 pb-1">
        You are about to remove this user and all its namespace associated data. This action will
        also bill usage for namespace subscriptions if any and might take some time to finish.
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text @click="showDialog = false"> Close </v-btn>

        <v-btn color="red darken-1" text @click="remove()"> Remove </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "@/components/Dialogs/BaseDialog.vue";
import handleError from "@/utils/handleError";

const props = defineProps<{
  id: string;
  redirect?: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const router = useRouter();
const snackbar = useSnackbar();
const usersStore = useUsersStore();

const remove = async () => {
  try {
    await usersStore.deleteUser(props.id);
    snackbar.showSuccess("User removed successfully.");
    if (props.redirect) router.push("/users");
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
