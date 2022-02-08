<template>
  <fragment>
    <v-container>
      <v-row
        align="center"
        justify="center"
        class="mt-4"
      >
        <v-col sm="8">
          <div
            class="mt-6"
            data-test="tenant-div"
          >
            <v-row class="mb-2">
              <v-col md="auto">
                <v-card
                  tile
                  :elevation="0"
                >
                  Tenant ID:
                </v-card>
              </v-col>

              <v-col
                md="auto"
                class="ml-auto"
              >
                <v-card
                  class="auto"
                  tile
                  :elevation="0"
                >
                  <v-chip>
                    <span>
                      {{ tenant }}
                    </span>

                    <v-icon
                      v-clipboard="tenant"
                      v-clipboard:success="() => {
                        $store.dispatch('snackbar/showSnackbarCopy', $copy.tenantId);
                      }"
                      right
                    >
                      mdi-content-copy
                    </v-icon>
                  </v-chip>
                </v-card>
              </v-col>
            </v-row>

            <v-divider />
            <v-divider />
          </div>

          <div
            class="mt-6"
            data-test="editOperation-div"
          >
            <NamespaceRename data-test="namespaceRename-component" />

            <v-divider />
            <v-divider />
          </div>

          <div
            class="mt-6"
            data-test="userOperation-div"
          >
            <v-row>
              <v-col>
                <h3>
                  Members
                </h3>
              </v-col>

              <v-spacer />

              <div>
                <v-col
                  md="auto"
                  class="ml-auto"
                >
                  <NamespaceMemberFormDialogAdd
                    data-test="namespaceMemberFormDialogAdd-component"
                    @update="refresh"
                  />
                </v-col>
              </div>
            </v-row>

            <NamespaceMemberList
              :namespace.sync="namespace"
            />
            <v-divider />
            <v-divider />
          </div>

          <div
            v-if="isEnterprise"
            class="mt-6"
            data-test="securityOperation-div"
          >
            <SettingSecurity :has-tenant="hasTenant()" />

            <v-divider />
            <v-divider />
          </div>

          <div
            class="mt-6"
            data-test="deleteOperation-div"
          >
            <h3 class="mb-5">
              Danger Zone
            </h3>

            <v-row class="mt-4 mb-2">
              <v-col class="ml-3">
                <h4>
                  Delete this namespace
                </h4>
                <div class="ml-2">
                  <p>
                    After deleting a namespace, there is no going back. Be sure.
                  </p>
                </div>
              </v-col>

              <v-col
                md="auto"
                class="ml-auto mb-4"
              >
                <NamespaceDelete
                  :ns-tenant="tenant"
                  data-test="namespaceDelete-component"
                />
              </v-col>
            </v-row>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </fragment>
</template>

<script>

import SettingSecurity from '@/components/setting/SettingSecurity';
import NamespaceMemberList from '@/components/namespace/NamespaceMemberList';
import NamespaceRename from '@/components/namespace/NamespaceRename';
import NamespaceMemberFormDialogAdd from '@/components/namespace/NamespaceMemberFormDialogAdd';
import NamespaceDelete from '@/components/namespace/NamespaceDelete';

export default {
  name: 'SettingNamespaceComponent',

  components: {
    NamespaceDelete,
    NamespaceMemberList,
    NamespaceRename,
    NamespaceMemberFormDialogAdd,
    SettingSecurity,
  },

  data() {
    return {
      namespaceMemberFormShow: false,
    };
  },

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    isEnterprise() {
      return this.$env.isEnterprise;
    },
  },

  async created() {
    if (this.hasTenant()) {
      await this.getNamespace();
    }
  },

  methods: {
    refresh() {
      this.getNamespace();
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceLoad);
        }
      }
    },

    async remove(userId) {
      try {
        await this.$store.dispatch('namespaces/removeUser', {
          user_id: userId,
          tenant_id: this.tenant,
        });
        this.refresh();
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceRemoveUser);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceRemoveUser);
      }
    },

    hasTenant() {
      return this.tenant !== '';
    },

    countDevicesHasNamespace() {
      return this.$store.getters['namespaces/get'].devices_count;
    },

    countDevicesHasNamespacePercent() {
      const maxDevices = this.$store.getters['namespaces/get'].max_devices;

      let percent = 0;
      if (maxDevices >= 0) {
        percent = (this.countDevicesHasNamespace() / maxDevices) * 100;
        return { maxDevices, percent };
      }
      return { maxDevices, percent };
    },
  },
};

</script>
