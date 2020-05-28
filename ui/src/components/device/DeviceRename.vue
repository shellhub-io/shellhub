<template>
  <fragment>
    <v-tooltip bottom>
      <template v-slot:activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          mdi-file-edit
        </v-icon>
      </template>
      <span>Edit</span>
    </v-tooltip>
    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title>
          Rename Device
        </v-card-title>
        <v-card-text>
          <v-text-field
            v-model="editName"
            required
            hint="Exs ::(localhost, foo-bar.ba-z.qux, foo.example.com, 127.0.0.1)"
            input="editName"
          />
        </v-card-text>
        <v-card-text
          v-if="invalid"
          color="red"
        >
          You entered an invalid hostname,
          it must follow the
          <a
            target="_blank"
            href="https://tools.ietf.org/html/rfc1123"
          >
            RFC1123
          </a>
          specifications
        </v-card-text>
        <v-card-actions>
          <v-btn
            color="primary"
            text
            @click="dialog=false;invalid=false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            text
            @click="check"
          >
            Rename
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>
import isValidHostname from 'is-valid-hostname';

export default {
  name: 'DeviceRename',

  props: {
    hostname: {
      type: String,
      required: true
    },
    uid: {
      type:String,
      required:true
    }
  },

  data() {
    return {
      dialog: false,
      editName: '',
      invalid: false,
    };
  },
  methods : {
    save() {
      this.$store.dispatch('devices/rename', {
        uid: this.uid,
        name: this.editName
      });
      this.dialog=false;
    },
    check(){
      if (isValidHostname(this.editName)){
        this.save();
        this.dialog=false;
        this.$emit('newHostname', this.editName);
      }else{
        this.invalid=true;
      }
    },
  }
};
</script>
