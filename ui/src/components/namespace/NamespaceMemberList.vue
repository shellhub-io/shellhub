<template>
  <fragment>
    <div class="mt-5">
      <v-data-table
        class="elevation-0"
        :headers="headers"
        :items="members"
        hide-default-footer
        data-test="dataTable-field"
      >
        <template #[`item.username`]="{ item }">
          <v-icon>
            mdi-account
          </v-icon>
          {{ item.username }}
        </template>

        <template #[`item.role`]="{ item }">
          {{ item.role }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            v-if="item.id !== namespace.owner"
            :ref="'menu'+members.indexOf(item)"
            offset-y
          >
            <template #activator="{ on, attrs}">
              <v-icon
                small
                class="icons"
                v-bind="attrs"
                v-on="on"
              >
                mdi-dots-horizontal
              </v-icon>
            </template>

            <v-list>
              <v-tooltip
                bottom
                :disabled="hasAuthorizationEditMember"
              >
                <template #activator="{ on, attrs }">
                  <div
                    v-bind="attrs"
                    v-on="on"
                  >
                    <v-list-item
                      :disabled="!hasAuthorizationEditMember"
                      @click.stop="showNamespaceMemberFormEdit(members.indexOf(item))"
                    >
                      <NamespaceMemberFormDialogEdit
                        :member="item"
                        :show.sync="namespaceMemberFormShow[members.indexOf(item)]"
                        data-test="NamespaceMemberFormDialogEdit-component"
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
                :disabled="hasAuthorizationRemoveMember"
              >
                <template #activator="{ on, attrs }">
                  <div
                    v-bind="attrs"
                    v-on="on"
                  >
                    <v-list-item
                      :disabled="!hasAuthorizationRemoveMember"
                      @click.stop="showNamespaceMemberDelete(members.indexOf(item))"
                    >
                      <NamespaceMemberDelete
                        :member="item"
                        :show.sync="namespaceMemberDeleteShow[members.indexOf(item)]"
                        data-test="namespaceMemberDelete-component"
                        @update="refresh"
                      />
                    </v-list-item>
                  </div>
                </template>

                <span>
                  You don't have this kind of authorization.
                </span>
              </v-tooltip>
            </v-list>
          </v-menu>
        </template>
      </v-data-table>
    </div>
  </fragment>
</template>

<script>

import NamespaceMemberFormDialogEdit from '@/components/namespace/NamespaceMemberFormDialogEdit';
import NamespaceMemberDelete from '@/components/namespace/NamespaceMemberDelete';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NamespaceMemberList',

  filters: { hasPermission },

  components: {
    NamespaceMemberFormDialogEdit,
    NamespaceMemberDelete,
  },

  props: {
    namespace: {
      type: Object,
      required: true,
    },
  },

  data() {
    return {
      menu: false,
      namespaceMemberFormShow: [],
      namespaceMemberDeleteShow: [],
      editMemberAction: 'editMember',
      removeMemberAction: 'removeMember',
      headers: [
        {
          text: 'Username',
          value: 'username',
          align: 'start',
          sortable: false,
        },
        {
          text: 'Role',
          value: 'role',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'end',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    members() {
      return this.namespace.members;
    },

    hasAuthorizationEditMember() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace[this.editMemberAction],
        );
      }

      return false;
    },

    hasAuthorizationRemoveMember() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace[this.removeMemberAction],
        );
      }

      return false;
    },
  },

  methods: {
    refresh() {
      this.getNamespace();
    },

    showNamespaceMemberFormEdit(index) {
      this.namespaceMemberFormShow[index] = this.namespaceMemberFormShow[index] === undefined
        ? true : !this.namespaceMemberFormShow[index];
      this.$set(this.namespaceMemberFormShow, index, this.namespaceMemberFormShow[index]);

      this.closeMenu(index);
    },

    showNamespaceMemberDelete(index) {
      this.namespaceMemberDeleteShow[index] = this.namespaceMemberDeleteShow[index] === undefined
        ? true : !this.namespaceMemberDeleteShow[index];
      this.$set(this.namespaceMemberDeleteShow, index, this.namespaceMemberDeleteShow[index]);

      this.closeMenu(index);
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
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
  },
};

</script>
