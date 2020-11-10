<template>
  <fragment>
    <v-menu
      :close-on-content-click="false"
      offset-y
    >
      <template #activator="{ on }">
        <v-chip v-on="on">
          <v-icon
            left
          >
            mdi-server
          </v-icon>
          {{ namespace.name }}
          <v-icon right>
            mdi-chevron-down
          </v-icon>
        </v-chip>
      </template>
      <v-card>
        <v-list-item three-line>
          <v-list-item-content>
            <v-list-item-title
              data-test="tenantID-field"
              class="mb-1"
            >
              Tenant ID
            </v-list-item-title>
            <v-list-item-subtitle>
              <v-chip>
                <span
                  data-test="tenantID-text"
                >
                  {{ tenant }}
                </span>
                <v-icon
                  v-clipboard="tenant"
                  v-clipboard:success="() => {
                    this.$store.dispatch('snackbar/showSnackbarCopy', this.$copy.tenantId);
                  }"
                  right
                >
                  mdi-content-copy
                </v-icon>
              </v-chip>
            </v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
        <v-list-group>
          <template #activator>
            <v-list-item-content>
              <v-list-item-title>
                Namespace List
              </v-list-item-title>
            </v-list-item-content>
          </template>
          <v-list-item
            v-show="show"
          >
            <v-list-item-title>
              <NamespaceAdd />
            </v-list-item-title>
          </v-list-item>
          <v-virtual-scroll
            :height="150"
            item-height="40"
            :items="namespaces"
          >
            <template #default="{ item }">
              <v-list-item :key="item.tenant_id">
                <v-row>
                  <v-col
                    v-if="formatName(item.name).mode"
                    class="mt-1"
                  >
                    <v-tooltip
                      bottom
                    >
                      <template
                        #activator="{ on }"
                      >
                        <v-list-item-title
                          v-on="on"
                        >
                          {{ formatName(item.name).name }}
                        </v-list-item-title>
                      </template>
                      <span>
                        {{ item.name }}
                      </span>
                    </v-tooltip>
                  </v-col>
                  <v-col
                    v-else
                  >
                    <v-list-item-title>
                      {{ item.name }}
                    </v-list-item-title>
                  </v-col>
                  <v-spacer />
                  <v-spacer />
                  <v-col>
                    <v-btn
                      small
                      class="v-btn--active"
                      text
                      color="primary"
                      @click="switchIn(item.tenant_id)"
                    >
                      <v-tooltip
                        bottom
                      >
                        <template #activator="{ on }">
                          <v-icon
                            v-on="on"
                          >
                            mdi-sync
                          </v-icon>
                        </template>
                        <span>
                          Switch namespace
                        </span>
                      </v-tooltip>
                    </v-btn>
                  </v-col>
                </v-row>
              </v-list-item>
            </template>
          </v-virtual-scroll>
        </v-list-group>
        <v-list-item
          key="settings"
        >
          <v-list-item-title>
            <v-btn
              to="/settings/namespace-manager"
              class="v-btn--active"
              text
              color="primary"
              small
            >
              Namespace manager
            </v-btn>
          </v-list-item-title>
        </v-list-item>
      </v-card>
    </v-menu>
  </fragment>
</template>

<script>

import NamespaceAdd from '@/components/namespace/NamespaceAdd';

export default {
  name: 'NamespaceMenu',

  components: {
    NamespaceAdd,
  },

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    namespaces() {
      return this.$store.getters['namespaces/list'];
    },

    tenant() {
      return localStorage.getItem('tenant');
    },

    show() {
      return this.$env.isHosted;
    },
  },

  created() {
    this.getNamespaces();
  },

  methods: {
    async getNamespaces() {
      try { // load namespaces
        await this.$store.dispatch('namespaces/fetch');
        await this.$store.dispatch('namespaces/get', this.tenant);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceList);
      }
    },

    async switchIn(tenant) {
      try {
        await this.$store.dispatch('namespaces/switchNamespace', {
          tenant_id: tenant,
        });
        localStorage.setItem('tenant', tenant);
        window.location.reload();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceSwitch);
      }
    },

    formatName(name, limit = 11) {
      const formatObj = { name, mode: true };
      if (name.length > limit) formatObj.name = `${name.slice(0, limit - 3)}...`; formatObj.format = false;
      return formatObj;
    },
  },
};

</script>
