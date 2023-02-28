<template>
  <v-app-bar flat floating class="bg-background">
    <v-app-bar-nav-icon
      class="hidden-lg-and-up"
      @click.stop="updateDrawer()"
      aria-label="Toggle Menu"
    />

    <v-spacer />

    <!-- <v-btn
      :size="defaultSize"
      class="ml-1 mr-1 js-gitter-toggle-chat-button"
      color="primary"
      aria-label="gitter-help-icon"
      icon="mdi-help-circle"
    /> -->
    <v-tooltip location="bottom" class="text-center">
      <template v-slot:activator="{ props }">
        <v-btn
          v-bind="props"
          :size="defaultSize"
          class="ml-1 mr-1"
          color="primary"
          aria-label="community-help-icon"
          icon="mdi-help-circle"
          @click="openShellhubHelp()"
        />
      </template>
      <span>Report an issue or make a question for the shellhub team</span>
    </v-tooltip>

    <Notification data-test="notification-component" />

    <v-menu>
      <template v-slot:activator="{ props }">
        <v-btn
          color="primary"
          v-bind="props"
          class="d-flex align-center justify-center"
        >
          <v-icon :size="defaultSize" class="mr-2" left> mdi-account </v-icon>

          <div>{{ currentUser || "USER" }}</div>

          <v-icon :size="defaultSize" class="ml-1 mr-1" right>
            mdi-chevron-down
          </v-icon>
        </v-btn>
      </template>
      <v-list class="bg-v-theme-surface">
        <v-list-item
          v-for="item in menu"
          :key="item.title"
          :value="item"
          :data-test="item.title"
          @click="triggerClick(item)"
        >
          <div class="d-flex align-center">
            <v-icon :icon="item.icon" class="mr-2" />

            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </div>
        </v-list-item>

        <v-divider />

        <v-list-item density="compact">
          <v-switch
            label="Dark Mode"
            :model-value="isDarkMode"
            :onchange="toggleDarkMode"
            data-test="dark-mode-switch"
            density="comfortable"
            color="primary"
            inset
            hide-details
          />
        </v-list-item>
      </v-list>
    </v-menu>
  </v-app-bar>
</template>

<script lang="ts">
import {
  defineComponent,
  computed,
  ref,
  onMounted,
  defineAsyncComponent,
} from "vue";
import { RouteLocationRaw, useRouter } from "vue-router";
import GitterSidecar from "gitter-sidecar";
import { useStore } from "../../store";
import { createNewClient } from "../../api/http";
import handleError from "@/utils/handleError";

type MenuItem = {
  title: string;
  icon: string;
  type: string;
  path: RouteLocationRaw;
  method: () => void;
};

export default defineComponent({
  name: "AppBar",
  inheritAttrs: true,
  setup() {
    const store = useStore();
    const router = useRouter();
    const getStatusDarkMode = computed(
      () => store.getters["layout/getStatusDarkMode"],
    );
    const currentUser = computed(() => store.getters["auth/currentUser"]);
    const defaultSize = ref(24);
    const drawer = ref(true);
    const isDarkMode = ref(getStatusDarkMode.value === "dark");
    const chat = ref(null);

    const triggerClick = (item: MenuItem): void => {
      switch (item.type) {
        case "path":
          router.push(item.path);
          break;
        case "method":
          item.method();
          break;
        default:
          break;
      }
    };

    const logout = async () => {
      try {
        await store.dispatch("auth/logout");
        await store.dispatch("stats/clear");
        await store.dispatch("namespaces/clearNamespaceList");
        await router.push({ name: "login" });
        createNewClient();
        store.dispatch("layout/setLayout", "simpleLayout");
      } catch (error: unknown) {
        handleError(error);
      }
    };

    const toggleDarkMode = () => {
      isDarkMode.value = !isDarkMode.value;
      store.dispatch("layout/setStatusDarkMode", isDarkMode.value);
    };

    const updateDrawer = () => {
      store.dispatch("layout/setStatusNavigationDrawer", true);
    };

    const openShellhubHelp = () => {
      window.open(
        "https://github.com/shellhub-io/shellhub/issues/new/choose",
        "_blank",
      );
    };

    onMounted(async () => {
      chat.value = await new GitterSidecar({
        room: "shellhub-io/community",
        activationElement: false,
        targetElement: chat.value,
      });
    });

    return {
      menu: [
        {
          title: "Settings",
          type: "path",
          path: "/settings",
          icon: "mdi-cog",
          // eslint-disable-next-line no-void
          method: () => void 0,
        },
        {
          title: "Logout",
          type: "method",
          icon: "mdi-logout",
          path: "",
          method: logout,
        },
      ],
      drawer,
      triggerClick,
      currentUser,
      isDarkMode,
      defaultSize,
      toggleDarkMode,
      updateDrawer,
      openShellhubHelp,
    };
  },
  components: {
    Notification: defineAsyncComponent(
      () => import("./Notifications/Notification.vue"),
    ),
  },
});
</script>

<style lang="scss">
.gitter-chat-embed {
  z-index: 9999 !important;
}
</style>
