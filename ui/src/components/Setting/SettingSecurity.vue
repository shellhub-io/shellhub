<template>
  <v-form>
    <v-row>
      <v-col class="mb-3">
        <h3 class="mb-3" data-test="security-title">Security</h3>

        <div class="ml-3">
          <v-checkbox
            v-model="sessionRecord"
            :disabled="!hasAuthorization"
            color="primary"
            hide-details
            label="Enable session record"
            data-test="security-checkbox"
          />

          <p>
            Session record is a feature that allows you to check logged activity
            when connecting to a device.
          </p>
        </div>
      </v-col>
    </v-row>
  </v-form>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { useStore } from "../../store";
import { INotificationsSuccess } from "../../interfaces/INotifications";
import handleError from "@/utils/handleError";

const props = defineProps({
  hasTenant: {
    type: Boolean,
    default: false,
  },
});
const store = useStore();

const sessionRecord = ref(store.getters["security/get"]);

watch(sessionRecord, async (value: boolean) => {
  const data = {
    id: localStorage.getItem("tenant"),
    status: value,
  };
  try {
    await store.dispatch("security/set", data);
    store.dispatch(
      "snackbar/showSnackbarSuccessAction",
      INotificationsSuccess.namespaceEdit,
    );
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
});

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.namespace.enableSessionRecord,
    );
  }
  return false;
});

onMounted(async () => {
  try {
    if (props.hasTenant) {
      await store.dispatch("security/get");
    }
  } catch (error: unknown) {
    store.dispatch("snackbar/showSnackbarErrorDefault");
    handleError(error);
  }
});
defineExpose({ sessionRecord });
</script>
