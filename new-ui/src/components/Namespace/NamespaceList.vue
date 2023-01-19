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

<script lang="ts">
import { defineComponent, computed } from "vue";
import { useStore } from "../../store";
import { INotificationsError } from "../../interfaces/INotifications";

export default defineComponent({
  inheritAttrs: false,
  setup() {
    const store = useStore();

    const namespace = computed(() => store.getters["namespaces/get"]);

    const namespaces = computed(() => store.getters["namespaces/list"].filter(
      (el: any) => el.name !== namespace.value.name,
    ));

    const switchIn = async (tenant: string) => {
      try {
        await store.dispatch("namespaces/switchNamespace", {
          tenant_id: tenant,
        });

        window.location.reload();
      } catch (error: any) {
        store.dispatch(
          "snackbar/showSnackbarErrorLoading",
          INotificationsError.namespaceSwitch,
        );
        throw new Error(error);
      }
    };
    return {
      namespace,
      namespaces,
      switchIn,
    };
  },
});
</script>
