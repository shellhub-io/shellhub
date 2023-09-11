<template>
  <v-list-item
    @click="showDialog = true"
    v-bind="$attrs"
    :disabled="notHasAuthorization"
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
          v-model="tagLocal"
          label="Tag name"
          :error-messages="tagLocalError"
          required
          variant="underlined"
          data-test="name-field"
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

<script lang="ts">
import { useField } from "vee-validate";
import { defineComponent, onMounted, ref, computed } from "vue";
import * as yup from "yup";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

export default defineComponent({
  props: {
    tag: {
      type: String,
      required: true,
    },
    notHasAuthorization: {
      type: Boolean,
      default: false,
    },
  },
  emits: ["update"],
  inheritAttrs: true,
  setup(props, ctx) {
    const store = useStore();
    const showDialog = ref(false);

    const prop = computed(() => props);
    const { value: tagLocal, errorMessage: tagLocalError, resetField: resetTagLocal } = useField<string>(
      "tagLocal",
      yup
        .string()
        .required()
        .min(3)
        .max(255)
        .matches(/^[^/|@|&|:]*$/, "The name must not contain /, @, &, and :"),
      {
        initialValue: prop.value.tag,
      },
    );

    const setLocalTag = () => {
      tagLocal.value = props.tag;
    };

    onMounted(() => {
      setLocalTag();
    });

    const close = () => {
      showDialog.value = false;
      resetTagLocal();
    };

    const update = () => {
      ctx.emit("update");
      close();
    };

    const edit = async () => {
      if (!tagLocalError.value) {
        try {
          await store.dispatch("tags/edit", {
            oldTag: props.tag,
            newTag: tagLocal.value,
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

    return {
      showDialog,
      tagLocal,
      tagLocalError,
      setLocalTag,
      edit,
      close,
      update,
    };
  },
});
</script>
