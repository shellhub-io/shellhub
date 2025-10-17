<template>
  <v-card data-test="super-admin-toggle-card">
    <v-card-title>Super Administrator</v-card-title>
    <v-card-text>
      <v-switch
        v-model="isSuperAdmin"
        :disabled="loading || isLastSuperAdmin"
        label="Grant super administrator privileges"
        color="purple"
        data-test="super-admin-switch"
        @update:model-value="toggleSuperAdmin"
      />
      <v-alert
        v-if="isLastSuperAdmin"
        type="warning"
        variant="tonal"
        density="compact"
        class="mt-2"
        data-test="last-super-admin-warning"
      >
        Cannot remove the last super administrator
      </v-alert>
      <v-alert
        v-if="error"
        type="error"
        variant="tonal"
        density="compact"
        class="mt-2"
        data-test="error-alert"
      >
        {{ error }}
      </v-alert>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import * as userApi from "@admin/store/api/users";
import useSnackbar from "@/helpers/snackbar";

interface Props {
  userId: string;
  initialValue: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
  updated: [];
}>();

const snackbar = useSnackbar();
const loading = ref(false);
const isSuperAdmin = ref(props.initialValue);
const isLastSuperAdmin = ref(false);
const error = ref("");

watch(() => props.initialValue, (newVal) => {
  isSuperAdmin.value = newVal;
  isLastSuperAdmin.value = false;
  error.value = "";
});

const toggleSuperAdmin = async () => {
  loading.value = true;
  error.value = "";
  isLastSuperAdmin.value = false;

  try {
    await userApi.toggleSuperAdmin(props.userId, isSuperAdmin.value);
    snackbar.showSuccess(
      isSuperAdmin.value
        ? "User promoted to super administrator"
        : "Super administrator privileges revoked",
    );
    emit("updated");
  } catch (err) {
    const axiosError = err as { response?: { status?: number } };
    if (axiosError.response?.status === 409) {
      isLastSuperAdmin.value = true;
      isSuperAdmin.value = true;
      error.value = "Cannot remove the last super administrator";
      snackbar.showError("Cannot remove the last super administrator");
    } else {
      error.value = "Failed to update super admin status";
      snackbar.showError("Failed to update super admin status");
      isSuperAdmin.value = props.initialValue;
    }
  } finally {
    loading.value = false;
  }
};
</script>
