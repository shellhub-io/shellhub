<template>
  <v-btn
    @click="showDialog = true"
    class="mr-2 mt-2 mt-md-0"
    color="primary"
    variant="outlined"
    data-test="namespaces-export-btn"
    text="Export CSV"
  />

  <FormDialog
    v-model="showDialog"
    title="Export namespaces data"
    icon="mdi-download"
    icon-color="primary"
    confirm-text="Export"
    cancel-text="Cancel"
    :confirm-disabled="!!numberOfDevicesError || isLoading"
    :confirm-loading="isLoading"
    @confirm="handleSubmit"
    @cancel="closeDialog"
    @close="closeDialog"
  >
    <v-card-text class="pa-6">
      <v-radio-group v-model="selectedFilter" @update:model-value="handleSelectedFilterUpdate">
        <v-radio label="Namespaces with more than:" :value="AdminNamespaceFilterOptions.MoreThan" />
        <v-text-field
          class="mt-2 mx-2"
          v-model="numberOfDevices"
          suffix="devices"
          :disabled="selectedFilter !== AdminNamespaceFilterOptions.MoreThan"
          label="Number of devices"
          color="primary"
          density="comfortable"
          :hide-details="hideNumberOfDevicesInputDetails"
          :error-messages="numberOfDevicesError"
        />
        <v-radio label="Namespaces with no devices" :value="AdminNamespaceFilterOptions.NoDevices" />
        <v-radio label="Namespace with devices, but no sessions" :value="AdminNamespaceFilterOptions.NoSessions" />
      </v-radio-group>
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import { saveAs } from "file-saver";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { AdminNamespaceFilterOptions } from "@admin/interfaces/IFilter";
import useSnackbar from "@/helpers/snackbar";
import getFilter from "@/utils/namespaceExport";
import handleError from "@/utils/handleError";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

const showDialog = ref(false);
const isLoading = ref(false);
const selectedFilter = ref(AdminNamespaceFilterOptions.MoreThan);
const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const { value: numberOfDevices,
  errorMessage: numberOfDevicesError,
} = useField<number>("numberOfDevices", yup.number().integer().required().min(0), { initialValue: 0 });
const hideNumberOfDevicesInputDetails = ref(false);

const handleSelectedFilterUpdate = () => {
  hideNumberOfDevicesInputDetails.value = selectedFilter.value !== AdminNamespaceFilterOptions.MoreThan;
};

const encodeFilter = () => btoa(JSON.stringify(getFilter(selectedFilter.value, numberOfDevices.value)));

const getFilename = () => {
  const filterSuffixes = {
    [AdminNamespaceFilterOptions.MoreThan]: `more_than_${numberOfDevices.value}_devices`,
    [AdminNamespaceFilterOptions.NoDevices]: "no_devices",
    [AdminNamespaceFilterOptions.NoSessions]: "with_devices_but_no_sessions",
  };

  const suffix = filterSuffixes[selectedFilter.value] ?? "export";
  return `namespaces_${suffix}.csv`;
};

const resetForm = () => {
  numberOfDevices.value = 0;
  selectedFilter.value = AdminNamespaceFilterOptions.MoreThan;
};

const closeDialog = () => {
  showDialog.value = false;
  resetForm();
};

const exportCsv = async () => {
  const encodedFilter = encodeFilter();
  const response = await namespacesStore.exportNamespacesToCsv(encodedFilter);
  const blob = new Blob([response], { type: "text/csv;charset=utf-8" });
  saveAs(blob, getFilename());
};

const handleSubmit = async () => {
  isLoading.value = true;
  try {
    await exportCsv();
    snackbar.showSuccess("Namespaces exported successfully.");
    closeDialog();
  } catch (error) {
    handleError(error);
    snackbar.showError("Error exporting namespaces.");
  } finally {
    isLoading.value = false;
  }
};
</script>
