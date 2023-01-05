<template>
  <v-card class="bg-v-theme-surface">
    <v-tabs background-color="secondary" stacked color="primary" align-tabs="center">
      <v-tab
        v-for="item in visibleItems"
        :key="item.title"
        :to="item.path"
        :data-test="item.title + '-tab'"
      >
        {{ item.title }}
      </v-tab>
    </v-tabs>

    <v-divider />
  </v-card>

  <v-card class="bg-v-theme-surface">
    <router-view />
  </v-card>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { envVariables } from "../envVariables";
import { useStore } from "../store";

export default defineComponent({
  setup() {
    const store = useStore();

    const visibleItems = computed(() => items.value.filter((item) => !item.hidden));

    const currentInANamespace = computed(() => localStorage.getItem("tenant") !== "" );

    const hasNamespace = computed(() => store.getters["namespaces/getNumberNamespaces"] !== 0);

    const items = computed(() => [
      {
        title: "Profile",
        path: "/settings",
      },
      {
        title: "Namespace",
        path: "/settings/namespace-manager",
        hidden: !currentInANamespace.value,
      },
      {
        title: "Private Keys",
        path: "/settings/private-keys",
      },
      {
        title: "Tags",
        path: "/settings/tags",
      },
      {
        title: "Billing",
        path: "/settings/billing",
        hidden: !(
          envVariables.billingEnable &&
          envVariables.isCloud &&
          hasNamespace.value
        ),
      },
    ]);

    return {
      items,
      visibleItems,
    };
  },
});
</script>
