<template>
  <v-app>
    <component :is="layout" :data-test="layout + '-component'" />
  </v-app>
</template>

<script lang="ts">
import { defineComponent, onMounted, computed } from "vue";
import { useRouter } from "vue-router";
import SimpleLayout from "./layouts/SimpleLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import { useStore } from "./store";
import { INotificationsSuccess, INotificationsError } from "@/interfaces/INotifications";

export default defineComponent({
  name: "App",
  components: {
    appLayout: AppLayout,
    simpleLayout: SimpleLayout,
  },
  setup() {
    const store = useStore();
    const router = useRouter();

    const layout = computed(() => store.getters["layout/getLayout"]);
    // const token = computed(() => window.location.search.replace("?token=", ""));

    const isLoggedIn = computed(() => store.getters["auth/isLoggedIn"]);
    const hasLoggedID = computed(() => store.getters["auth/id"] !== "");

    onMounted(async () => {
      if (!isLoggedIn.value) {
        store.dispatch("layout/setLayout", "simpleLayout");
      }

      if (!hasLoggedID.value && isLoggedIn.value) {
        try {
          await store.dispatch("auth/logout");

          store.dispatch("layout/setLayout", "simpleLayout");
          router.push("/login");

          store.dispatch(
            "snackbar/showSnackbarSuccessAction",
            INotificationsSuccess.namespaceReload,
          );
        } catch {
          store.dispatch(
            "snackbar/showSnackbarErrorAction",
            INotificationsError.namespaceLoad,
          );
        }
      }
    });

    return {
      layout,
    };
  },
});
</script>
