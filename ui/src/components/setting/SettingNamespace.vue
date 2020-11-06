<template>
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
          <h3
            class="ml-6"
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
                rules="required"
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
                      @click="passes(rename)"
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
              <NamespaceNewMember />
            </v-col>
          </v-row>
          <div
            class="mt-6 pl-4 pr-4"
          >
            <v-list>
              <v-list-item
                v-for="item in namespaceNames"
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
                    <v-list-item-title>
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
          </div>
          <v-divider class="mt-6" />
          <v-divider class="mb-6" />
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
              <NamespaceDelete />
            </v-col>
          </v-row>
        </v-col>
      </v-row>
    </v-container>
  </v-form>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import NamespaceNewMember from '@/components/namespace/NamespaceNewMember';
import NamespaceDelete from '@/components/namespace/NamespaceDelete';

export default {
  name: 'SettingNamespace',

  components: {
    ValidationProvider,
    ValidationObserver,
    NamespaceNewMember,
    NamespaceDelete,
  },

  data() {
    return {
      name: '',
      namespaceNames: ['user1', 'user2', 'user3', 'user4', 'user5', 'user6', 'user7'],
    };
  },

  methods: {
    rename() {
      console.log('Namespace renamed');
    },
  },
};
</script>
