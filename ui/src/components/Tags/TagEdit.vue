<template>
  <v-list-item
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="open-tag-edit"
    @click="open"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-pencil </v-icon>
      </div>

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
    <div class="px-6 pt-4">
      <v-text-field
        v-model="inputTags"
        label="Tag name"
        :error-messages="tagsError"
        required
        data-test="tag-field"
      />
    </div>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import FormDialog from "@/components/Dialogs/FormDialog.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import useTagsStore from "@/store/modules/tags";

const props = defineProps({
  tagName: { type: String, required: true },
  hasAuthorization: { type: Boolean, default: false },
});

const emit = defineEmits(["update"]);
const tagsStore = useTagsStore();
const snackbar = useSnackbar();

const showDialog = ref(false);
const inputTags = ref<string>("");
const tagsError = ref("");

const tenant = computed(() => localStorage.getItem("tenant"));

const tagsHasLessThan3Characters = computed(() => inputTags.value.length < 3);

watch(inputTags, () => {
  if (inputTags.value.length > 255) {
    tagsError.value = "The maximum length is 255 characters";
  } else if (tagsHasLessThan3Characters.value) {
    tagsError.value = "The minimum length is 3 characters";
  } else {
    tagsError.value = "";
  }
});

const confirmDisabled = computed(() => !inputTags.value || !!tagsError.value);

const open = () => {
  showDialog.value = true;
  inputTags.value = props.tagName;
};

const close = () => {
  showDialog.value = false;
  inputTags.value = "";
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (tagsError.value) return;

  try {
    await tagsStore.editTag({
      tenant: tenant.value || "",
      currentName: props.tagName,
      newName: { name: inputTags.value },
    });

    snackbar.showSuccess("Tag updated successfully.");
    update();
  } catch (error: unknown) {
    snackbar.showError("Failed to update tag.");
    handleError(error);
  }
};

defineExpose({ inputTags });
</script>
