<template>
  <FormDialog
    v-model="showDialog"
    title="Create Tag"
    icon="mdi-tag"
    confirm-text="Create"
    cancel-text="Close"
    :confirm-disabled="confirmDisabled"
    confirm-data-test="create-btn"
    cancel-data-test="close-btn"
    data-test="tag-create-dialog"
    @close="close"
    @cancel="close"
    @confirm="create"
  >
    <div class="pa-6">
      <v-text-field
        v-model="tagInput"
        label="Tag name"
        :error-messages="tagError"
        required
        hide-details="auto"
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

const emit = defineEmits(["update"]);
const tagsStore = useTagsStore();
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });

const tagInput = ref<string>("");
const tagError = ref("");
const confirmDisabled = computed(() => !tagInput.value || !!tagError.value);

const validateTagInput = () => {
  const inputLength = tagInput.value.length;
  if (inputLength > 255) tagError.value = "The maximum length is 255 characters";
  else if (inputLength < 3) tagError.value = "The minimum length is 3 characters";
  else tagError.value = "";
};

const close = () => {
  tagInput.value = "";
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const create = async () => {
  if (tagError.value) return;

  try {
    await tagsStore.createTag(tagInput.value);

    snackbar.showSuccess("Successfully created tag");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to create tag.");
    handleError(error);
  }
};

defineExpose({ tagInput, showDialog });
</script>
