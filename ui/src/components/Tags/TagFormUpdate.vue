<template>
  <v-list-item
    v-bind="$attrs"
    @click="showDialog = true"
    :disabled="notHasAuthorization"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon> mdi-tag </v-icon>
      </div>

      <v-list-item-title data-test="mdi-information-list-item">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog v-model="showDialog" min-width="280" max-width="450">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-4 bg-primary">
        {{ hasTags ? "Edit tags" : "Add Tags" }}
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-5 w-100">
        <v-combobox
          id="targetInput"
          full-width
          ref="tags"
          v-model="inputTags"
          :error-messages="tagsError"
          label="Tag"
          hint="Maximum of 3 tags"
          multiple
          chips
          variant="outlined"
          data-test="deviceTag-combobox"
          :deletable-chips="true"
          :delimiters="[',', ' ']"
        />
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn
          variant="text"
          data-test="close-btn"
          @click="close()"
          class="mr-2"
        >
          Close
        </v-btn>

        <v-btn variant="text" data-test="save-btn" @click="save()">
          Save
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
import { computed, defineComponent, ref, watch } from "vue";
import { useStore } from "../../store";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";

export default defineComponent({
  props: {
    deviceUid: {
      type: String,
      required: true,
    },

    tagsList: {
      type: Array<string>,
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

    const hasTags = computed(() => props.tagsList.length > 0);
    const inputTags = ref(props.tagsList);
    const tagsError = ref("");

    const tagsHasLessThan3Characters = computed(() => inputTags.value.some((tag) => tag.length < 3));

    watch(inputTags, () => {
      if (inputTags.value.length > 3) {
        tagsError.value = "Maximum of 3 tags";
      } else if (tagsHasLessThan3Characters.value) {
        tagsError.value = "The minimum length is 3 characters";
      } else {
        tagsError.value = "";
      }
    });

    const save = async () => {
      if (tagsError.value) return;
      try {
        tagsError.value = "";
        await store.dispatch("devices/updateDeviceTag", {
          uid: props.deviceUid,
          tags: { tags: inputTags.value },
        });

        await store.dispatch("tags/setTags", {
          data: inputTags.value,
          headers: {
            "x-total-count": inputTags.value.length,
          },
        });
        showDialog.value = false;
        store.dispatch(
          "snackbar/showSnackbarSuccessAction",
          INotificationsSuccess.deviceTagUpdate,
        );

        ctx.emit("update");
      } catch (error: any) {
        switch (error.response.status) {
          // when the name the format is invalid.
          case 400: {
            tagsError.value = "The format is invalid. Min 3, Max 255 characters!";
            break;
          }
          // when the user is not authorized.
          case 403: {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.deviceTagUpdate,
            );
            break;
          }
          // When the array tag size reached the max capacity.
          case 406: {
            tagsError.value = "The maximum capacity has reached.";
            break;
          }
          default: {
            store.dispatch(
              "snackbar/showSnackbarErrorAction",
              INotificationsError.deviceTagUpdate,
            );
            throw new Error(error);
          }
        }
      }
    };

    const close = () => {
      showDialog.value = false;
      inputTags.value = props.tagsList;
    };

    return {
      inputTags,
      tagsError,
      showDialog,
      hasTags,
      tagsHasLessThan3Characters,
      save,
      close,
    };
  },
});
</script>
