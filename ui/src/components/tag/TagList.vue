<template>
  <fragment>
    <v-data-table
      :headers="headers"
      :items="getListTags"
      :server-items-length="getNumberTags"
      hide-default-footer
      data-test="tagListList-dataTable"
    >
      <template #[`item.name`]="{ item }">
        {{ item.name }}
      </template>

      <template #[`item.actions`]="{ item }">
        <v-menu
          :ref="'menu'+getListTags.indexOf(item)"
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
            <v-list-item @click="showTagDialog(getListTags.indexOf(item))">
              <TagFormDialog
                action="edit"
                :tag-name="item.name"
                :show.sync="tagDialogShow[getListTags.indexOf(item)]"
                data-test="tagFormDialog-component"
                @update="getTags()"
              />
            </v-list-item>

            <v-tooltip
              bottom
              :disabled="hasAuthorizationRemove"
            >
              <template #activator="{ on, attrs }">
                <div
                  v-bind="attrs"
                  v-on="on"
                >
                  <v-list-item
                    :disabled="!hasAuthorizationRemove"
                    @click="showTagDelete(getListTags.indexOf(item))"
                  >
                    <TagDelete
                      :tag-name="item.name"
                      :show.sync="tagDeleteShow[getListTags.indexOf(item)]"
                      data-test="tagDelete-component"
                      @update="getTags()"
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
  </fragment>
</template>

<script>

import TagFormDialog from '@/components/tag/TagFormDialog';
import TagDelete from '@/components/tag/TagDelete';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'TagListComponent',

  filters: { hasPermission },

  components: {
    TagFormDialog,
    TagDelete,
  },

  data() {
    return {
      tagDialogShow: [],
      tagDeleteShow: [],
      removeAction: 'remove',
      headers: [
        {
          text: 'Name',
          value: 'name',
          align: 'center',
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'center',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    getListTags() {
      return this.$store.getters['tags/list'].map((str) => ({ name: str }));
    },

    getNumberTags() {
      return this.$store.getters['tags/getNumberTags'];
    },

    hasAuthorizationRemove() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.tag[this.removeAction],
        );
      }

      return false;
    },
  },

  created() {
    this.getTags();
  },

  methods: {
    async getTags() {
      try {
        await this.$store.dispatch('tags/fetch');

        this.setArrays();
      } catch (error) {
        this.$store.dispatch('snackbar/showSnackbarErrorLoading', this.$errors.snackbar.deviceTagList);
      }
    },

    showTagDialog(index) {
      this.tagDialogShow[index] = this.tagDialogShow[index] === undefined
        ? true : !this.tagDialogShow[index];
      this.$set(this.tagDialogShow, index, this.tagDialogShow[index]);

      this.closeMenu(index);
    },

    showTagDelete(index) {
      this.tagDeleteShow[index] = this.tagDeleteShow[index] === undefined
        ? true : !this.tagDeleteShow[index];
      this.$set(this.tagDeleteShow, index, this.tagDeleteShow[index]);

      this.closeMenu(index);
    },

    setArrays() {
      const numberTags = this.getListTags.length;

      if (numberTags > 0) {
        this.tagDialogShow = new Array(numberTags).fill(false);
        this.tagDeleteShow = new Array(numberTags).fill(false);
      }
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
    },
  },
};
</script>
