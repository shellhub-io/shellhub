<template>
  <v-app>
    <component :is="layout" :data-test="`${layout}-component`" />
  </v-app>
</template>

<script lang="ts">
import { defineComponent, onMounted, computed } from "vue";
import { useRouter } from "vue-router";
import SimpleLayout from "./layouts/SimpleLayout.vue";
import AppLayout from "./layouts/AppLayout.vue";
import { useStore } from "./store";
import { INotificationsError } from "./interfaces/INotifications";

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

    const isLoggedIn = computed(() => store.getters["auth/isLoggedIn"]);

    const currentRoute = computed(() => router.currentRoute.value.path);

    onMounted(async () => {
      if (!isLoggedIn.value) {
        try {
          await store.dispatch("auth/logout");
          store.dispatch("layout/setLayout", "simpleLayout");
          router.push("/login");
        } catch {
          store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.namespaceLoad);
        }
      }

      if (isLoggedIn.value && currentRoute.value !== "/login") {
        const license = await store.dispatch("license/get");
        if (!license || license.expired) {
          store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.license);
          store.dispatch("layout/setLayout", "appLayout");
          router.push("/license");
        }
        store.dispatch("layout/setLayout", "appLayout");
      }
    });

    return {
      layout,
    };
  },
});
</script>
