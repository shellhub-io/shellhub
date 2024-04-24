<template>
  <v-list-item
    @click="showDialog = true"
    v-bind="$attrs"
    :disabled="notHasAuthorization"
    data-test="firewall-delete-dialog-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="remove-icon"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <v-dialog max-width="450" v-model="showDialog">
    <v-card class="bg-v-theme-surface" data-test="firewallRuleDelete-card">
      <v-card-title class="text-h5 pa-5 bg-primary" data-test="text-title">
        Are you sure?
      </v-card-title>
      <v-divider />

      <v-card-text class="mt-4 mb-0 pb-1" data-test="text-text">
        <p class="text-body-2 mb-2">
          You are about to remove this firewall rule.
        </p>

        <p class="text-body-2 mb-2">
          After confirming this action cannot be redone.
        </p>
      </v-card-text>

      <v-card-actions>
        <v-spacer />

        <v-btn variant="text" data-test="close-btn" @click="showDialog = false">
          Close
        </v-btn>

        <v-btn
          color="red darken-1"
          data-test="remove-btn"
          variant="text"
          @click="remove()"
        >
          Remove
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  INotificationsError,
  INotificationsSuccess,
} from "../../interfaces/INotifications";
import { useStore } from "../../store";
import handleError from "@/utils/handleError";

const props = defineProps({
  id: {
    type: String,
    required: true,
  },
  notHasAuthorization: {
    type: Boolean,
    default: false,
  },
});
const emit = defineEmits(["update"]);
const showDialog = ref(false);
const store = useStore();

const remove = async () => {
  try {
    await store.dispatch("firewallRules/remove", props.id);

    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.firewallRuleDeleting,
    );
    emit("update");
  } catch (error: unknown) {
    store.dispatch(
      "snackbar/showSnackbarErrorAction",
      INotificationsError.firewallRuleDeleting,
    );
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
