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
            <ValidationObserver
              ref="obs"
              v-slot="{ passes }"
            >
              <v-row>
                <v-col>
                  <h3>
                    Namespace
                  </h3>
                </v-col>

                <v-spacer />

                <v-col
                  md="auto"
                  class="ml-auto"
                >
                  <v-tooltip
                    bottom
                    :disabled="hasAuthorizationRenameNamespace"
                  >
                    <template #activator="{ on }">
                      <div v-on="on">
                        <v-btn
                          :disabled="!hasAuthorizationRenameNamespace"
                          outlined
                          @click="passes(editNamespace)"
                        >
                          Rename Namespace
                        </v-btn>
                      </div>
                    </template>

                    <span>
                      You don't have this kind of authorization.
                    </span>
                  </v-tooltip>
                </v-col>
              </v-row>

              <div class="mt-4 mb-2">
                <ValidationProvider
                  v-slot="{ errors }"
                  ref="providerName"
                  vid="name"
                  name="Priority"
                  rules="required|rfc1123|noDot|namespace"
                >
                  <v-text-field
                    v-model="name"
                    class="ml-3"
                    label="Name"
                    :error-messages="errors"
                    required
                    data-test="name-text"
                  />
                </ValidationProvider>
              </div>
            </ValidationObserver>

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
                  <NamespaceMemberFormDialog
                    :add-user="true"
                    :show.sync="namespaceMemberFormShow"
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

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import SettingSecurity from '@/components/setting/SettingSecurity';
import NamespaceMemberList from '@/components/app_bar/namespace/NamespaceMemberList';
import NamespaceMemberFormDialog from '@/components/app_bar/namespace/NamespaceMemberFormDialog';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'SettingNamespaceComponent',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
    NamespaceDelete,
    NamespaceMemberList,
    NamespaceMemberFormDialog,
    SettingSecurity,
  },

  data() {
    return {
      name: '',
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

    hasAuthorizationRenameNamespace() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace.rename,
        );
      }

      return false;
    },

    hasAuthorizationRemoveUser() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace.removeMember,
        );
      }

      return false;
    },

    hasAuthorizationDeleteNamespace() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace.remove,
        );
      }

      return false;
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

    async editNamespace() {
      try {
        await this.$store.dispatch('namespaces/put', { id: this.tenant, name: this.name });
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceEdit);
      } catch (error) {
        if (error.response.status === 400) {
          this.$refs.obs.setErrors({
            namespace: this.$errors.form.invalid('namespace', 'nonStandardCharacters'),
          });
        } else if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            namespace: this.$errors.form.invalid('namespace', 'nameUsed'),
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceEdit);
        }
      }
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.name = this.namespace.name;
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
