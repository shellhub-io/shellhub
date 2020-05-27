<template>
  <fragment>
    <v-edit-dialog
      :return-value="editName"
      large
      @open="editName = device.name"
      @save="save()"
    >
      <v-text-field
        slot="input"
        label="Edit"
        single-line
      />
      <v-toolbar-title>
        <v-icon
          v-if="device.online"
          color="success"
        >
          check_circle
        </v-icon>
        <v-tooltip
          v-else
          bottom
        >
          <template #activator="{ on }">
            <v-icon v-on="on">
              check_circle
            </v-icon>
          </template>
          <span>active {{ device.last_seen | moment("from", "now") }}</span>
        </v-tooltip>
        {{ device.name }}
        <v-icon
          small
          left
        >
          mdi-file-edit
        </v-icon>
      </v-toolbar-title>
    </v-edit-dialog>
  </fragment>
</template>

<script>
export default {
  name: 'DeviceRename',

  data() {
    return {
      editName: '',
    };
  },
  methods : {
    save() {
      this.$store.dispatch('devices/rename', {
        uid: this.device.uid,
        name: this.editName
      });

      this.device.name = this.editName;
    }
  }
};
</script>
