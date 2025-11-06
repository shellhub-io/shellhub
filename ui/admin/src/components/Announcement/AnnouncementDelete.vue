<template>
  <v-tooltip
    bottom
    anchor="bottom"
  >
    <template #activator="{ props }">
      <v-icon
        tag="button"
        v-bind="props"
        data-test="delete-button"
        icon="mdi-delete"
        @click="showDialog = true"
      />
    </template>
    <span>Remove</span>
  </v-tooltip>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this announcement. This action cannot be undone."
    icon="mdi-alert-circle"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Cancel"
    @confirm="remove"
    @cancel="showDialog = false"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import useAnnouncementStore from "@admin/store/modules/announcement";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";

const props = defineProps<{ uuid: string }>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const announcementStore = useAnnouncementStore();
const snackbar = useSnackbar();

const remove = async () => {
  try {
    await announcementStore.deleteAnnouncement(props.uuid);
    emit("update");
    snackbar.showSuccess("Announcement deleted successfully.");
    showDialog.value = false;
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to delete announcement.");
  }
};

defineExpose({ showDialog });
</script>
