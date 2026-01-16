<template>
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
        control-variant="hidden"
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
import { computed } from "vue";
import * as yup from "yup";
import useAdminNamespacesStore from "@admin/store/modules/namespaces";
import useNamespacesStore from "@/store/modules/namespaces";
import useSnackbar from "@/helpers/snackbar";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

const props = defineProps<{ namespace: IAdminNamespace }>();
const emit = defineEmits(["update"]);

const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const adminNamespacesStore = useAdminNamespacesStore();
const showDialog = defineModel<boolean>({ required: true });

const {
  value: name,
  errorMessage: nameError,
  resetField: resetName,
} = useField<string>("name", yup.string().required(), {
  initialValue: props.namespace.name || "",
});

const {
  value: maxDevices,
  errorMessage: maxDevicesError,
  resetField: resetMaxDevices,
} = useField<number>(
  "maxDevices",
  yup
    .number()
    .integer()
    .required()
    .min(-1, "Maximum devices must be -1 (unlimited) or greater"),
  {
    initialValue: props.namespace.max_devices ?? -1,
  },
);

const {
  value: sessionRecord,
  errorMessage: sessionRecordError,
  resetField: resetSessionRecord,
} = useField<boolean>("sessionRecord", yup.boolean(), {
  initialValue: props.namespace.settings?.session_record ?? true,
});

const hasErrors = computed(
  () => nameError.value || maxDevicesError.value || sessionRecordError.value,
);

const closeDialog = () => {
  showDialog.value = false;
  resetName();
  resetMaxDevices();
  resetSessionRecord();
};

const submitForm = async () => {
  if (hasErrors.value) return;

  try {
    await adminNamespacesStore.updateNamespace({
      ...props.namespace,
      name: name.value,
      max_devices: Number(maxDevices.value),
      settings: {
        ...props.namespace.settings,
        session_record: sessionRecord.value,
      },
    });
    await namespacesStore.fetchNamespaceList({ perPage: 30 });
    snackbar.showSuccess("Namespace updated successfully.");
    showDialog.value = false;
    emit("update");
  } catch {
    snackbar.showError("Failed to update namespace.");
  }
};
</script>
