<template>
  <v-list-item
    v-for="item in namespaces"
    :key="item.tenant_id"
    link
    @click="switchIn(item.tenant_id)"
  >
    <div>
      <v-list-item-title :data-test="item.name + '-namespace'">
        {{ item.name }}
      </v-list-item-title>
    </div>
  </v-list-item>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";
import { INamespace } from "@/interfaces/INamespace";
import useSnackbar from "@/helpers/snackbar";

defineOptions({
  inheritAttrs: false,
});

const store = useStore();
const snackbar = useSnackbar();
const namespace = computed(() => store.getters["namespaces/get"]);

const namespaces = computed(() => store.getters["namespaces/list"].filter(
  (el: INamespace) => el.name !== namespace.value.name,
));

const switchIn = async (tenant: string) => {
  try {
    await store.dispatch("namespaces/switchNamespace", {
      tenant_id: tenant,
    });

    window.location.reload();
  } catch (error: unknown) {
    snackbar.showError("An error occurred while switching namespaces.");
    handleError(error);
  }
};
</script>
