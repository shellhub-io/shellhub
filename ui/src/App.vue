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

    <v-app-bar
      v-if="isLoggedIn"
      app
      clipped-left
      dark
      color="primary"
    >
      <v-app-bar-nav-icon
        class="hidden-lg-and-up"
        @click.stop="drawer = !drawer"
      />
      <router-link to="/">
        <v-img
          src="@/assets/logo-inverted.png"
          max-width="160"
        />
      </router-link>
      <span class="overline mt-3">BETA</span>
      <v-spacer />

      <NamespaceMenu
        :in-a-namespace="hasNamespaces"
      />

      <v-chip>
        <v-icon>help</v-icon>
      </v-chip>

      <Notification />

      <v-menu
        offset-y
      >
        <template #activator="{ on }">
          <v-chip v-on="on">
            <v-icon left>
              mdi-account
            </v-icon>
            {{ $store.getters["auth/currentUser"] }}
            <v-icon right>
              mdi-chevron-down
            </v-icon>
          </v-chip>
        </template>

        <v-card>
          <v-list-item
            v-for="(item, index) in menu"
            :key="index"
            router
            :data-test="item.title"
            @click.prevent="triggerClick(item)"
          >
            <v-icon left>
              {{ item.icon }}
            </v-icon>
            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </v-list-item>
        </v-card>
      </v-menu>
    </v-app-bar>
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

import Notification from '@/components/app_bar/notification/Notification';
import NamespaceMenu from '@/components/app_bar/namespace/NamespaceMenu';

export default {
  name: 'App',

  components: {
    Notification,
    NamespaceMenu,
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      showMenu: false,
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
          hidden: !this.$env.isHosted,
        },
      ],
      menu: [
        {
          title: 'Settings',
          type: 'path',
          path: '/settings',
          icon: 'mdi-cog',
          items: [{ title: 'Profile', path: '/settings/profile' }],
        },
        {
          title: 'Logout',
          type: 'method',
          icon: 'mdi-logout',
          method: 'logout',
        },
      ],
      admins: [
        ['Management', 'people_outline'],
        ['Settings', 'settings'],
      ],
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },

    visibleItems() {
      return this.items.filter((item) => !item.hidden);
    },

    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    currentInANamespace() {
      return localStorage.getItem('tenant') !== '';
    },
  },

  methods: {
    logout() {
      this.$store.dispatch('auth/logout').then(() => {
        this.$router.push('/login');
      });
    },
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
