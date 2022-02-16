<template>
  <fragment>
    <v-card-text class="pa-0">
      <v-data-table
        :headers="headers"
        :items="getPublicKeys"
        item-key="fingerprint"
        :sort-by="['started_at']"
        :sort-desc="[true]"
        :items-per-page="10"
        :footer-props="{'items-per-page-options': [10, 25, 50, 100]}"
        :server-items-length="getNumberPublicKeys"
        :options.sync="pagination"
        :disable-sort="true"
        data-test="publicKeyList-dataTable"
      >
        <template #[`item.name`]="{ item }">
          {{ item.name }}
        </template>

        <template #[`item.fingerprint`]="{ item }">
          {{ item.fingerprint }}
        </template>

        <template #[`item.filter`]="{ item }">
          <div
            v-if="filterKey(item.filter)=='hostname'"
          >
            {{ formatHostnameFilter(item.filter) }}
          </div>

          <div v-else-if="filterKey(item.filter)=='tags'">
            <v-tooltip
              v-for="(tag, index) in item.filter.tags"
              :key="index"
              bottom
              :disabled="!showTag(tag)"
            >
              <template #activator="{ on, attrs }">
                <v-chip
                  class="ml-1 mb-1"
                  small
                  outlined
                  v-bind="attrs"
                  v-on="on"
                >
                  {{ displayOnlyTenCharacters(tag) }}
                </v-chip>
              </template>

              <span v-if="showTag(tag)">
                {{ tag }}
              </span>
            </v-tooltip>
          </div>
        </template>

        <template #[`item.username`]="{ item }">
          {{ item.username === '' ? 'All users' : item.username }}
        </template>

        <template #[`item.created_at`]="{ item }">
          {{ item.created_at | moment("ddd, MMM Do YY, h:mm:ss a") }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            :ref="'menu'+getPublicKeys.indexOf(item)"
            offset-y
          >
            <template #activator="{ on, attrs }">
              <v-chip
                color="transparent"
                v-on="on"
              >
                <v-icon
                  small
                  class="icons"
                  v-bind="attrs"
                  v-on="on"
                >
                  mdi-dots-horizontal
                </v-icon>
              </v-chip>
            </template>

            <v-card>
              <v-tooltip
                bottom
                :disabled="hasAuthorizationFormDialogEdit"
              >
                <template #activator="{ on, attrs }">
                  <div
                    v-bind="attrs"
                    v-on="on"
                  >
                    <v-list-item
                      :disabled="!hasAuthorizationFormDialogEdit"
                      @click="showPublicKeyFormDialog(getPublicKeys.indexOf(item))"
                    >
                      <PublicKeyFormDialogEdit
                        :key-object="item"
                        :show.sync="publicKeyFormDialogShow[getPublicKeys.indexOf(item)]"
                        @update="refresh"
                      />
                    </v-list-item>
                  </div>
                </template>

                <span>
                  You don't have this kind of authorization.
                </span>
              </v-tooltip>

              <v-tooltip
                bottom
                :disabled="hasAuthorizationFormDialogRemove"
              >
                <template #activator="{ on, attrs }">
                  <div
                    v-bind="attrs"
                    v-on="on"
                  >
                    <v-list-item
                      :disabled="!hasAuthorizationFormDialogRemove"
                      @click="showPublicKeyDelete(getPublicKeys.indexOf(item))"
                    >
                      <PublicKeyDelete
                        :fingerprint="item.fingerprint"
                        :show.sync="publicKeyDeleteShow[getPublicKeys.indexOf(item)]"
                        @update="refresh"
                      />
                    </v-list-item>
                  </div>
                </template>

                <span>
                  You don't have this kind of authorization.
                </span>
              </v-tooltip>
            </v-card>
          </v-menu>
        </template>
      </v-data-table>
    </v-card-text>
  </fragment>
</template>

<script>

import PublicKeyFormDialogEdit from '@/components/public_key/PublicKeyFormDialogEdit';
import PublicKeyDelete from '@/components/public_key/PublicKeyDelete';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'PublicKeyListComponent',

  filters: { hasPermission },

  components: {
    PublicKeyFormDialogEdit,
    PublicKeyDelete,
  },

  data() {
    return {
      pagination: {},
      publicKeyFormDialogShow: [],
      publicKeyDeleteShow: [],
      editAction: 'edit',
      removeAction: 'remove',
      headers: [
        {
          text: 'Name',
          value: 'name',
          align: 'center',
        },
        {
          text: 'Fingerprint',
          value: 'fingerprint',
          align: 'center',
        },
        {
          text: 'Filter',
          value: 'filter',
          align: 'center',
        },
        {
          text: 'Username',
          value: 'username',
          align: 'center',
        },
        {
          text: 'Created At',
          value: 'created_at',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
        },
      ],
    };
  },

  computed: {
    getPublicKeys() {
      return this.$store.getters['publickeys/list'];
    },

    getNumberPublicKeys() {
      return this.$store.getters['publickeys/getNumberPublicKeys'];
    },

    hasAuthorizationFormDialogEdit() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.publicKey[this.editAction],
        );
      }

      return false;
    },

    hasAuthorizationFormDialogRemove() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.publicKey[this.removeAction],
        );
      }

      return false;
    },
  },

  watch: {
    pagination: {
      handler() {
        this.getPublicKeysList();
      },
      deep: true,
    },
  },

  methods: {
    showTag(str) {
      if (str !== undefined) {
        if (str.length > 10) {
          return true;
        }
      }
      return false;
    },

    filterKey(filter) {
      return Reflect.ownKeys(filter)[0];
    },

    hasKeys(filter) {
      return Object.keys(filter).length > 0;
    },

    displayOnlyTenCharacters(str) {
      if (str !== undefined) {
        if (str.length > 10) return `${str.substr(0, 10)}...`;
      }
      return str;
    },

    refresh() {
      this.getPublicKeysList();
    },

    async getPublicKeysList() {
      if (!this.$store.getters['boxs/getStatus']) {
        const data = {
          perPage: this.pagination.itemsPerPage,
          page: this.pagination.page,
        };

        try {
          await this.$store.dispatch('publickeys/fetch', data);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.publicKeyList);
        }
      } else {
        this.setArrays();
        this.$store.dispatch('boxs/setStatus', false);
      }
    },

    showPublicKeyFormDialog(index) {
      this.publicKeyFormDialogShow[index] = this.publicKeyFormDialogShow[index] === undefined
        ? true : !this.publicKeyFormDialogShow[index];
      this.$set(this.publicKeyFormDialogShow, index, this.publicKeyFormDialogShow[index]);

      this.closeMenu(index);
    },

    showPublicKeyDelete(index) {
      this.publicKeyDeleteShow[index] = this.publicKeyDeleteShow[index] === undefined
        ? true : !this.publicKeyDeleteShow[index];
      this.$set(this.publicKeyDeleteShow, index, this.publicKeyDeleteShow[index]);

      this.closeMenu(index);
    },

    formatHostnameFilter(filter) {
      return filter.hostname === '.*' ? 'All devices' : filter.hostname;
    },

    setArrays() {
      const numberPublicKey = this.getPublicKeys.length;

      if (numberPublicKey > 0) {
        this.publicKeyFormDialogShow = new Array(numberPublicKey).fill(false);
        this.publicKeyDeleteShow = new Array(numberPublicKey).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};

</script>
