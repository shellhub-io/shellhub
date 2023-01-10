<template>
  <fragment>
    <v-list v-if="hasNamespace">
      <v-list-group
        v-model="listing"
      >
        <template #activator>
          <v-list-item-content>
            <v-list-item-title class="primary--text primary--icon">
              {{ namespace.name }}
            </v-list-item-title>
          </v-list-item-content>
        </template>

        <v-icon
          slot="appendIcon"
          color="primary"
        >
          mdi-chevron-down
        </v-icon>

        <NamespaceList data-test="namespaceList-component" />

        <v-list-item v-if="isEnterprise">
          <NamespaceAdd data-test="namespaceAdd-component" />
        </v-list-item>
      </v-list-group>
    </v-list>

    <div v-else>
      <NamespaceAdd
        data-test="namespaceAddNoNamespace-component"
      />
    </div>
  </fragment>
</template>

<script>

import NamespaceList from '@/components/namespace/NamespaceList';
import NamespaceAdd from '@/components/namespace/NamespaceAdd';

export default {
  name: 'NamespaceMenuComponent',

  components: {
    NamespaceList,
    NamespaceAdd,
  },

  data() {
    return {
      inANamespace: false,
      listing: false,
      isChecking: false,
    };
  },

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    hasNamespace() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    tenant() {
      return localStorage.getItem('tenant');
    },

    isEnterprise() {
      return this.$env.isEnterprise;
    },

    openVersion() {
      return !this.$env.isEnterprise;
    },
  },

  watch: {
    hasNamespace(status) {
      this.inANamespace = status;
      this.getNamespace();
    },

    listing(val) {
      if (val) {
        this.getNamespaces();
      }
    },
  },

  async created() {
    await this.getNamespaces();
    if (this.inANamespace) {
      await this.getNamespace();
    }
    if (Object.keys(this.namespace).length === 0 && this.openVersion) {
      this.isChecking = true;
      // Interval to check if the namespace has been added by cli
      setInterval(() => {
        this.checkNewNamespace();
      }, 3000);
    }
  },

  methods: {
    async getNamespace() {
      if (this.isChecking) return;

      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
      } catch (error) {
        switch (true) {
        case (error.response.status === 404): { // detects namespace inserted
          const namespaceFind = this.$store.getters['namespaces/list'][0];
          if (this.tenant === '' && namespaceFind !== undefined) {
            this.switchIn(namespaceFind.tenant_id);
          }
          break;
        }
        case (error.response.status === 500 && this.tenant === null): {
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceLoad);
        }
        }
      }
    },

    async getNamespaces() {
      try {
        await this.$store.dispatch('namespaces/fetch');
      } catch (e) {
        switch (true) {
        case (!this.inANamespace && e.response.status === 403): { // dialog pops
          break;
        }
        case (e.response.status === 403): {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.namespaceList);
        }
        }
      }
    },

    async checkNewNamespace() {
      if (!this.$store.getters['auth/isLoggedIn']) return;

      await this.$store.dispatch('namespaces/fetch', {
        page: 1,
        perPage: 10,
        fitler: '',
      });
      if (this.$store.getters['namespaces/list'].length > 0) {
        this.switchIn(this.$store.getters['namespaces/list'][0].tenant_id);
      }
    },

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
