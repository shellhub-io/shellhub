<template>
  <v-btn @click="showDialog = true" class="mr-2" data-test="namespaces-export-btn">Export CSV</v-btn>

  <BaseDialog v-model="showDialog" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2">Export namespaces data</v-card-title>
      <v-divider />
      <v-form @submit.prevent="handleSubmit" data-test="form">
        <v-card-text>
          <v-radio-group v-model="selectedFilter">
            <v-radio label="Namespaces with more than:" :value="AdminNamespaceFilterOptions.MoreThan" />
            <v-text-field
              class="mt-2 mx-2"
              v-model="numberOfDevices"
              suffix="devices"
              :disabled="selectedFilter !== AdminNamespaceFilterOptions.MoreThan"
              label="Number of devices"
              color="primary"
              density="comfortable"
              variant="outlined"
              :error-messages="numberOfDevicesError"
            />
            <v-radio label="Namespaces with no devices" :value="AdminNamespaceFilterOptions.NoDevices" />
            <v-radio label="Namespace with devices, but no sessions" :value="AdminNamespaceFilterOptions.NoSessions" />
          </v-radio-group>
        </v-card-text>

        <v-card-actions class="pa-4 d-flex justify-end ga-2">
          <v-btn @click="closeDialog">Cancel</v-btn>
          <v-btn color="primary" type="submit" :loading="isLoading" :disabled="!!numberOfDevicesError || isLoading">Export</v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import * as yup from "yup";
import { useField } from "vee-validate";
import { saveAs } from "file-saver";
import useNamespacesStore from "@admin/store/modules/namespaces";
import getFilter from "@admin/hooks/namespaceExport";
import { AdminNamespaceFilterOptions } from "@admin/interfaces/IFilter";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import BaseDialog from "@/components/BaseDialog.vue";

const showDialog = ref(false);
const isLoading = ref(false);
const selectedFilter = ref(AdminNamespaceFilterOptions.MoreThan);
const snackbar = useSnackbar();
const namespacesStore = useNamespacesStore();
const { value: numberOfDevices,
  errorMessage: numberOfDevicesError,
  setErrors: setNumberOfDevicesErrors,
} = useField<number>("numberOfDevices", yup.number().integer().required().min(0), { initialValue: 0 });

watch(selectedFilter, (newValue) => {
  if (newValue !== AdminNamespaceFilterOptions.MoreThan) {
    setNumberOfDevicesErrors("");
  }
});

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

const exportCsv = async () => {
  const encodedFilter = encodeFilter();
  await namespacesStore.setFilterNamespaces(encodedFilter);
  const response = await namespacesStore.exportNamespacesToCsv();
  const blob = new Blob([response], { type: "text/csv;charset=utf-8" });
  saveAs(blob, getFilename());
};

const handleSubmit = async () => {
  isLoading.value = true;
  try {
    await exportCsv();
    snackbar.showSuccess("Namespaces exported successfully.");
  } catch (error) {
    handleError(error);
    snackbar.showError("Error exporting namespaces.");
  }
  isLoading.value = false;
};

const resetForm = () => {
  numberOfDevices.value = 0;
  selectedFilter.value = AdminNamespaceFilterOptions.MoreThan;
};

const closeDialog = () => {
  showDialog.value = false;
  resetForm();
};
</script>
