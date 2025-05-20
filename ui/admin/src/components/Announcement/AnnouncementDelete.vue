<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        tag="a"
        dark
        v-bind="props"
        @click="dialog = !dialog"
      >
        mdi-delete
      </v-icon>
    </template>
    <span>Remove</span>
  </v-tooltip>

  <v-dialog max-width="450" v-model="dialog">
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

        <v-btn text @click="dialog = !dialog"> Cancel </v-btn>

        <v-btn color="red darken-1" text @click="remove()"> Remove </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import useAnnouncementStore from "@admin/store/modules/announcement";
import useSnackbar from "@/helpers/snackbar";

const props = defineProps({
  uuid: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update"]);
const dialog = ref(false);
const announcement = useAnnouncementStore();
const snackbar = useSnackbar();
const remove = async () => {
  dialog.value = !dialog.value;

  try {
    await announcement.deleteAnnouncement(props.uuid);
    emit("update");
    snackbar.showSuccess("Announcement deleted successfully.");
  } catch {
    snackbar.showError("Failed to delete announcement.");
  }
};

defineExpose({ dialog });
</script>
