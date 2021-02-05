<template>
  <fragment>
    <v-row>
      <v-col
        v-if="!loggedInNamespace && isEnterprise"
      >
        <v-btn
          class="v-btn--active float-right mr-3"
          text
          small
          @click="addNamespace"
        >
          Add Namespace
        </v-btn>
      </v-col>

      <v-col
        v-else
      >
        <v-menu
          v-show="displayMenu"
          :close-on-content-click="true"
          offset-y
        >
          <template #activator="{ on }">
            <v-chip
              v-show="loggedInNamespace"
              class="float-right"
              @click="openMenu"
              v-on="on"
            >
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
            <v-subheader>Tenant ID</v-subheader>

            <v-list
              class="pt-0 pb-0 mx-2"
            >
              <v-list-item>
                <v-row
                  justify="center"
                  align="center"
                >
                  <v-chip>
                    <v-list-item-title>
                      <span data-test="tenantID-text">{{ tenant }}</span>
                    </v-list-item-title>
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
                </v-row>
              </v-list-item>
            </v-list>

            <v-divider />

            <v-list
              class="pt-0"
            >
              <v-subheader>Namespaces</v-subheader>

              <v-list-item-group>
                <v-virtual-scroll
                  :height="150"
                  item-height="50"
                  :items="availableNamespaces"
                >
                  <template #default="{ item }">
                    <v-list-item
                      :key="item.tenant_id"
                      @click="switchIn(item.tenant_id)"
                    >
                      <v-list-item-icon>
                        <v-icon>mdi-login</v-icon>
                      </v-list-item-icon>
                      <v-list-item-content>
                        <v-list-item-title>
                          {{ item.name }}
                        </v-list-item-title>
                      </v-list-item-content>
                    </v-list-item>
                  </template>
                </v-virtual-scroll>
              </v-list-item-group>
            </v-list>

            <v-divider />

            <v-list
              class="pt-0 pb-0"
            >
              <v-list-item
                v-show="isEnterprise"
                @click="dialog=!dialog"
              >
                <v-list-item-icon>
                  <v-icon>mdi-plus-box</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  Create Namespace
                </v-list-item-content>
              </v-list-item>

              <v-divider />

              <v-list-item
                to="/settings/namespace-manager"
              >
                <v-list-item-icon>
                  <v-icon>mdi-cog</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  Settings
                </v-list-item-content>
              </v-list-item>
            </v-list>
          </v-card>
        </v-menu>
      </v-col>
      <NamespaceAdd
        :show.sync="dialog"
        :first-namespace="first"
      />
    </v-row>
  </fragment>
</template>

<script>
import NamespaceAdd from '@/components/app_bar/namespace/NamespaceAdd';

export default {
  name: 'NamespaceMenu',

  components: {
    NamespaceAdd,
  },

  props: {
    inANamespace: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      model: true,
      dialog: false,
      displayMenu: false,
      first: false,
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },

    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    namespaces() {
      return this.$store.getters['namespaces/list'];
    },

    availableNamespaces() {
      return this.namespaces.filter((ns) => ns.tenant_id !== this.namespace.tenant_id);
    },

    loggedInNamespace() {
      return this.$props.inANamespace;
    },

    tenant() {
      return localStorage.getItem('tenant');
    },

    isEnterprise() {
      return this.$env.isEnterprise;
    },
  },

  watch: {
    dialog(value) {
      if (!value) {
        this.model = false;
      }
    },
  },

  async created() {
    await this.getNamespaces();
    await this.getNamespace();
  },

  methods: {
    addNamespace() {
      this.dialog = !this.dialog;
      this.first = true;
    },

    async openMenu() {
      if (!this.displayMenu) {
        await this.getNamespaces();
      }
      this.displayMenu = !this.displayMenu;
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
      } catch (e) {
        switch (true) {
        case (e.response.status === 404): { // detects namespace inserted
          const namespaceFind = this.namespaces[0];
          if (this.tenant === '' && namespaceFind !== undefined) {
            this.switchIn(namespaceFind.tenant_id);
          }
          break;
        }
        case (e.response.status === 500 && this.tenant === null): {
          break;
        }
        default: {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceLoad);
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
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceList);
        }
        }
      }
    },

    async switchIn(tenant) {
      try {
        await this.$store.dispatch('namespaces/switchNamespace', {
          tenant_id: tenant,
        });

        const isOwner = this.$store.getters['namespaces/get'].owner === this.$store.getters['auth/id'];
        this.$store.dispatch('namespaces/setOwnerStatus', isOwner);

        window.location.reload();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.namespaceSwitch);
      }
    },
  },
};
</script>
