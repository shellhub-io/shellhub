<template>
  <fragment>
    <v-form>
      <v-container>
        <v-row
          align="center"
          justify="center"
          class="mt-4"
        >
          <v-col
            sm="8"
          >
            <div
              v-show="!isOwner"
              style="text-align:center"
            >
              <h3
                class="pl-6"
              >
                You're not the owner of this namespace.
              </h3>
              <br>
            </div>
            <div
              class="mt-6 pl-4 pr-4"
            >
              <v-row>
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
            </div>

            <v-divider />
            <v-divider />

            <div
              v-show="isOwner"
              id="editOperation"
            >
              <h3
                class="ml-6 mt-8"
              >
                Edit namespace
              </h3>
              <ValidationObserver
                ref="data"
                v-slot="{ passes }"
              >
                <div
                  class="mt-6 pl-4 pr-4"
                >
                  <ValidationProvider
                    v-slot="{ errors }"
                    ref="providerName"
                    vid="name"
                    name="Priority"
                    rules="required|rfc1123"
                  >
                    <v-row>
                      <v-col
                        cols="8"
                        class="ml-3"
                      >
                        <v-text-field
                          v-model="name"
                          label="Name"
                          :error-messages="errors"
                          required
                          data-test="name-text"
                        />
                      </v-col>
                      <v-col
                        cols="2"
                        class="mt-2"
                      >
                        <v-btn
                          outlined
                          @click="passes(editNamespace)"
                        >
                          Rename Namespace
                        </v-btn>
                      </v-col>
                    </v-row>
                  </ValidationProvider>
                </div>
              </ValidationObserver>
              <v-divider class="mt-6" />
              <v-divider class="mb-6" />
            </div>
            <div
              v-show="show"
              id="userOperation"
            >
              <v-row
                class="ml-3"
              >
                <v-col>
                  <h3>
                    Members
                  </h3>
                </v-col>
                <v-spacer />
                <v-col>
                  <NamespaceNewMember :ns-tenant="tenant" />
                </v-col>
              </v-row>
              <div
                class="mt-6 pl-4 pr-4"
              >
                <v-list>
                  <v-list-item
                    v-for="item in namespace.member_names"
                    :key="item"
                  >
                    <v-row>
                      <v-col>
                        <v-icon>
                          mdi-account
                        </v-icon>
                      </v-col>
                      <v-col
                        class="mt-1 mr-10"
                      >
                        <v-list-item-title
                          :data-test="item"
                        >
                          {{ item }}
                        </v-list-item-title>
                      </v-col>
                      <v-spacer />
                      <v-col
                        class="ml-10"
                      >
                        <v-btn
                          class="ml-10"
                          outlined
                          @click="remove(item)"
                        >
                          <v-tooltip
                            bottom
                          >
                            <template #activator="{ on }">
                              <v-icon
                                v-on="on"
                              >
                                delete
                              </v-icon>
                            </template>
                            <span>
                              Remove user
                            </span>
                          </v-tooltip>
                        </v-btn>
                      </v-col>
                    </v-row>
                  </v-list-item>
                </v-list>
                <v-divider class="mt-6" />
                <v-divider class="mb-6" />
              </div>
            </div>
            <div
              v-show="isOwner"
              id="deleteOperation"
            >
              <h3
                class="pl-6"
              >
                Danger Zone
              </h3>
              <br>
              <v-row
                class="ml-3"
              >
                <v-col
                  cols="8"
                >
                  <p
                    class="mt-4"
                  >
                    Delete this namespace
                  </p>
                </v-col>
                <v-col
                  cols="2"
                  class="mt-2"
                >
                  <NamespaceDelete :ns-tenant="tenant" />
                </v-col>
              </v-row>
              <v-divider class="mt-6" />
              <v-divider class="mb-6" />
            </div>
            <SettingSecurity
              :show="show"
            />
          </v-col>
        </v-row>
      </v-container>
    </v-form>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import SettingSecurity from '@/components/setting/SettingSecurity';
import NamespaceNewMember from '@/components/namespace/NamespaceNewMember';
import NamespaceDelete from '@/components/namespace/NamespaceDelete';

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
      return this.owner === this.$store.getters['auth/id'];
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

    show() {
      return this.$env.isHosted && this.isOwner;
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
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceLoad);
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
