<template>
  <v-tooltip bottom anchor="bottom">
    <template v-slot:activator="{ props }">
      <v-icon
        @click="showDialog = true"
        tag="button"
        dark
        v-bind="props"
        tabindex="0"
        aria-label="Edit Namespace"
        data-test="dialog-btn"
        icon="mdi-pencil"
      />
    </template>
    <span>Edit</span>
  </v-tooltip>

  <FormDialog
    v-model="showDialog"
    title="Edit Namespace"
    icon="mdi-folder-edit"
    icon-color="primary"
    confirm-text="Save"
    cancel-text="Cancel"
    :confirm-disabled="!!hasErrors"
    @close="closeDialog"
    @confirm="submitForm"
    @cancel="closeDialog"
  >
    <v-card-text class="pa-6">
      <v-text-field
        v-model="name"
        label="Name"
        required
        :error-messages="nameError"
        data-test="name-text"
      />
      <v-number-input
        v-model="maxDevices"
        label="Maximum Devices"
        required
        variant="outlined"
        density="comfortable"
        inset
        controlVariant="hidden"
        :error-messages="maxDevicesError"
        data-test="maxDevices-text"
      />
      <v-switch
        v-model="sessionRecord"
        :error-messages="sessionRecordError"
        color="primary"
        hide-details
        label="Session Record"
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { useField } from "vee-validate";
import { computed, ref } from "vue";
import * as yup from "yup";
import useNamespacesStore from "@admin/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";
import { IAdminNamespace } from "../../interfaces/INamespace";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

const props = defineProps<{ namespace: IAdminNamespace }>();

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const showDialog = ref(false);

const { value: name, errorMessage: nameError, resetField: resetName } = useField<string | undefined>(
  "name",
  yup.string().required(),
  { initialValue: props.namespace.name },
);

const { value: maxDevices, errorMessage: maxDevicesError, resetField: resetMaxDevices } = useField<number | undefined>(
  "maxDevices",
  yup.number().integer().required().min(-1, "Maximum devices must be -1 (unlimited) or greater"),
  { initialValue: props.namespace.max_devices },
);

const { value: sessionRecord, errorMessage: sessionRecordError, resetField: resetSessionRecord } = useField<boolean>(
  "sessionRecord",
  yup.boolean(),
  { initialValue: props.namespace.settings.session_record || false },
);

const hasErrors = computed(() => nameError.value || maxDevicesError.value || sessionRecordError.value);

const closeDialog = () => {
  showDialog.value = false;
  resetName();
  resetMaxDevices();
  resetSessionRecord();
};

const submitForm = async () => {
  if (hasErrors.value) return;
  try {
    await namespacesStore.updateNamespace({
      ...props.namespace as IAdminNamespace,
      name: name.value as string,
      max_devices: Number(maxDevices.value),
      settings: { session_record: sessionRecord.value },
    });
    await namespacesStore.fetchNamespaceList();
    snackbar.showSuccess("Namespace updated successfully.");
    showDialog.value = false;
  } catch (error) {
    snackbar.showError("Failed to update namespace.");
  }
};

defineExpose({ name, maxDevices, sessionRecord, submitForm });
</script>
