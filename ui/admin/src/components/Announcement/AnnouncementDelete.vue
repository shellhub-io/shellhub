<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        tag="button"
        dark
        v-bind="props"
        data-test="delete-button"
        @click="showDialog = true"
      >
        mdi-delete
      </v-icon>
    </template>
    <span>Removex</span>
  </v-tooltip>

  <BaseDialog v-model="showDialog">
    <v-card>
      <v-card-title class="lighten-2 text-center mt-2">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-3 pb-1">
        You are about to remove this announcement.
        <p>This action cannot be undone.</p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn text @click="showDialog = false"> Cancel </v-btn>

        <v-btn color="red darken-1" text @click="remove()"> Remove </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useAnnouncementStore from "@admin/store/modules/announcement";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import BaseDialog from "@/components/BaseDialog.vue";

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
