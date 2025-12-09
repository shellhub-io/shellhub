<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="open-tag-remove"
    @click="showDialog = true"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        Remove
      </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    title="Are you sure?"
    description="You are about to remove this tag. After confirming this action cannot be redone."
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Remove"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="confirm-btn"
    cancel-data-test="close-btn"
    data-test="delete-tag-dialog"
    @close="showDialog = false"
    @cancel="showDialog = false"
    @confirm="remove"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "@/components/Dialogs/MessageDialog.vue";
import useTagsStore from "@/store/modules/tags";

defineOptions({ inheritAttrs: false });

const props = defineProps<{
  tagName: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);

const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const showDialog = ref(false);

const update = () => {
  emit("update");
  showDialog.value = false;
};

const remove = async () => {
  try {
    await tagsStore.removeTag({
      currentName: props.tagName,
    });
    snackbar.showSuccess(`${props.tagName} was removed successfully.`);
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to remove tag.");
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
