<template>
  <v-btn
    data-test="users-export-btn"
    class="mr-6"
    v-bind="$attrs"
    color="primary"
    variant="outlined"
    text="Export CSV"
    @click="showDialog = true"
  />

  <FormDialog
    v-model="showDialog"
    title="Export users data"
    icon="mdi-download"
    icon-color="primary"
    confirm-text="Export"
    cancel-text="Cancel"
    :confirm-disabled="!!numberOfNamespacesError || isLoading"
    :confirm-loading="isLoading"
    @confirm="handleSubmit"
    @cancel="closeDialog"
    @close="closeDialog"
  >
    <v-card-text class="pa-6">
      <v-radio-group
        v-model="selectedFilter"
        hide-details
      >
        <v-radio
          data-test="radio-more-than"
          class="mb-1"
          label="Users with more than:"
          :value="FilterOptions.MoreThan"
        />
        <v-radio
          data-test="radio-exactly"
          class="mb-3"
          label="Users with exactly:"
          :value="FilterOptions.Exactly"
        />
      </v-radio-group>
      <v-text-field
        v-model.number="numberOfNamespaces"
        data-test="number-of-namespaces-input"
        suffix="namespaces"
        label="Number of namespaces"
        color="primary"
        density="comfortable"
        variant="outlined"
        :error-messages="numberOfNamespacesError"
      />
    </v-card-text>
  </FormDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { saveAs } from "file-saver";
import * as yup from "yup";
import { useField } from "vee-validate";
import useUsersStore from "@admin/store/modules/users";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";
import FormDialog from "@/components/Dialogs/FormDialog.vue";

enum FilterOptions {
  MoreThan = "moreThan",
  Exactly = "exactly",
}

const isLoading = ref(false);
const showDialog = ref(false);
const selectedFilter = ref<FilterOptions>(FilterOptions.MoreThan);
const snackbar = useSnackbar();
const usersStore = useUsersStore();
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

const closeDialog = () => {
  showDialog.value = false;
  numberOfNamespaces.value = 0;
  selectedFilter.value = FilterOptions.MoreThan;
};

const handleSubmit = async () => {
  isLoading.value = true;
  const encodedFilter = encodeFilter();
  try {
    const response = await usersStore.exportUsersToCsv(encodedFilter);
    const blob = new Blob([response], { type: "text/csv;charset=utf-8" });
    saveAs(blob, getFilename());
    snackbar.showSuccess("Exported users successfully.");
    closeDialog();
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to export users.");
  }

  isLoading.value = false;
};
</script>
