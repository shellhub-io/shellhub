<template>
  <v-list-item
    @click="open"
    v-bind="$attrs"
    :disabled="notHasAuthorization"
    data-test="open-tag-edit"
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

  <v-dialog v-model="showDialog" min-width="300" max-width="600">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary"> Update Tag </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1">
        <v-text-field
          v-model="inputTags"
          label="Tag name"
          :error-messages="tagsError"
          required
          variant="underlined"
          data-test="tag-field"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" data-test="close-btn" @click="close()">
          Close
        </v-btn>

        <v-btn
          color="primary"
          variant="text"
          data-test="edit-btn"
          @click="edit()"
        >
          Edit
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const props = defineProps({
  tag: {
    type: String,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update"]);
const store = useStore();
const showDialog = ref(false);

const inputTags = ref<string>("");
const tagsError = ref("");
const tagsHasLessThan3Characters = computed(() => inputTags.value.length < 3);

watch(inputTags, () => {
  if (inputTags.value.length > 255) {
    tagsError.value = "Maximum of 3 tags";
  } else if (tagsHasLessThan3Characters.value) {
    tagsError.value = "The minimum length is 3 characters";
  } else {
    tagsError.value = "";
  }
});

const open = () => {
  inputTags.value = props.tag;
  showDialog.value = true;
};

const close = () => {
  showDialog.value = false;
};

const update = () => {
  emit("update");
  close();
};

const edit = async () => {
  if (!tagsError.value) {
    try {
      await store.dispatch("tags/edit", {
        oldTag: props.tag,
        newTag: inputTags.value,
      });

      update();
      store.dispatch(
        "snackbar/showSnackbarSuccessAction",
        INotificationsSuccess.deviceTagEdit,
      );
    } catch (error: unknown) {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.deviceTagEdit,
      );
      handleError(error);
    }
  }
};

defineExpose({ inputTags });
</script>
