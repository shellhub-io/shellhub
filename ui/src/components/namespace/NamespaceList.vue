<template>
  <fragment>
    <v-list-item
      v-for="item in namespaces"
      :key="item.tenant_id"
      link
      @click="switchIn(item.tenant_id)"
    >
      <v-list-item-content>
        <v-list-item-title
          :data-test="item.name+'-namespace'"
          v-text="item.name"
        />
      </v-list-item-content>
    </v-list-item>
  </fragment>
</template>

<script>

export default {
  name: 'NamespaceList',

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    namespaces() {
      return this.$store.getters['namespaces/list'].filter((el) => el.name !== this.namespace.name);
    },
  },

  methods: {
    async switchIn(tenant) {
      try {
        await this.$store.dispatch('namespaces/switchNamespace', {
          tenant_id: tenant,
        });

        window.location.reload();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceSwitch);
      }
    },
  },
};
</script>
