<template>
  <v-app>
    <v-navigation-drawer
      v-if="isLoggedIn && hasNamespaces"
      v-model="drawer"
      app
      clipped
      dark
    >
      <v-list>
        <v-list-item
          v-for="item in visibleItems"
          :key="item.title"
          :to="item.path"
          two-line
        >
          <v-list-item-action>
            <v-icon v-text="item.icon" />
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title
              :data-test="item.icon"
              v-text="item.title"
            />
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <AppBar />

    <v-main class="grey lighten-4">
      <v-container
        class="pa-8"
        fluid
      >
        <router-view :key="$route.fullPath" />
      </v-container>

      <snackbar />
    </v-main>
  </v-app>
</template>

<script>

import AppBar from '@/components/app_bar/AppBar';

export default {
  name: 'App',

  components: {
    AppBar,
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      items: [
        {
          icon: 'dashboard',
          title: 'Dashboard',
          path: '/',
        },
        {
          icon: 'devices',
          title: 'Devices',
          path: '/devices',
        },
        {
          icon: 'history',
          title: 'Sessions',
          path: '/sessions',
        },
        {
          icon: 'security',
          title: 'Firewall Rules',
          path: '/firewall/rules',
          hidden: !this.$env.isEnterprise,
        },
        {
          icon: 'vpn_key',
          title: 'Public Keys',
          path: '/sshkeys/public-keys',
        },
      ],
      admins: [
        ['Management', 'people_outline'],
        ['Settings', 'settings'],
      ],
    };
  },

  computed: {
    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },

    visibleItems() {
      return this.items.filter((item) => !item.hidden);
    },

    hasLoggedID() {
      return this.$store.getters['auth/id'] !== '';
    },

    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },
  },

  created() { // previous user tenant changed into a namespace
    if (!this.hasLoggedID && this.isLoggedIn) {
      try {
        this.$store.dispatch('auth/logout').then(() => {
          this.$router.push('/login');
        });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceReload);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$error.namespaceLoad);
      }
    }
  },

  mounted() {
    this.$store.dispatch('privatekeys/fetch');
  },

  methods: {
    triggerClick(item) {
      switch (item.type) {
      case 'path':
        this.$router.push(item.path).catch(() => {});
        break;
      case 'method':
        this[item.method]();
        break;
      default:
        break;
      }
    },
  },
};

</script>

<style>
.v-list-active {
  border-left: 4px solid var(--v-primary-base);
}

.text-shadow {
  text-shadow: #000 0 0 6px;
  color: transparent;
}
</style>
