<template>
  <v-navigation-drawer
    v-model="showNavigationDrawer"
    :permanent="permanent"
    class="bg-v-theme-surface"
    app
  >
    <v-list
      density="compact"
      :items="visibleItems"
      item-value="path"
      class="pa-0"
    >
      <template #item="{ props }">
        <v-list-item
          :key="props.value"
          :title="props.title"
          :to="props.value"
        />
      </template>
    </v-list>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useDisplay } from "vuetify";
import { envVariables } from "@/envVariables";
import { useStore } from "@/store";

const { lgAndUp } = useDisplay();

const permanent = computed(() => lgAndUp.value);
const showNavigationDrawer = ref(true);
const store = useStore();
const namespacedInstance = computed(() => localStorage.getItem("tenant") !== "");
const hasNamespace = computed(() => store.getters["namespaces/getNumberNamespaces"] !== 0);

const items = computed(() => [
  {
    title: "Profile",
    path: "/settings/profile",
  },
  {
    title: "Namespace",
    path: "/settings/namespace",
    hidden: !namespacedInstance.value,
  },
  {
    title: "Private Keys",
    path: "/settings/private-keys",
  },
  {
    title: "Tags",
    path: "/settings/tags",
    hidden: !namespacedInstance.value,
  },
  {
    title: "Billing",
    path: "/settings/billing",
    hidden: !(
      envVariables.billingEnable
          && envVariables.isCloud
          && hasNamespace.value
    ),
  },
]);

const visibleItems = computed(() => items.value.filter((item) => !item.hidden));

defineExpose({ visibleItems });
</script>
