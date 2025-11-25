<template>
  <v-tooltip
    location="bottom"
    :disabled="!showTooltip"
    text="Remove"
  >
    <template #activator="{ props: tooltipProps }">
      <span
        v-bind="tooltipProps"
        role="button"
      >
        <slot :open-dialog="openDialog" />
      </span>
    </template>
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
import { useRouter } from "vue-router";
import useAnnouncementStore from "@admin/store/modules/announcement";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";

const props = defineProps<{
  uuid: string;
  showTooltip?: boolean;
  redirect?: boolean;
}>();

const emit = defineEmits(["update"]);
const router = useRouter();
const showDialog = ref(false);
const announcementStore = useAnnouncementStore();
const snackbar = useSnackbar();

const openDialog = () => { showDialog.value = true; };

const remove = async () => {
  try {
    await announcementStore.deleteAnnouncement(props.uuid);
    snackbar.showSuccess("Announcement deleted successfully.");
    showDialog.value = false;
    if (props.redirect) void router.push({ name: "announcements" });
    else emit("update");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to delete announcement.");
  }
};

defineExpose({ showDialog, openDialog });
</script>
