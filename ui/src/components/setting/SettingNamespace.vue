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
            v-if="!isOwner"
            style="text-align:center"
            data-test="notTheOwner"
          >
            <h3 class="pl-6">
              You're not the owner of this namespace.
            </h3>
            <br>
          </div>

          <div class="mt-6">
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
                    <span data-test="tenant">
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
                </v-card>
              </v-col>
            </v-row>

            <v-divider />
            <v-divider />
          </div>

          <div
            v-if="isOwner"
            class="mt-6"
            data-test="editOperation"
          >
            <ValidationObserver
              ref="data"
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
                  <v-btn
                    outlined
                    @click="passes(editNamespace)"
                  >
                    Rename Namespace
                  </v-btn>
                </v-col>
              </v-row>

              <div class="mt-4 mb-2">
                <ValidationProvider
                  v-slot="{ errors }"
                  ref="providerName"
                  vid="name"
                  name="Priority"
                  rules="required|rfc1123"
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
            v-if="isOwner"
            data-test="userOperation"
            class="mt-6"
          >
            <v-row>
              <v-col>
                <h3>
                  Members
                </h3>
              </v-col>

              <v-spacer />

              <div
                v-if="isEnterpriseOwner"
                data-test="new-member"
              >
                <v-col
                  md="auto"
                  class="ml-auto"
                >
                  <NamespaceNewMember :ns-tenant="tenant" />
                </v-col>
              </div>
            </v-row>

            <div class="mt-5">
              <v-list>
                <v-list-item
                  v-for="item in namespace.members"
                  :key="item.id"
                >
                  <v-row>
                    <v-col
                      md="auto"
                      class="ml-auto"
                    >
                      <v-icon>
                        mdi-account
                      </v-icon>
                    </v-col>

                    <v-col>
                      <v-list-item-title :data-test="item.name">
                        {{ item.name }}
                      </v-list-item-title>
                    </v-col>

                    <v-spacer />

                    <div
                      v-if="isEnterpriseOwner"
                    >
                      <v-col
                        md="auto"
                        class="ml-auto"
                      >
                        <v-btn
                          v-if="item.id!==owner"
                          data-test="remove-member"
                          outlined
                          @click="remove(item.name)"
                        >
                          <v-tooltip
                            bottom
                          >
                            <template #activator="{ on }">
                              <v-icon v-on="on">
                                delete
                              </v-icon>
                            </template>
                            <span>
                              Remove user
                            </span>
                          </v-tooltip>
                        </v-btn>

                        <p
                          v-else
                          data-test="owner"
                          class="mr-3"
                        >
                          Owner
                        </p>
                      </v-col>
                    </div>
                    <div
                      v-else
                      data-test="role"
                    >
                      <v-col
                        md="auto"
                        class="ml-auto"
                      >
                        <p data-test="role-text">
                          {{ item.id === owner ? 'Owner' : 'Member' }}
                        </p>
                      </v-col>
                    </div>
                  </v-row>
                </v-list-item>
              </v-list>
            </div>

            <v-divider />
            <v-divider />
          </div>

          <div
            v-if="isEnterpriseOwner"
            class="mt-6"
            data-test="securityOperation"
          >
            <SettingSecurity />

            <v-divider />
            <v-divider />
          </div>

          <div
            v-if="isOwner"
            class="mt-6"
            data-test="deleteOperation"
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
                <NamespaceDelete :ns-tenant="tenant" />
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
import NamespaceNewMember from '@/components/app_bar/namespace/NamespaceNewMember';
import NamespaceDelete from '@/components/app_bar/namespace/NamespaceDelete';

export default {
  name: 'SettingNamespace',

  components: {
    ValidationProvider,
    ValidationObserver,
    NamespaceNewMember,
    NamespaceDelete,
    SettingSecurity,
  },

  data() {
    return {
      name: '',
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },

    owner() {
      return this.$store.getters['namespaces/get'].owner;
    },

    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    tenant() {
      return localStorage.getItem('tenant');
    },

    isEnterpriseOwner() {
      return this.$env.isEnterprise && this.isOwner;
    },
  },

  async created() {
    await this.getNamespace();
    this.name = this.namespace.name;
  },

  methods: {
    async editNamespace() {
      try {
        await this.$store.dispatch('namespaces/put', { id: this.tenant, name: this.name });
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceEdit);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceEdit);
      }
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
      } catch (e) {
        if (e.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceLoad);
        }
      }
    },

    async remove(username) {
      try {
        await this.$store.dispatch('namespaces/removeUser', {
          username,
          tenant_id: this.tenant,
        });
        this.dialog = false;
        this.username = '';
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceRemoveUser);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceRemoveUser);
      }
    },
  },
};

</script>
