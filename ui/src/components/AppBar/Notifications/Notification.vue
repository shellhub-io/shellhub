<template>
  <v-menu>
    <template v-slot:activator="{ props }">
      <v-badge
        v-bind="$props"
        v-if="showNumberNotifications > 0"
        :content="showNumberNotifications"
        :value="showNumberNotifications"
        offset-x="10"
        location="top left"
        color="success"
        data-test="notifications-badge"
      >
        <v-icon
          v-bind="props"
          class="ml-2 mr-2"
          color="primary"
          :size="defaultSize"
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
        :size="defaultSize"
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
      <v-card-subtitle>Pending Devices</v-card-subtitle>
      <v-divider />

      <v-list class="pa-0" density="compact">
        <v-list-item-group :v-model="1">
          <v-list-item class="d-flex" v-for="item in listNotifications" :key="item.uid">
            <template v-slot:prepend>
              <v-list-item-title>
                <router-link
                  :to="{ name: 'detailsDevice', params: { id: item.uid } }"
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
                :notification-status="true"
                :show="!getStatusNotifications"
                action="accept"
                :data-test="item.uid + '-btn'"
                @update="refresh"
              />
            </v-list-item-action>
          </v-list-item>
        </v-list-item-group>
      </v-list>

      <v-divider />

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

<script lang="ts">
import {
  defineComponent,
  ref,
  computed,
  watch,
  defineAsyncComponent,
} from "vue";
import { useStore } from "../../../store";
import { authorizer, actions } from "../../../authorizer";
import hasPermission from "../../../utils/permission";
import { INotificationsError } from "../../../interfaces/INotifications";
import DeviceActionButton from "../../../components/Devices/DeviceActionButton.vue";

export default defineComponent({
  name: "Notification",
  inheritAttrs: false,
  components: {
    DeviceActionButton,
  },
  setup() {
    const store = useStore();
    const numberNotifications = ref(0);
    const show = ref(false);
    const inANamespace = ref(false);
    const defaultSize = ref(24);

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
        } catch (error: any) {
          switch (true) {
            case !inANamespace.value && error.response.status === 403: {
              // dialog pops
              break;
            }
            case error.response.status === 403: {
              store.dispatch("snackbar/showSnackbarErrorAssociation");
              throw new Error(error);
            }
            default: {
              store.dispatch(
                "snackbar/showSnackbarErrorLoading",
                INotificationsError.notificationList,
              );
              throw new Error(error);
            }
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
    return {
      showNumberNotifications,
      getNotifications,
      defaultSize,
      getStatusNotifications,
      listNotifications,
      hasAuthorization,
      refresh,
      show,
    };
  },
});
</script>
