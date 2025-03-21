<template>
  <v-menu :close-on-content-click="false">
    <template v-slot:activator="{ props }">
      <v-badge
        v-bind="$props"
        v-if="showNumberNotifications > 0"
        :content="showNumberNotifications"
        offset-y="-5"
        location="top right"
        color="success"
        size="x-small"
        data-test="notifications-badge"
      >
        <v-icon
          v-bind="props"
          color="primary"
          aria-label="notifications-icon"
          @click="getNotifications()"
        >
          mdi-bell
        </v-icon>
      </v-badge>
      <v-icon
        v-bind="props"
        v-else
        class="ml-2 mr-2"
        color="primary"
        aria-label="notifications-icon"
        @click="getNotifications()"
      >
        mdi-bell
      </v-icon>
    </template>

    <v-card
      v-if="!getStatusNotifications"
      data-test="hasNotifications-subheader"
      offset-x="20"
    >

      <v-list
        class="pa-0"
        density="compact"
      >
        <v-list-subheader>Pending Devices</v-list-subheader>
        <v-divider />

        <v-list-item
          class="d-flex"
          v-for="item in listNotifications"
          :key="item.uid"
        >
          <template v-slot:prepend>
            <v-list-item-title>

              <router-link
                :to="{ name: 'DeviceDetails', params: { id: item.uid } }"
                :data-test="item.uid + '-field'"
              >
                {{ item.name }}
              </router-link>

            </v-list-item-title>
          </template>

          <v-list-item-action>
            <DeviceActionButton
              v-if="hasAuthorization"
              :uid="item.uid"
              :name="item.name"
              variant="device"
              :notification-status="true"
              :show="!getStatusNotifications"
              action="accept"
              :data-test="item.uid + '-btn'"
              @update="refresh"
            />
          </v-list-item-action>
        </v-list-item>
      </v-list>
      <v-btn
        to="/devices/pending"
        variant="tonal"
        link
        block
        size="small"
        data-test="show-btn"
        @click="show = false"
      >
        Show all Pending Devices
      </v-btn>
    </v-card>

    <v-card
      v-else
      data-test="noNotifications-subheader"
      class="pa-2 bg-v-theme-surface"
    >
      <v-card-subtitle> You don't have notifications </v-card-subtitle>
    </v-card>
  </v-menu>
</template>

<script setup lang="ts">
import {
  ref,
  computed,
  watch,
} from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import { authorizer, actions } from "@/authorizer";
import hasPermission from "@/utils/permission";
import { INotificationsError } from "@/interfaces/INotifications";
import DeviceActionButton from "@/components/Devices/DeviceActionButton.vue";
import handleError from "@/utils/handleError";

const store = useStore();
defineProps({
  style: {
    type: [String, Object],
    default: undefined,
  },
});
const show = ref(false);
const inANamespace = ref(false);

const listNotifications = computed(
  () => store.getters["notifications/list"],
);

const getNumberNotifications = computed(
  () => store.getters["notifications/getNumberNotifications"],
);

const showNumberNotifications = computed(() => {
  const numberNotifications = getNumberNotifications.value;
  const pendingDevices = store.getters["stats/stats"].pending_devices;
  if (numberNotifications === 0 && pendingDevices !== undefined) {
    return store.getters["stats/stats"].pending_devices;
  }
  return numberNotifications;
});

const getStatusNotifications = computed(() => {
  if (getNumberNotifications.value === 0) return true;
  return false;
});

const hasNamespace = computed(
  () => store.getters["namespaces/getNumberNamespaces"] !== 0,
);

const hasAuthorization = computed(() => {
  const role = store.getters["auth/role"];
  if (role !== "") {
    return hasPermission(
      authorizer.role[role],
      actions.notification.view,
    );
  }
  return false;
});

watch(hasNamespace, (status) => {
  inANamespace.value = status;
});

const getNotifications = async () => {
  if (hasNamespace.value) {
    try {
      await store.dispatch("notifications/fetch");
      show.value = true;
    } catch (error: unknown) {
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError;
        switch (true) {
          case !inANamespace.value && axiosError.response?.status === 403: {
            // dialog pops
            break;
          }
          case axiosError.response?.status === 403: {
            store.dispatch("snackbar/showSnackbarErrorAssociation");
            handleError(error);
            break;
          }
          default: {
            store.dispatch(
              "snackbar/showSnackbarErrorLoading",
              INotificationsError.notificationList,
            );
            handleError(error);
          }
        }
      } else {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.notificationList,
        );
        handleError(error);
      }
    }
  }
};

const refresh = () => {
  if (hasNamespace.value) {
    getNotifications();
    if (getNumberNotifications.value === 0) {
      store.dispatch("stats/get");
    }
  }
};
</script>
