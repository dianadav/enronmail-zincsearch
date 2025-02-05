import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import vuetify from './plugins/vuetify'
import App from './App.vue'
import router from './router'
import '@mdi/font/css/materialdesignicons.css'
import VueTablerIcons from 'vue-tabler-icons'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(VueTablerIcons)
app.use(vuetify).mount('#app')
