<template>
  <v-form>
    <v-row>
      <v-col class="mb-3">
        <h3 class="mb-3">Security</h3>

        <div class="ml-3">
          <v-checkbox
            v-model="sessionRecord"
            :disabled="!hasAuthorization"
            color="primary"
            hide-details
            label="Enable session record"
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

<script lang="ts">
import { defineComponent, computed, onMounted, ref, watch } from "vue";
import hasPermission from "../../utils/permission";
import { actions, authorizer } from "../../authorizer";
import { useStore } from "../../store";
import { INotificationsSuccess } from "../../interfaces/INotifications";

export default defineComponent({
  props: {
    hasTenant: {
      type: Boolean,
      default: false,
    },
  },
  setup(props: { hasTenant: any }) {
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
          INotificationsSuccess.namespaceEdit
        );
      } catch {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    });

    const hasAuthorization = computed(() => {
      const role = store.getters["auth/role"];
      if (role !== "") {
        return hasPermission(
          authorizer.role[role],
          actions.namespace["enableSessionRecord"]
        );
      }
      return false;
    });

    onMounted(async () => {
      try {
        if (props.hasTenant) {
          await store.dispatch("security/get");
        }
      } catch {
        store.dispatch("snackbar/showSnackbarErrorDefault");
      }
    });

    return {
      sessionRecord,
      hasAuthorization,
    };
  },
});
</script>
