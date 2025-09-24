<template>
  <v-list-item
    @click="showDialog = true"
    v-bind="$attrs"
    :disabled="!hasAuthorization"
    data-test="firewall-delete-dialog-btn"
  >
    <div class="d-flex align-center">
      <div class="mr-2">
        <v-icon data-test="remove-icon"> mdi-delete </v-icon>
      </div>

      <v-list-item-title data-test="remove-title"> Remove </v-list-item-title>
    </div>
  </v-list-item>

  <MessageDialog
    v-model="showDialog"
    @close="showDialog = false"
    @confirm="remove"
    @cancel="showDialog = false"
    title="Are you sure?"
    description="You are about to delete this firewall rule"
    icon="mdi-alert"
    icon-color="error"
    confirm-text="Delete"
    confirm-color="error"
    cancel-text="Close"
    confirm-data-test="confirm-btn"
    cancel-data-test="close-btn"
    data-test="delete-firewall-rule-dialog"
  />
</template>

<script setup lang="ts">
import { ref } from "vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "../MessageDialog.vue";
import useFirewallRulesStore from "@/store/modules/firewall_rules";

const props = defineProps<{
  id: string;
  hasAuthorization: boolean;
}>();

const emit = defineEmits(["update"]);
const showDialog = ref(false);
const firewallRulesStore = useFirewallRulesStore();
const snackbar = useSnackbar();

const remove = async () => {
  try {
    await firewallRulesStore.removeFirewallRule(props.id);
    snackbar.showSuccess("Firewall rule deleted successfully.");
    emit("update");
  } catch (error: unknown) {
    snackbar.showError("Failed to delete firewall rule.");
    handleError(error);
  } finally {
    showDialog.value = false;
  }
};
</script>
