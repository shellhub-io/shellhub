<template>
  <v-btn
    @click="showDialog = true"
    variant="plain"
    class="border rounded bg-v-theme-background"
    density="comfortable"
    size="default"
    icon="mdi-delete"
    :disabled="!hasAuthorizationDeleteWebEndpoint"
    data-test="web-endpoint-delete-dialog-btn"
  />
  <BaseDialog v-model="showDialog">
    <v-card class="bg-v-theme-surface">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="text">
        <p class="text-body-2 mb-2">
          You are about to remove this Web Endpoint.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" @click="showDialog = false" data-test="close-btn"> Close </v-btn>

        <v-btn color="red darken-1" variant="text" @click="remove()" data-test="delete-btn">
          Delete Web Endpoint
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useStore } from "@/store";
import hasPermission from "@/utils/permission";
import { actions, authorizer } from "@/authorizer";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import BaseDialog from "../BaseDialog.vue";
import useAuthStore from "@/store/modules/auth";

defineOptions({
  inheritAttrs: false,
});

const props = defineProps({
  address: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(["update"]);
const showDialog = defineModel({ default: false });
const store = useStore();
const authStore = useAuthStore();
const snackbar = useSnackbar();

const update = () => {
  emit("update");
  showDialog.value = false;
};

const hasAuthorizationDeleteWebEndpoint = () => {
  const { role } = authStore;
  return !!role && hasPermission(authorizer.role[role], actions.webendpoint.delete);
};

const remove = async () => {
  try {
    await store.dispatch("webEndpoints/delete", {
      address: props.address,
    });
    update();
    snackbar.showSuccess("Web Endpoint deleted successfully.");
  } catch (error: unknown) {
    snackbar.showError("Failed to delete Web Endpoint.");
    handleError(error);
  }
};
</script>
