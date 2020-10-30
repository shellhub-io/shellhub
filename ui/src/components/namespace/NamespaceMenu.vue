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
          My Device Fleet
          <v-icon right>
            mdi-chevron-down
          </v-icon>
        </v-chip>
      </template>
      <v-card>
        <v-list-group>
          <template #activator>
            <v-list-item-content>
              <v-list-item-title>
                Namespace List
              </v-list-item-title>
            </v-list-item-content>
          </template>
          <v-virtual-scroll
            :height="adaptHeight"
            item-height="40"
            :items="namespaceNames"
          >
            <template #default="{item}">
              <v-list-item :key="item">
                <v-row>
                  <v-col
                    class="mt-1"
                  >
                    <v-list-item-title>
                      {{ item }}
                    </v-list-item-title>
                  </v-col>
                  <v-spacer />
                  <v-col>
                    <v-btn
                      small
                      class="v-btn--active"
                      text
                      color="primary"
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

export default {
  name: 'NamespaceMenu',

  computed: {
    namespaceNames() {
      return this.$store.getters['namespaces/list'].map((namespace) => namespace.name);
    },

    adaptHeight() {
      return (this.namespaceNames).length > 3 ? 150 : (this.namespaceNames).length * 45;
    },
  },

  created() {
    this.getNamespaces();
  },

  methods: {
    async getNamespaces() {
      try {
        await this.$store.dispatch('namespaces/fetch');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceList);
      }
    },
  },
};

</script>
