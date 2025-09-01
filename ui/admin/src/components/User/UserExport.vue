<template>
  <v-btn class="mr-6" @click="showDialog = true" v-bind="$attrs">Export CSV</v-btn>

  <BaseDialog v-model="showDialog" @close="closeDialog" transition="dialog-bottom-transition">
    <v-card>
      <v-card-title class="text-h5 pb-2">Export users data</v-card-title>
      <v-divider />
      <v-form @submit.prevent="handleSubmit">
        <v-card-text>
          <v-radio-group v-model="selectedFilter">
            <v-radio class="mb-1" label="Users with more than:" :value="FilterOptions.MoreThan" />
            <v-radio class="mb-1" label="Users with exactly:" :value="FilterOptions.Exactly" />
          </v-radio-group>
          <v-row no-gutters class="d-flex justify-center align-center ml-3">
            <v-text-field
              v-model="numberOfNamespaces"
              suffix="namespaces"
              label="Number of namespaces"
              color="primary"
              density="comfortable"
              variant="outlined"
              :error-messages="numberOfNamespacesError"
            />
          </v-row>
        </v-card-text>

        <v-card-actions class="pa-4 d-flex justify-end ga-2">
          <v-btn @click="closeDialog">Cancel</v-btn>
          <v-btn color="primary" type="submit" :loading="isLoading" :disabled="!!numberOfNamespacesError || isLoading">Export</v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { saveAs } from "file-saver";
import * as yup from "yup";
import { useField } from "vee-validate";
import useUsersStore from "@admin/store/modules/users";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import BaseDialog from "@/components/BaseDialog.vue";

enum FilterOptions {
  MoreThan = "moreThan",
  Exactly = "exactly",
}

const isLoading = ref(false);
const showDialog = ref(false);
const selectedFilter = ref<FilterOptions>(FilterOptions.MoreThan);
const snackbar = useSnackbar();
const userStore = useUsersStore();
const { value: numberOfNamespaces,
  errorMessage: numberOfNamespacesError,
} = useField<number>("numberOfNamespaces", yup.number().integer().required().min(0), { initialValue: 0 });

const encodeFilter = () => {
  const filter = [
    {
      type: "property",
      params: {
        name: "namespaces",
        operator: selectedFilter.value === FilterOptions.MoreThan ? "gt" : "eq",
        value: numberOfNamespaces.value,
      },
    },
  ];

  return btoa(JSON.stringify(filter));
};

const getFilename = () => {
  const filterType = selectedFilter.value === FilterOptions.MoreThan ? "more_than" : "exactly";
  return `users_with_${filterType}_${numberOfNamespaces.value}_namespaces.csv`;
};

const handleSubmit = async () => {
  isLoading.value = true;
  const encodedFilter = encodeFilter();
  try {
    await userStore.setFilterUsers(encodedFilter);
    const response = await userStore.exportUsersToCsv();
    const blob = new Blob([response], { type: "text/csv;charset=utf-8" });
    saveAs(blob, getFilename());
    snackbar.showSuccess("Exported users successfully.");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to export users.");
  }

  isLoading.value = false;
};

const resetForm = () => {
  numberOfNamespaces.value = 0;
  selectedFilter.value = FilterOptions.MoreThan;
};

const closeDialog = () => {
  showDialog.value = false;
  resetForm();
};

defineExpose({ numberOfNamespaces, showDialog, selectedFilter });
</script>
