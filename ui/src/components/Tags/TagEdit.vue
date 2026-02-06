<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="open-tag-edit"
    @click="open"
  >
    <div class="d-flex align-center ga-2">
      <v-icon icon="mdi-pencil" />
      <v-list-item-title data-test="mdi-information-list-item">
        Edit
      </v-list-item-title>
    </div>
  </v-list-item>

  <FormDialog
    v-model="showDialog"
    title="Update Tag"
    icon="mdi-tag"
    confirm-text="Edit"
    cancel-text="Close"
    :confirm-disabled="confirmDisabled"
    confirm-data-test="edit-btn"
    cancel-data-test="close-btn"
    data-test="tag-edit-dialog"
    @close="close"
    @cancel="close"
    @confirm="edit"
  >
    <div class="pa-6">
      <v-text-field
        v-model="tagInput"
        label="Tag name"
        hide-details="auto"
        :error-messages="tagError"
        required
        data-test="tag-field"
        @update:model-value="validateTagInput"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";

const props = defineProps<{
  tagName: string;
  hasAuthorization?: boolean;
}>();

const emit = defineEmits(["update"]);
const tagsStore = useTagsStore();
const snackbar = useSnackbar();

const showDialog = ref(false);
const tagInput = ref<string>("");
const tagError = ref("");
const confirmDisabled = computed(() => !tagInput.value || !!tagError.value);

const validateTagInput = () => {
  const inputLength = tagInput.value.length;
  if (inputLength > 255) tagError.value = "The maximum length is 255 characters";
  else if (inputLength < 3) tagError.value = "The minimum length is 3 characters";
  else tagError.value = "";
};

const open = () => {
  showDialog.value = true;
  tagInput.value = props.tagName;
};

const close = () => {
  showDialog.value = false;
  tagInput.value = "";
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (tagError.value) return;

  try {
    await tagsStore.updateTag(
      props.tagName,
      { name: tagInput.value },
    );

    snackbar.showSuccess("Tag updated successfully.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to update tag.");
    handleError(error);
  }
};
</script>
